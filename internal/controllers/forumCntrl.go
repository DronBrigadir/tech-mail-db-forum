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
	"time"
)

func ForumCreate(w http.ResponseWriter, r *http.Request) {
	forum := models.Forum{}
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	_ = forum.UnmarshalJSON(body)

	db := database.Connection

	// checking for the user's existence
	user, err := tools.GetUserByNickname(db, forum.User)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("User '%s' not found", forum.User)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	// checking for the forum's existence
	if _, err := db.Exec("INSERT INTO Forum (title, forumUser, slug) VALUES ($1, $2, $3)", forum.Title, user.Nickname, forum.Slug); err != nil {
		existingForum, _ := tools.GetForumBySlug(db, forum.Slug)

		tools.ObjectResponce(w, http.StatusConflict, existingForum)
		return
	}

	newForum, _ := tools.GetForumBySlug(db, forum.Slug)

	tools.ObjectResponce(w, http.StatusCreated, newForum)
	return
}

func ForumGet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	slug := vars["slug"]

	db := database.Connection

	forum, err := tools.GetForumBySlug(db, slug)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Forum with slug '%s' not found", slug)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	tools.ObjectResponce(w, http.StatusOK, forum)
	return
}

func CreateForumThread(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	forumSlug := vars["slug"]

	thread := models.Thread{}
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	_ = thread.UnmarshalJSON(body)
	thread.Forum = forumSlug

	db := database.Connection

	// checking for the user's existence
	user, err := tools.GetUserByNickname(db, thread.Author)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("User '%s' not found", thread.Author)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	// checking for the forum's existence
	forum, err := tools.GetForumBySlug(db, forumSlug)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Forum with slug '%s' not found", forumSlug)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	if thread.Created.IsZero() {
		thread.Created = time.Now()
	}

	var insertedID int
	if err := db.QueryRow(
		`INSERT INTO Thread (title, author, forum, message, slug, created)
 			  VALUES ($1, $2, $3, $4, NULLIF($5, ''), $6) RETURNING id`,
		thread.Title,
		user.Nickname,
		forum.Slug,
		thread.Message,
		thread.Slug,
		thread.Created).Scan(&insertedID); err != nil {
		existingThread, _ := tools.GetThreadBySlug(db, thread.Slug)

		tools.ObjectResponce(w, http.StatusConflict, existingThread)
		return
	}

	newThread, _ := tools.GetThreadByID(db, insertedID)

	tools.ObjectResponce(w, http.StatusCreated, newThread)
	return
}

func GetForumThreads(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	forumSlug := vars["slug"]

	db := database.Connection

	if !tools.IsForumExists(db, forumSlug) {
		e := models.Error{Message: fmt.Sprintf("Forum with slug '%s' not found", forumSlug)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since := query.Get("since")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	if limit == 0 {
		limit = 100
	}

	order := "ASC"
	if desc == true {
		order = "DESC"
	}

	comparison := ""
	if since != "" {
		if desc {
			comparison = fmt.Sprintf("AND created <= '%s'::timestamptz", since)
		} else {
			comparison = fmt.Sprintf("AND created >= '%s'::timestamptz", since)
		}
	}

	sqlQuery := fmt.Sprintf(
		"SELECT id, title, author, forum, message, votes, coalesce(slug, ''), created "+
			"FROM Thread "+
			"WHERE forum = $1 %s "+
			"ORDER BY created %s "+
			"LIMIT $2",
		comparison,
		order,
	)

	rows, err := db.Query(
		sqlQuery,
		forumSlug,
		limit,
	)

	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	threads := models.Threads{}
	for rows.Next() {
		thread := models.Thread{}
		_ = rows.Scan(&thread.ID, &thread.Title, &thread.Author, &thread.Forum, &thread.Message, &thread.Votes, &thread.Slug, &thread.Created)
		threads = append(threads, thread)
	}

	tools.ObjectResponce(w, http.StatusOK, threads)
	return
}

func GetForumUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	forumSlug := vars["slug"]

	db := database.Connection

	if !tools.IsForumExists(db, forumSlug) {
		e := models.Error{Message: fmt.Sprintf("Forum with slug '%s' not found", forumSlug)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	query := r.URL.Query()
	limit, _ := strconv.Atoi(query.Get("limit"))
	since := query.Get("since")
	desc, _ := strconv.ParseBool(query.Get("desc"))

	if limit == 0 {
		limit = 100
	}

	order := "ASC"
	if desc == true {
		order = "DESC"
	}

	sinceStr := ""
	if since != "" {
		comparisonSign := ">"
		if desc {
			comparisonSign = "<"
		}
		sinceStr = fmt.Sprintf("AND f.nickname %s '%s'", comparisonSign, since)
	}

	sqlQuery := fmt.Sprintf(
		"SELECT u.nickname, u.fullname, u.about, u.email "+
			"FROM ForumUser AS f "+
			"JOIN Users as u ON u.nickname = f.nickname "+
			"WHERE f.slug = $1 %s "+
			"ORDER BY f.nickname %s "+
			"LIMIT $2",
		sinceStr,
		order,
	)

	rows, err := db.Query(
		sqlQuery,
		forumSlug,
		limit,
	)

	if err != nil {
		log.Println(err)
		return
	}

	defer rows.Close()

	users := models.Users{}
	for rows.Next() {
		user := models.User{}
		_ = rows.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
		users = append(users, user)
	}

	tools.ObjectResponce(w, http.StatusOK, users)
	return
}
