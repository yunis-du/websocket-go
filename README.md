# websocket-go

## Installation

```bash
go get github.com/yunis-du/websocket-go@v1.0.3
```

## Usage

### create websocket server

```go
package main

import (
	"fmt"
	"os"

	"github.com/yunis-du/websocket-go/http"
	"github.com/yunis-du/websocket-go/websocket"
)

type serverEventHandler struct {
}

func (s *serverEventHandler) OnMessage(c websocket.Speaker, m websocket.Message) {
	bytes, _ := m.Bytes(websocket.JSONProtocol)
	fmt.Printf("OnMessage: %s\n", string(bytes))
}

func (s *serverEventHandler) OnConnected(c websocket.Speaker) {

}

func (s *serverEventHandler) OnDisconnected(c websocket.Speaker) {

}

func server() {
    hostname, _ := os.Hostname()
    httpServer := http.NewServer(hostname, "0.0.0.0", 8080)
    httpServer.ListenAndServe()
    
    wsServer := websocket.NewServer(httpServer, "/ws/server")
    
    wsServer.AddEventHandler(&serverEventHandler{})
    
    wsServer.Start()
    
    select {}
}

func main() {
    server()
}

```

### create websocket struct server

```go
package main

import (
	"fmt"
	"os"

	"github.com/yunis-du/websocket-go/http"
	"github.com/yunis-du/websocket-go/websocket"
)
type structEventHandler struct {
}

func (s *structEventHandler) OnMessage(c websocket.Speaker, m websocket.Message) {}

func (s *structEventHandler) OnStructMessage(c websocket.Speaker, m *websocket.StructMessage) {
	fmt.Printf("OnStructMessage: %s\n", string(m.Obj))
}

func (s *structEventHandler) OnConnected(c websocket.Speaker) {
	// do something
}

func (s *structEventHandler) OnDisconnected(c websocket.Speaker) {
	// do something
}

func structServer() {
	hostname, _ := os.Hostname()
	httpServer := http.NewServer(hostname, "0.0.0.0", 8080)
	httpServer.ListenAndServe()

	wsServer := websocket.NewServer(httpServer, "/ws/server")
	sServer := websocket.NewStructServer(wsServer)

	sServer.AddStructMessageHandler(&structEventHandler{}, []string{"test"})

	sServer.Start()
	select {}
}

func main() {
	structServer()
}
```

### create websocket client

```go
package main

import (
    "fmt"
    "net/url"
    "os"

	"github.com/yunis-du/websocket-go/http"
    "github.com/yunis-du/websocket-go/websocket"
)

type clientEventHandler struct {
}

func (cl *clientEventHandler) OnMessage(c websocket.Speaker, m websocket.Message) {
	bytes, _ := m.Bytes(websocket.JSONProtocol)
	fmt.Printf("OnMessage: %s\n", string(bytes))
}

func (cl *clientEventHandler) OnConnected(c websocket.Speaker) {

}

func (cl *clientEventHandler) OnDisconnected(c websocket.Speaker) {

}

func client() {
	hostname, _ := os.Hostname()

	u := &url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/ws/server"}
	opt := websocket.ClientOpts{
		Protocol:  websocket.RawProtocol,
		QueueSize: 1000,
	}

	wsClient := websocket.NewClient(hostname, u, opt)
	wsClient.AddEventHandler(&clientEventHandler{})
	
	wsClient.Start()

	err := wsClient.SendRaw([]byte("msg"))
	if err != nil {
		fmt.Println(err)
	}

	err = wsClient.SendMessage(websocket.NewStructMessage("namespace", "type", "data"))
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	client()
}
```

### create websocket client pool

```go
package main

import (
    "fmt"
    "net/url"
    "os"

	"github.com/yunis-du/websocket-go/http"
    "github.com/yunis-du/websocket-go/websocket"
)

type clientStructEventHandler struct {
}

func (s *clientStructEventHandler) OnMessage(c websocket.Speaker, m websocket.Message) {}

func (s *clientStructEventHandler) OnStructMessage(c websocket.Speaker, m *websocket.StructMessage) {
	fmt.Printf("OnStructMessage: %s\n", string(m.Obj))
}

func (s *clientStructEventHandler) OnConnected(c websocket.Speaker) {
	// do something
}

func (s *clientStructEventHandler) OnDisconnected(c websocket.Speaker) {
	// do something
}


func clientPool()  {
	hostname, _ := os.Hostname()

	u := &url.URL{Scheme: "ws", Host: "127.0.0.1:8080", Path: "/ws/server"}
	opt := websocket.ClientOpts{
		Protocol:  websocket.RawProtocol,
		QueueSize: 1000,
	}

	pool := websocket.NewStructClientPool("client-pool")

	wsClient := websocket.NewClient(hostname, u, opt)

	_ = pool.AddClient(wsClient)

	pool.AddStructMessageHandler(&clientStructEventHandler{}, []string{"test"})

	pool.ConnectAll()

	err := pool.GetSpeakerByRemoteHost(hostname).SendRaw([]byte("msg"))
	if err != nil {
		fmt.Println(err)
	}

	err = pool.GetSpeakerByRemoteHost(hostname).SendMessage(websocket.NewStructMessage("namespace", "type", "data"))
	if err != nil {
		fmt.Println(err)
	}

}

func main() {
	clientPool()
}

```
