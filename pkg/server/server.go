package server

import (
	"github.com/dronbrigadir/tech-mail-db-forum/pkg/service"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	port   string
	router *mux.Router
}

func NewServer(port string) *Server {
	return &Server{
		port:   ":" + port,
		router: mux.NewRouter(),
	}
}

func (s *Server) InitRoutes() {
	s.router.PathPrefix("/").Handler(service.SwaggerUI())
}

func (s *Server) Run() error {
	return http.ListenAndServe(s.port, s.router)
}
