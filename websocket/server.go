package websocket

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yunis-du/websocket-go/common"
	htp "github.com/yunis-du/websocket-go/http"
)

const maxMessageSize = 0

type clientPromoter func(c *wsIncomingClient) (Speaker, error)

type IncomerHandler func(*websocket.Conn, *http.Request, clientPromoter) (Speaker, error)

type Server struct {
	common.RWMutex
	*incomerPool
	server         *htp.Server
	incomerHandler IncomerHandler
}

func (s *Server) serverMessage(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:    1024 * 10,
		WriteBufferSize:   1024 * 10,
		EnableCompression: true,
	}
	upgrader.Error = func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		// don't return errors to maintain backwards compatibility
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {
		// allow all connections by default
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Default().Printf("upgrade: %s", err)
		return
	}
	_, err = s.incomerHandler(conn, r, func(c *wsIncomingClient) (Speaker, error) { return c, nil })
	if err != nil {
		log.Default().Printf("Unable to accept incomer from %s: %s", r.RemoteAddr, err)
		w.Header().Set("Connection", "close")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
}

func (s *Server) newIncomingClient(conn *websocket.Conn, r *http.Request, promoter clientPromoter) (Speaker, error) {
	log.Default().Printf("New WebSocket Connection from %s : URI path %s", conn.RemoteAddr().String(), r.URL.Path)

	var clientProtocol Protocol
	if err := clientProtocol.parse(getRequestParameter(r, "X-Client-Protocol")); err != nil {
		return nil, fmt.Errorf("protocol requested error: %s", err)
	}

	ur, _ := url.Parse(fmt.Sprintf("http://%s%s", conn.RemoteAddr().String(), r.URL.Path+"?"+r.URL.RawQuery))

	wsConn := newConn(s.server.Host, clientProtocol, ur, r.Header)
	wsConn.conn = conn
	wsConn.RemoteHost = getRequestParameter(r, "X-Host-ID")

	// NOTE(safchain): fallback to remote addr if host id not provided
	// should be removed, connection should be refused if host id not provided
	if wsConn.RemoteHost == "" {
		wsConn.RemoteHost = r.RemoteAddr
	}
	conn.SetReadLimit(maxMessageSize)
	c := &wsIncomingClient{
		Conn: wsConn,
	}
	wsConn.wsSpeaker = c

	pc, err := promoter(c)
	if err != nil {
		return nil, err
	}

	c.State.Store(common.RunningState)

	// add the new Speaker to the server pool
	s.AddClient(pc)

	// notify the pool listeners that the speaker is connected
	s.OnConnected(pc)

	// send a first ping to help firefox and some other client which wait for a
	// first ping before doing something
	c.sendPing()

	wsConn.pingTicker = time.NewTicker(time.Second * 30)

	c.Start()

	return c, nil
}

func getRequestParameter(r *http.Request, name string) string {
	param := r.Header.Get(name)
	if param == "" {
		param = r.URL.Query().Get(strings.ToLower(name))
	}
	return param
}

type StructServer struct {
	*Server
	*structSpeakerPoolEventDispatcher
}

func (s *StructServer) Request(host string, request *StructMessage, timeout time.Duration) (*StructMessage, error) {
	c := s.Server.GetSpeakerByRemoteHost(host)
	if c == nil {
		return nil, errors.New("no result found")
	}

	return c.(*StructSpeaker).Request(request, timeout)
}

// OnMessage websocket event.
func (s *StructServer) OnMessage(c Speaker, m Message) {
}

// OnConnected websocket event.
func (s *StructServer) OnConnected(c Speaker) {
	for _, handlers := range s.nsEventHandlers {
		for _, handler := range handlers {
			handler.OnConnected(c)
		}
	}
}

// OnDisconnected removes the Speaker from the incomer pool.
func (s *StructServer) OnDisconnected(c Speaker) {
	for _, handlers := range s.nsEventHandlers {
		for _, handler := range handlers {
			handler.OnDisconnected(c)
		}
	}
}

func NewServer(server *htp.Server, endpoint string) *Server {
	s := &Server{
		incomerPool: newIncomerPool(endpoint), // server inherits from a Speaker pool
		server:      server,
	}
	s.incomerHandler = func(conn *websocket.Conn, r *http.Request, promoter clientPromoter) (Speaker, error) {
		return s.newIncomingClient(conn, r, promoter)
	}
	server.HandleFunc(endpoint, s.serverMessage)
	return s
}

func NewStructServer(server *Server) *StructServer {
	s := &StructServer{
		Server:                           server,
		structSpeakerPoolEventDispatcher: newStructSpeakerPoolEventDispatcher(server),
	}

	s.Server.incomerPool.AddEventHandler(s)

	// This incomerHandler upgrades the incomers to StructSpeaker thus being able to parse StructMessage.
	// The server set also the StructSpeaker with the proper namspaces it subscribes to thanks to the
	// headers.
	s.Server.incomerHandler = func(conn *websocket.Conn, r *http.Request, promoter clientPromoter) (Speaker, error) {
		// the default incomer handler creates a standard wsIncomingClient that we upgrade to a StructSpeaker
		// being able to handle the StructMessage
		uc, err := s.Server.newIncomingClient(conn, r, func(ic *wsIncomingClient) (Speaker, error) {
			c := ic.upgradeToStructSpeaker()

			s.structSpeakerPoolEventDispatcher.AddStructSpeaker(c)

			return c, nil
		})
		if err != nil {
			return nil, err
		}

		return uc, nil
	}

	return s
}
