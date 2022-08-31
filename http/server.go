package http

import (
	"context"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	sync.RWMutex
	http.Server
	Host     string
	Router   *mux.Router
	Addr     string
	Port     int
	lock     sync.Mutex
	listener net.Listener
	wg       sync.WaitGroup
}

func (s *Server) Listen() error {
	listenAddrPort := fmt.Sprintf("%s:%d", s.Addr, s.Port)
	ln, err := net.Listen("tcp", listenAddrPort)
	if err != nil {
		return fmt.Errorf("Failed to listen on %s:%d: %s", s.Addr, s.Port, err)
	}

	s.listener = ln
	log.Default().Printf("Listening on socket %s:%d", s.Addr, s.Port)
	return nil
}

func (s *Server) ListenAndServe() {
	if err := s.Listen(); err != nil {
		log.Default().Println(err)
	}
	go s.Serve()
}

func (s *Server) Serve() {
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
	log.Default().Printf("Failed to serve on %s:%d: %s", s.Addr, s.Port, err)
}

func (s *Server) HandleFunc(path string, f func(w http.ResponseWriter, r *http.Request)) {
	s.Router.HandleFunc(path, f)
}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Server.Shutdown(ctx); err != nil {
		log.Default().Printf("Shutdown error: %s", err)
	}
	s.listener.Close()
	s.wg.Wait()
}

func NewServer(host string, addr string, port int) *Server {
	router := mux.NewRouter().StrictSlash(true)
	router.Headers("X-Host-ID", host)

	return &Server{
		Server: http.Server{},
		Host:   host,
		Router: router,
		Addr:   addr,
		Port:   port,
	}
}
