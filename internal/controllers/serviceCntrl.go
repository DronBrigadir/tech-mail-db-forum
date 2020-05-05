package controllers

import (
	"github.com/dronbrigadir/tech-mail-db-forum/internal/database"
	"github.com/dronbrigadir/tech-mail-db-forum/internal/models"
	"github.com/dronbrigadir/tech-mail-db-forum/tools"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
)

func ClearDB(ctx *fasthttp.RequestCtx) {
	db := database.Connection

	_, _ = db.Exec("TRUNCATE TABLE Forum, ForumUser, Post, Thread, Users, Vote CASCADE")

	return
}

func StatusDB(ctx *fasthttp.RequestCtx) {
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

	tools.ObjectResponce(ctx, http.StatusOK, status)
	return
}
