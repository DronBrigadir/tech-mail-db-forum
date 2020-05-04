package controllers

import (
	"fmt"
	"github.com/dronbrigadir/tech-mail-db-forum/internal/database"
	"github.com/dronbrigadir/tech-mail-db-forum/internal/models"
	"github.com/dronbrigadir/tech-mail-db-forum/tools"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func GetPostDetails(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	db := database.Connection

	post, err := tools.GetPostByID(db, id)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Post with id '%d' not found", id)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	query := r.URL.Query()
	related := strings.Split(query.Get("related"), ",")

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

	tools.ObjectResponce(w, http.StatusOK, fullPost)
	return
}

func UpdatePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])

	db := database.Connection

	post, err := tools.GetPostByID(db, id)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Post with id '%d' not found", id)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	newPost := models.Post{}
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
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
	tools.ObjectResponce(w, http.StatusOK, post)
	return
}
