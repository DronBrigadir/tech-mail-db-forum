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
	"time"
)

func ForumCreate(ctx *fasthttp.RequestCtx) {
	forum := models.Forum{}
	body := ctx.Request.Body()
	_ = forum.UnmarshalJSON(body)

	db := database.Connection

	// checking for the user's existence
	user, err := tools.GetUserByNickname(db, forum.User)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("User '%s' not found", forum.User)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	// checking for the forum's existence
	if _, err := db.Exec("INSERT INTO Forum (title, forumUser, slug) VALUES ($1, $2, $3)", forum.Title, user.Nickname, forum.Slug); err != nil {
		existingForum, _ := tools.GetForumBySlug(db, forum.Slug)

		tools.ObjectResponce(ctx, http.StatusConflict, existingForum)
		return
	}

	newForum, _ := tools.GetForumBySlug(db, forum.Slug)

	tools.ObjectResponce(ctx, http.StatusCreated, newForum)
	return
}

func ForumGet(ctx *fasthttp.RequestCtx) {
	slug := fmt.Sprintf("%v", ctx.UserValue("slug"))

	db := database.Connection

	forum, err := tools.GetForumBySlug(db, slug)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Forum with slug '%s' not found", slug)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	tools.ObjectResponce(ctx, http.StatusOK, forum)
	return
}

func CreateForumThread(ctx *fasthttp.RequestCtx) {
	forumSlug := fmt.Sprintf("%v", ctx.UserValue("slug"))

	thread := models.Thread{}
	body := ctx.Request.Body()
	_ = thread.UnmarshalJSON(body)
	thread.Forum = forumSlug

	db := database.Connection

	// checking for the user's existence
	user, err := tools.GetUserByNickname(db, thread.Author)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("User '%s' not found", thread.Author)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	// checking for the forum's existence
	forum, err := tools.GetForumBySlug(db, forumSlug)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Forum with slug '%s' not found", forumSlug)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
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

		tools.ObjectResponce(ctx, http.StatusConflict, existingThread)
		return
	}

	newThread, _ := tools.GetThreadByID(db, insertedID)

	tools.ObjectResponce(ctx, http.StatusCreated, newThread)
	return
}

func GetForumThreads(ctx *fasthttp.RequestCtx) {
	forumSlug := fmt.Sprintf("%v", ctx.UserValue("slug"))

	db := database.Connection

	if !tools.IsForumExists(db, forumSlug) {
		e := models.Error{Message: fmt.Sprintf("Forum with slug '%s' not found", forumSlug)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	limit := ctx.QueryArgs().GetUintOrZero("limit")
	since := string(ctx.QueryArgs().Peek("since"))
	desc := ctx.QueryArgs().GetBool("desc")

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

	tools.ObjectResponce(ctx, http.StatusOK, threads)
	return
}

func GetForumUsers(ctx *fasthttp.RequestCtx) {
	forumSlug := fmt.Sprintf("%v", ctx.UserValue("slug"))

	db := database.Connection

	if !tools.IsForumExists(db, forumSlug) {
		e := models.Error{Message: fmt.Sprintf("Forum with slug '%s' not found", forumSlug)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	limit := ctx.QueryArgs().GetUintOrZero("limit")
	since := string(ctx.QueryArgs().Peek("since"))
	desc, _ := strconv.ParseBool(string(ctx.QueryArgs().Peek("desc")))

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

	tools.ObjectResponce(ctx, http.StatusOK, users)
	return
}
