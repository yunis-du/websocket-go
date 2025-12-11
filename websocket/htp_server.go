package websocket

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type HtpServer struct {
	http.Server
	Host     string
	Router   *mux.Router
	Addr     string
	Port     int
	listener net.Listener
	wg       sync.WaitGroup
	Logger   Logger
}

func (s *HtpServer) listen() error {
	if s.listener == nil {
		listenAddrPort := fmt.Sprintf("%s:%d", s.Addr, s.Port)
		ln, err := net.Listen("tcp", listenAddrPort)
		if err != nil {
			return fmt.Errorf("failed to listen on %s:%d: %s", s.Addr, s.Port, err)
		}

		s.listener = ln
		s.Logger.Infof("Listening on socket %s:%d", s.Addr, s.Port)
	}
	return nil
}

func (s *HtpServer) listenAndServe() {
	if err := s.listen(); err != nil {
		s.Logger.Error(err)
	}
	go s.serve()
}

func (s *HtpServer) serve() {
	defer s.wg.Done()
	s.wg.Add(1)

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "X-Auth-Token"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "DELETE", "OPTIONS"})

	s.Handler = handlers.CompressHandler(handlers.CORS(headersOk, originsOk, methodsOk)(s.Router))

	var err error
	if s.TLSConfig != nil {
		err = s.Server.ServeTLS(s.listener, "", "")
	} else {
		err = s.Server.Serve(s.listener)
	}

	if err == http.ErrServerClosed {
		return
	}
	s.Logger.Errorf("Failed to serve on %s:%d: %s", s.Addr, s.Port, err)
}

func (s *HtpServer) handleFunc(path string, f func(w http.ResponseWriter, r *http.Request)) {
	s.Router.HandleFunc(path, f)
}

// func (s *HtpServer) stop() {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	if err := s.Server.Shutdown(ctx); err != nil {
// 		log.Default().Printf("Shutdown error: %s", err)
// 	}
// 	s.listener.Close()
// 	s.wg.Wait()
// }

func newHtpServer(host string, addr string, port int) *HtpServer {
	router := mux.NewRouter().StrictSlash(true)
	router.Headers("X-Host-ID", host)

	return &HtpServer{
		Server: http.Server{},
		Host:   host,
		Router: router,
		Addr:   addr,
		Port:   port,
		Logger: NewStdLogger(),
	}
}

func newHtpServerWithListener(host string, listener net.Listener) *HtpServer {
	router := mux.NewRouter().StrictSlash(true)
	router.Headers("X-Host-ID", host)

	return &HtpServer{
		Server:   http.Server{},
		Host:     host,
		Router:   router,
		listener: listener,
		Logger:   NewStdLogger(),
	}
}
