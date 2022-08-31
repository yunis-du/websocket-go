package websocket

import (
	"errors"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	"websocket-go/common"
)

const writeWait = 60 * time.Second

type ConnState int64

type Speaker interface {
	GetStatus() ConnStatus
	GetHost() string
	GetAddrPort() (string, int)
	GetClientProtocol() Protocol
	GetHeaders() http.Header
	GetURL() *url.URL
	IsConnected() bool
	SendMessage(m Message) error
	SendRaw(r []byte) error
	Connect() error
	Start()
	Stop()
	StopAndWait()
	AddEventHandler(SpeakerEventHandler)
	GetRemoteHost() string
}

type SpeakerEventHandler interface {
	OnMessage(c Speaker, m Message)
	OnConnected(c Speaker)
	OnDisconnected(c Speaker)
}

type DefaultSpeakerEventHandler struct {
}

// OnMessage is called when a message is received.
func (d *DefaultSpeakerEventHandler) OnMessage(c Speaker, m Message) {
}

// OnConnected is called when the connection is established.
func (d *DefaultSpeakerEventHandler) OnConnected(c Speaker) {
}

// OnDisconnected is called when the connection is closed or lost.
func (d *DefaultSpeakerEventHandler) OnDisconnected(c Speaker) {
}

type StructSpeaker struct {
	Speaker
	eventHandlersLock common.RWMutex
	nsEventHandlers   map[string][]SpeakerStructMessageHandler
	replyChanMutex    common.RWMutex
	replyChan         map[string]chan *StructMessage
}

func (s *StructSpeaker) AddStructMessageHandler(h SpeakerStructMessageHandler, namespaces []string) {
	s.eventHandlersLock.Lock()
	// add this handler per namespace
	for _, ns := range namespaces {
		if _, ok := s.nsEventHandlers[ns]; !ok {
			s.nsEventHandlers[ns] = []SpeakerStructMessageHandler{h}
		} else {
			s.nsEventHandlers[ns] = append(s.nsEventHandlers[ns], h)
		}
	}
	s.eventHandlersLock.Unlock()
}

func (s *StructSpeaker) OnMessage(c Speaker, m Message) {
	if c, ok := c.(*StructSpeaker); ok {
		bytes, _ := m.Bytes(RawProtocol)

		var structMsg StructMessage
		if err := structMsg.unmarshalByProtocol(bytes, c.GetClientProtocol()); err != nil {
			return
		}
		s.dispatchMessage(c, &structMsg)
	}
}

func (s *StructSpeaker) OnConnected(c Speaker) {
	for _, handlers := range s.nsEventHandlers {
		for _, handler := range handlers {
			handler.OnConnected(c)
		}
	}
}

func (s *StructSpeaker) OnDisconnected(c Speaker) {
	for _, handlers := range s.nsEventHandlers {
		for _, handler := range handlers {
			handler.OnDisconnected(c)
		}
	}
}

func (s *StructSpeaker) dispatchMessage(c Speaker, m *StructMessage) {
	s.eventHandlersLock.RLock()
	var handlers []SpeakerStructMessageHandler
	handlers = append(handlers, s.nsEventHandlers[m.Namespace]...)
	handlers = append(handlers, s.nsEventHandlers[WildcardNamespace]...)
	s.eventHandlersLock.RUnlock()

	for _, h := range handlers {
		h.OnStructMessage(c, m)
	}
}

func (s *StructSpeaker) Request(m *StructMessage, timeout time.Duration) (*StructMessage, error) {
	ch := make(chan *StructMessage, 1)

	s.replyChanMutex.Lock()
	s.replyChan[m.UUID] = ch
	s.replyChanMutex.Unlock()

	defer func() {
		s.replyChanMutex.Lock()
		delete(s.replyChan, m.UUID)
		close(ch)
		s.replyChanMutex.Unlock()
	}()

	_ = s.SendMessage(m)

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(timeout):
		return nil, errors.New("timeout")
	}
}

type structSpeakerPoolEventDispatcher struct {
	eventHandlersLock common.RWMutex
	nsEventHandlers   map[string][]SpeakerStructMessageHandler
	pool              SpeakerPool
}

func newStructSpeakerPoolEventDispatcher(pool SpeakerPool) *structSpeakerPoolEventDispatcher {
	return &structSpeakerPoolEventDispatcher{
		nsEventHandlers: make(map[string][]SpeakerStructMessageHandler),
		pool:            pool,
	}
}

func (a *structSpeakerPoolEventDispatcher) AddStructMessageHandler(h SpeakerStructMessageHandler, namespaces []string) {
	a.eventHandlersLock.Lock()
	// add this handler per namespace
	for _, ns := range namespaces {
		if _, ok := a.nsEventHandlers[ns]; !ok {
			a.nsEventHandlers[ns] = []SpeakerStructMessageHandler{h}
		} else {
			a.nsEventHandlers[ns] = append(a.nsEventHandlers[ns], h)
		}
	}
	a.eventHandlersLock.Unlock()
}

func (a *structSpeakerPoolEventDispatcher) dispatchMessage(c *StructSpeaker, m *StructMessage) {
	a.eventHandlersLock.RLock()
	var handlers []SpeakerStructMessageHandler
	handlers = append(handlers, a.nsEventHandlers[m.Namespace]...)
	handlers = append(handlers, a.nsEventHandlers[WildcardNamespace]...)
	a.eventHandlersLock.RUnlock()

	for _, h := range handlers {
		h.OnStructMessage(c, m)
	}
}

func (a *structSpeakerPoolEventDispatcher) AddStructSpeaker(c *StructSpeaker) {
	a.eventHandlersLock.RLock()
	for ns, handlers := range a.nsEventHandlers {
		for _, handler := range handlers {
			c.AddStructMessageHandler(handler, []string{ns})
		}
	}
	a.eventHandlersLock.RUnlock()
}

type ConnStatus struct {
	ClientProtocol Protocol
	Addr           string
	Port           int
	Host           string      `json:"-"`
	State          *ConnState  `json:"IsConnected"`
	URL            *url.URL    `json:"-"`
	Headers        http.Header `json:"-"`
	ConnectTime    string
	RemoteHost     string `json:",omitempty"`
}

type Conn struct {
	common.RWMutex
	ConnStatus
	flush            chan struct{}
	send             chan []byte
	read             chan []byte
	quit             chan bool
	wg               sync.WaitGroup
	conn             *websocket.Conn
	running          atomic.Value
	pingTicker       *time.Ticker // only used by incoming connections
	eventHandlers    []SpeakerEventHandler
	wsSpeaker        Speaker // speaker owning the connection
	writeCompression bool
	messageType      int
}

func (s *ConnState) CompareAndSwap(old, new common.ServiceState) bool {
	return atomic.CompareAndSwapInt64((*int64)(s), int64(old), int64(new))
}

func (s *ConnState) Store(state common.ServiceState) {
	(*common.ServiceState)(s).Store(state)
}

func (s *ConnState) Load() common.ServiceState {
	return (*common.ServiceState)(s).Load()
}

func (c *Conn) GetHost() string {
	return c.Host
}

func (c *Conn) GetAddrPort() (string, int) {
	return c.Addr, c.Port
}

func (c *Conn) GetURL() *url.URL {
	return c.URL
}

func (c *Conn) IsConnected() bool {
	return c.State.Load() == common.RunningState
}

func (c *Conn) GetStatus() ConnStatus {
	c.RLock()
	defer c.RUnlock()

	status := c.ConnStatus
	status.State = new(ConnState)
	return c.ConnStatus
}

func (c *Conn) SendMessage(m Message) error {
	if !c.IsConnected() {
		return errors.New("not connected")
	}

	b, err := m.Bytes(c.GetClientProtocol())
	if err != nil {
		return err
	}

	c.send <- b
	return nil
}

func (c *Conn) SendRaw(b []byte) error {
	if !c.IsConnected() {
		return errors.New("Not connected")
	}

	c.send <- b

	return nil
}

func (c *Conn) GetClientProtocol() Protocol {
	return c.ClientProtocol
}

func (c *Conn) GetHeaders() http.Header {
	return c.Headers
}

func (c *Conn) GetRemoteHost() string {
	return c.RemoteHost
}

func (c *Conn) Run() {
	c.wg.Add(2)
	c.run()
}

func (c *Conn) Start() {
	c.wg.Add(2)
	go c.run()
}

func (c *Conn) AddEventHandler(h SpeakerEventHandler) {
	c.Lock()
	c.eventHandlers = append(c.eventHandlers, h)
	c.Unlock()
}

func (c *Conn) Connect() error {
	return nil
}

func (c *Conn) Flush() {
	c.flush <- struct{}{}
}

func (c *Conn) Stop() {
	c.running.Store(false)
	if c.State.CompareAndSwap(common.RunningState, common.StoppingState) {
		c.quit <- true
	}
}

func (c *Conn) StopAndWait() {
	c.Stop()
	c.wg.Wait()
}

func (c *Conn) cloneEventHandlers() (handlers []SpeakerEventHandler) {
	c.RLock()
	handlers = append(handlers, c.eventHandlers...)
	c.RUnlock()

	return
}

func (c *Conn) write(msg []byte) error {
	err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	c.conn.EnableWriteCompression(c.writeCompression)
	w, err := c.conn.NextWriter(c.messageType)
	if err != nil {
		return err
	}

	if _, err = w.Write(msg); err != nil {
		return err
	}

	return w.Close()
}

func (c *Conn) sendPing() error {
	err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err != nil {
		return err
	}
	return c.conn.WriteMessage(websocket.PingMessage, []byte{})
}

func (c *Conn) run() {
	flushChannel := func(c chan []byte, cb func(msg []byte) error) error {
		for {
			select {
			case m := <-c:
				if err := cb(m); err != nil {
					return err
				}
			default:
				return nil
			}
		}
	}

	// notify all the listeners that a message was received
	handleReceivedMessage := func(m []byte) error {
		for _, l := range c.cloneEventHandlers() {
			l.OnMessage(c.wsSpeaker, RawMessage(m))
		}
		return nil
	}

	// goroutine to read messages from the socket and put them into a channel
	go func() {
		defer c.wg.Done()

		for c.running.Load() == true {
			_, m, err := c.conn.ReadMessage()
			ip, port := c.GetAddrPort()
			if err != nil {
				log.Default().Printf("websocket error: %s, from %s:%d", err, ip, port)
				if c.running.Load() != false {
					c.quit <- true
				}
				break
			}
			c.read <- m
		}
		log.Default().Printf("websocket is stop, state: %v", c.running.Load())
	}()

	defer func() {
		c.conn.Close()
		c.State.Store(common.StoppedState)

		// handle all the pending received messages
		err := flushChannel(c.read, handleReceivedMessage)
		if err != nil {
			log.Default().Println(err)
		}

		for _, l := range c.cloneEventHandlers() {
			l.OnDisconnected(c.wsSpeaker)
		}

	LOOP:
		for {
			select {
			case <-c.send:
			case <-c.quit:
			case <-c.flush:
			default:
				break LOOP
			}
		}

		c.wg.Done()
	}()

	for {
		select {
		case <-c.quit:
			log.Default().Println("websocket quit")
			return
		case m := <-c.read:
			go func() {
				err := handleReceivedMessage(m)
				if err != nil {
					log.Default().Printf("Handle received message error: %s", err)
				}
			}()
		case m := <-c.send:
			if err := c.write(m); err != nil {
				log.Default().Printf("Error while sending message to %+v: %s", c, err)
				return
			}
		case <-c.flush:
			if err := flushChannel(c.send, c.write); err != nil {
				log.Default().Printf("Error while flushing send queue for %+v: %s", c, err)
				return
			}
		case <-c.pingTicker.C:
			if err := c.sendPing(); err != nil {
				log.Default().Printf("Error while sending ping to %+v: %s", c, err)

				// stop the ticker and request a quit
				c.pingTicker.Stop()
				return
			}
		}
	}
}

func newConn(host string, clientProtocol Protocol, url *url.URL, headers http.Header) *Conn {
	if headers == nil {
		headers = http.Header{}
	}

	port, _ := strconv.Atoi(url.Port())
	c := &Conn{
		ConnStatus: ConnStatus{
			Host:           host,
			ClientProtocol: clientProtocol,
			Addr:           url.Hostname(),
			Port:           port,
			State:          new(ConnState),
			URL:            url,
			Headers:        headers,
			ConnectTime:    time.Now().Format("2006-01-02 15:04:05"),
		},
		send:             make(chan []byte, 10000),
		read:             make(chan []byte, 10000),
		flush:            make(chan struct{}, 2),
		quit:             make(chan bool, 2),
		pingTicker:       &time.Ticker{},
		writeCompression: true,
	}

	if clientProtocol == JSONProtocol {
		c.messageType = websocket.TextMessage
	} else {
		c.messageType = websocket.BinaryMessage
	}

	c.State.Store(common.StoppedState)
	c.running.Store(true)
	return c
}

func newStructSpeaker(c Speaker) *StructSpeaker {
	s := &StructSpeaker{
		Speaker:         c,
		nsEventHandlers: make(map[string][]SpeakerStructMessageHandler),
	}

	// subscribing to itself so that the StructSpeaker can get Message and can convert them
	// to StructMessage and then forward them to its own even listeners.
	s.AddEventHandler(s)
	return s
}
