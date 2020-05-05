package controllers

import (
	"fmt"
	"github.com/dronbrigadir/tech-mail-db-forum/internal/database"
	"github.com/dronbrigadir/tech-mail-db-forum/internal/models"
	"github.com/dronbrigadir/tech-mail-db-forum/tools"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func GetPostDetails(ctx *fasthttp.RequestCtx) {
	id, _ := strconv.Atoi(ctx.UserValue("id").(string))

	db := database.Connection

	post, err := tools.GetPostByID(db, id)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Post with id '%d' not found", id)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	related := strings.Split(string(ctx.QueryArgs().Peek("related")), ",")

	fullPost := models.PostFull{
		Post: &post,
	}

	for _, item := range related {
		switch item {
		case "user":
			user, _ := tools.GetUserByNickname(db, post.Author)
			fullPost.Author = &user
		case "forum":
			forum, _ := tools.GetForumBySlug(db, post.Forum)
			fullPost.Forum = &forum
		case "thread":
			thread, _ := tools.GetThreadByID(db, int(post.Thread))
			fullPost.Thread = &thread
		}
	}

	tools.ObjectResponce(ctx, http.StatusOK, fullPost)
	return
}

func UpdatePost(ctx *fasthttp.RequestCtx) {
	id, _ := strconv.Atoi(ctx.UserValue("id").(string))

	db := database.Connection

	post, err := tools.GetPostByID(db, id)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Post with id '%d' not found", id)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	newPost := models.Post{}
	body := ctx.Request.Body()
	_ = newPost.UnmarshalJSON(body)

	if _, err := db.Exec(
		"UPDATE Post SET "+
			"message = COALESCE(NULLIF($1, ''), message) "+
			"WHERE id = $2",
		newPost.Message,
		id,
	); err != nil {
		log.Println(err)
		return
	}

	post, _ = tools.GetPostByID(db, id)
	tools.ObjectResponce(ctx, http.StatusOK, post)
	return
}
