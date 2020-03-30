package controllers

import (
	"github.com/dronbrigadir/tech-mail-db-forum/internal/database"
	"github.com/dronbrigadir/tech-mail-db-forum/internal/models"
	"github.com/dronbrigadir/tech-mail-db-forum/tools"
	"log"
	"net/http"
)

func ClearDB(w http.ResponseWriter, r *http.Request) {
	db := database.Connection

	_, _ = db.Exec("TRUNCATE TABLE Forum, ForumUser, Post, Thread, Users, Vote CASCADE")

	return
}

func StatusDB(w http.ResponseWriter, r *http.Request) {
	db := database.Connection

	status := models.Status{}
	err := db.QueryRow(
		"SELECT "+
			"(SELECT COUNT(*) FROM users), "+
			"(SELECT COUNT(*) FROM forum), "+
			"(SELECT COUNT(*) FROM thread), "+
			"(SELECT COUNT(*) FROM post)",
	).Scan(&status.UserQuantity, &status.ForumQuantity, &status.ThreadQuantity, &status.PostQuantity)

	if err != nil {
		log.Println(err)
		return
	}

	tools.ObjectResponce(w, http.StatusOK, status)
	return
}
