package main

import (
	"github.com/dronbrigadir/tech-mail-db-forum/internal/server"
	"log"
)

func main() {
	s := server.NewServer("5000")
	s.InitRoutes()

	log.Println("starting server at :5000")
	if err := s.Run(); err != nil {
		log.Fatal(err, "can't run server")
	}
}
