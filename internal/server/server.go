package server

import (
	"github.com/dronbrigadir/tech-mail-db-forum/internal/controllers"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

type Server struct {
	port   string
	router *router.Router
}

func NewServer(port string) *Server {
	return &Server{
		port:   ":" + port,
		router: router.New(),
	}
}

func (s *Server) InitRoutes() {
	s.router.POST("/api/forum/create", controllers.ForumCreate)
	s.router.POST("/api/forum/{slug}/create", controllers.CreateForumThread)
	s.router.GET("/api/forum/{slug}/details", controllers.ForumGet)
	s.router.GET("/api/forum/{slug}/threads", controllers.GetForumThreads)
	s.router.GET("/api/forum/{slug}/users", controllers.GetForumUsers)

	s.router.POST("/api/thread/{slug_or_id}/create", controllers.CreateThreadPost)
	s.router.POST("/api/thread/{slug_or_id}/vote", controllers.VoteThread)
	s.router.POST("/api/thread/{slug_or_id}/details", controllers.UpdateThread)
	s.router.GET("/api/thread/{slug_or_id}/details", controllers.GetThread)
	s.router.GET("/api/thread/{slug_or_id}/posts", controllers.GetThreadPosts)

	s.router.POST("/api/user/{nickname}/create", controllers.UserCreate)
	s.router.POST("/api/user/{nickname}/profile", controllers.UpdateUser)
	s.router.GET("/api/user/{nickname}/profile", controllers.GetUser)

	s.router.POST("/api/service/clear", controllers.ClearDB)
	s.router.GET("/api/service/status", controllers.StatusDB)

	s.router.POST("/api/post/{id}/details", controllers.UpdatePost)
	s.router.GET("/api/post/{id}/details", controllers.GetPostDetails)

	//s.router.PathPrefix("/swaggerui/").Handler(controllers.SwaggerUI())
	//s.router.HandleFunc("/api/swagger.yml", controllers.SwaggerApi).Methods(http.MethodGet)
}

func (s *Server) Run() error {
	return fasthttp.ListenAndServe(s.port, s.router.Handler)
}
