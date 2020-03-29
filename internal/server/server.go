package server

import (
	"github.com/dronbrigadir/tech-mail-db-forum/internal/controllers"
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
	s.router.PathPrefix("/swaggerui/").Handler(controllers.SwaggerUI())
	s.router.HandleFunc("/api/swagger.yml", controllers.SwaggerApi).Methods(http.MethodGet)
}

func (s *Server) Run() error {
	return http.ListenAndServe(s.port, s.router)
}
