package websocket

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	*Conn
	Path string
	Opts ClientOpts
}

type wsIncomingClient struct {
	*Conn
}

type ClientOpts struct {
	Protocol         Protocol
	Headers          http.Header
	QueueSize        int
	WriteCompression bool
	TLSConfig        *tls.Config
}

func (c *wsIncomingClient) upgradeToStructSpeaker() *StructSpeaker {
	s := newStructSpeaker(c)
	c.Lock()
	c.wsSpeaker = s
	c.Unlock()
	return s
}

// func (c *Client) scheme() string {
// 	return "ws://"
// }

func (c *Client) Connect() error {
	var err error
	endpoint := c.URL.String()
	headers := http.Header{}

	for k, v := range c.Headers {
		headers[k] = v
	}

	c.Logger.Infof("Connecting to %s", endpoint)

	d := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		ReadBufferSize:    1024 * 10,
		WriteBufferSize:   1024 * 10,
		EnableCompression: true,
	}

	var resp *http.Response
	c.conn, resp, err = d.Dial(endpoint, headers)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("unable to create a WebSocket connection %s(%v) : %s", endpoint, resp, err)
		} else {
			return fmt.Errorf("unable to create a WebSocket connection %s : %s", endpoint, err)
		}
	}

	c.conn.SetPingHandler(nil)
	c.conn.EnableWriteCompression(c.writeCompression)

	c.State.Store(RunningState)

	c.Logger.Infof("Connected to %s", endpoint)

	c.RemoteHost = resp.Header.Get("X-Host-ID")

	// NOTE(safchain): fallback to remote addr if host id not provided
	// should be removed, connection should be refused if host id not provided
	if c.RemoteHost == "" {
		c.RemoteHost = c.conn.RemoteAddr().String()
	}

	// notify connected
	for _, l := range c.cloneEventHandlers() {
		l.OnConnected(c)
	}

	return nil
}

// Start connects to the server - and reconnect if necessary
func (c *Client) Start() {
	go func() {
		for c.running.Load() == true {
			if err := c.Connect(); err == nil {
				c.Run()
				if c.running.Load() == true {
					c.wg.Wait()
				}
			} else {
				c.Logger.Error(err)
			}
			time.Sleep(5 * time.Second)
		}
	}()
}

func (c *Client) UpgradeToStructSpeaker() *StructSpeaker {
	s := newStructSpeaker(c)
	c.Lock()
	c.wsSpeaker = s
	c.Unlock()
	return s
}

func NewClient(host string, url *url.URL, opts ClientOpts) *Client {
	wsConn := newConn(host, opts.Protocol, url, opts.Headers)
	c := &Client{
		Conn: wsConn,
		Opts: opts,
	}

	wsConn.wsSpeaker = c
	return c
}
