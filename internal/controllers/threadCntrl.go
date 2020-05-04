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
	"time"
)

func CreateThreadPost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	db := database.Connection

	tx, err := db.Begin()
	if err != nil {
		log.Println(err)
		return
	}

	defer tx.Rollback()

	thread, err := tools.GetThreadBySlugOrID(tx, slugOrId)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Thread with slug_or_id '%s' not found", slugOrId)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	posts := models.Posts{}
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	_ = posts.UnmarshalJSON(body)

	var querySql strings.Builder
	querySql.WriteString("INSERT INTO Post (parent, author, message, forum, thread, created) VALUES ")
	var vals []interface{}

	timeNow := time.Now()

	for index, post := range posts {
		if index != 0 {
			querySql.WriteString(",")
		}

		// checking for the user's existence
		user, err := tools.GetUserByNickname(tx, post.Author)
		if err != nil {
			e := models.Error{Message: fmt.Sprintf("User with nickname '%s' not found", post.Author)}
			tools.ObjectResponce(w, http.StatusNotFound, e)
			return
		}

		// checking for the post's existence
		if res := tools.IsPostExists(tx, int(post.Parent), int(thread.ID)); post.Parent != 0 && !res {
			e := models.Error{Message: fmt.Sprintf("There is no parent post withd id '%d'", post.Parent)}
			tools.ObjectResponce(w, http.StatusConflict, e)
			return
		}

		post.Created = timeNow

		querySql.WriteString(fmt.Sprintf("(NULLIF($%d, 0), $%d, $%d, $%d, $%d, $%d)", (index*6)+1, (index*6)+2, (index*6)+3, (index*6)+4, (index*6)+5, (index*6)+6))
		vals = append(vals, post.Parent, user.Nickname, post.Message, thread.Forum, thread.ID, post.Created)
	}

	postsCreated := models.Posts{}
	if len(vals) == 0 {
		_ = tx.Commit()
		tools.ObjectResponce(w, http.StatusCreated, postsCreated)
		return
	}

	querySql.WriteString(" RETURNING id, coalesce(parent, 0), author, message, forum, thread, created;")

	rows, err := tx.Query(querySql.String(), vals...)
	if err != nil {
		log.Println("After tx.query: ", err)
		return
	}

	for rows.Next() {
		post := models.Post{}
		_ = rows.Scan(&post.ID, &post.Parent, &post.Author, &post.Message, &post.Forum, &post.Thread, &post.Created)
		postsCreated = append(postsCreated, post)
	}

	rows.Close()

	err = tx.Commit()
	if err != nil {
		log.Println("After commit post create", err)
		return
	}

	tools.ObjectResponce(w, http.StatusCreated, postsCreated)
	return
}

func GetThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	db := database.Connection

	thread, err := tools.GetThreadBySlugOrID(db, slugOrId)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Thread with slug_or_id '%s' not found", slugOrId)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	tools.ObjectResponce(w, http.StatusOK, thread)
	return
}

func UpdateThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	db := database.Connection

	thread, err := tools.GetThreadBySlugOrID(db, slugOrId)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Thread with slug_or_id '%s' not found", slugOrId)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	newThread := models.Thread{}
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	_ = newThread.UnmarshalJSON(body)

	_, err = db.Exec(
		"UPDATE Thread SET "+
			"title = COALESCE(NULLIF($1, ''), title), "+
			"message = COALESCE(NULLIF($2, ''), message) "+
			"WHERE id = $3",
		newThread.Title,
		newThread.Message,
		thread.ID,
	)

	if err != nil {
		log.Println(err)
		return
	}

	thread, err = tools.GetThreadBySlugOrID(db, slugOrId)
	if err != nil {
		log.Println(err)
		return
	}

	tools.ObjectResponce(w, http.StatusOK, thread)
	return
}

func GetThreadPosts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	db := database.Connection

	thread, err := tools.GetThreadBySlugOrID(db, slugOrId)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Thread with slug_or_id '%s' not found", slugOrId)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since, _ := strconv.Atoi(query.Get("since"))
	sort := query.Get("sort")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	if limit == 0 {
		limit = 100
	}

	sqlQuery := tools.GetQueryForThreadPosts(since, desc, sort)

	rows, err := db.Query(sqlQuery, thread.ID, limit)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	posts := models.Posts{}
	for rows.Next() {
		post := models.Post{}
		_ = rows.Scan(&post.Author, &post.Created, &post.Forum, &post.ID, &post.Message, &post.Thread, &post.Parent)
		posts = append(posts, post)
	}

	tools.ObjectResponce(w, http.StatusOK, posts)
	return
}

func VoteThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slugOrId := vars["slug_or_id"]

	db := database.Connection

	thread, err := tools.GetThreadBySlugOrID(db, slugOrId)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Thread with slug_or_id '%s' not found", slugOrId)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	vote := models.Vote{}
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	_ = vote.UnmarshalJSON(body)

	if !tools.IsUserExists(db, vote.Nickname) {
		e := models.Error{Message: fmt.Sprintf("User with nickname '%s' not found", vote.Nickname)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	_, err = db.Exec(
		"INSERT INTO Vote (threadID, author, voice) VALUES ($1, $2, $3) "+
			"ON CONFLICT ON CONSTRAINT unique_vote DO UPDATE "+
			"SET voice = EXCLUDED.voice;",
		thread.ID,
		vote.Nickname,
		vote.Voice,
	)
	if err != nil {
		log.Println(err)
		return
	}

	thread, err = tools.GetThreadBySlugOrID(db, slugOrId)
	if err != nil {
		log.Println(err)
		return
	}
	tools.ObjectResponce(w, http.StatusOK, thread)
	return
}
