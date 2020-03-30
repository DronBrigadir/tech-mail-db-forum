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

	s.router.HandleFunc("/api/forum/create", controllers.ForumCreate).Methods(http.MethodPost)
	s.router.HandleFunc("/api/forum/{slug}/details", controllers.ForumGet).Methods(http.MethodGet)
	s.router.HandleFunc("/api/forum/{slug}/create", controllers.CreateForumThread).Methods(http.MethodPost)
	s.router.HandleFunc("/api/forum/{slug}/threads", controllers.GetForumThreads).Methods(http.MethodGet)
	s.router.HandleFunc("/api/forum/{slug}/users", controllers.GetForumUsers).Methods(http.MethodGet)

	s.router.HandleFunc("/api/thread/{slug_or_id}/create", controllers.CreateThreadPost).Methods(http.MethodPost)
	s.router.HandleFunc("/api/thread/{slug_or_id}/details", controllers.GetThread).Methods(http.MethodGet)
	s.router.HandleFunc("/api/thread/{slug_or_id}/details", controllers.UpdateThread).Methods(http.MethodPost)
	s.router.HandleFunc("/api/thread/{slug_or_id}/posts", controllers.GetThreadPosts).Methods(http.MethodGet)
	s.router.HandleFunc("/api/thread/{slug_or_id}/vote", controllers.VoteThread).Methods(http.MethodPost)

	s.router.HandleFunc("/api/user/{nickname}/create", controllers.UserCreate).Methods(http.MethodPost)
	s.router.HandleFunc("/api/user/{nickname}/profile", controllers.GetUser).Methods(http.MethodGet)
	s.router.HandleFunc("/api/user/{nickname}/profile", controllers.UpdateUser).Methods(http.MethodPost)

	s.router.HandleFunc("/api/service/clear", controllers.ClearDB).Methods(http.MethodPost)
	s.router.HandleFunc("/api/service/status", controllers.StatusDB).Methods(http.MethodGet)

	s.router.HandleFunc("/api/post/{id}/details", controllers.GetPostDetails).Methods(http.MethodGet)
	s.router.HandleFunc("/api/post/{id}/details", controllers.UpdatePost).Methods(http.MethodPost)
}

func (s *Server) Run() error {
	return http.ListenAndServe(s.port, s.router)
}
