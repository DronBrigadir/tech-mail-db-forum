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
	"time"
)

func CreateThreadPost(ctx *fasthttp.RequestCtx) {
	slugOrId := fmt.Sprintf("%v", ctx.UserValue("slug_or_id"))

	db := database.Connection

	tx, err := db.Begin()
	if err != nil {
		log.Println("Can't begin transaction: ", err)
		return
	}

	defer tx.Rollback()

	thread, err := tools.GetThreadBySlugOrID(tx, slugOrId)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Thread with slug_or_id '%s' not found", slugOrId)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	posts := models.Posts{}
	body := ctx.Request.Body()
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
			tools.ObjectResponce(ctx, http.StatusNotFound, e)
			return
		}

		// checking for the parent post existence
		if res := tools.IsPostExists(tx, int(post.Parent), int(thread.ID)); post.Parent != 0 && !res {
			e := models.Error{Message: fmt.Sprintf("There is no parent post withd id '%d'", post.Parent)}
			tools.ObjectResponce(ctx, http.StatusConflict, e)
			return
		}

		post.Created = timeNow

		querySql.WriteString(fmt.Sprintf("(NULLIF($%d, 0), $%d, $%d, $%d, $%d, $%d)", (index*6)+1, (index*6)+2, (index*6)+3, (index*6)+4, (index*6)+5, (index*6)+6))
		vals = append(vals, post.Parent, user.Nickname, post.Message, thread.Forum, thread.ID, post.Created)
	}

	postsCreated := models.Posts{}
	if len(vals) == 0 {
		_ = tx.Commit()
		tools.ObjectResponce(ctx, http.StatusCreated, postsCreated)
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

	if len(postsCreated) > 0 {
		_, err = tx.Exec(`
		UPDATE forum
		SET posts = posts + $1
		WHERE slug = $2;`,
			len(postsCreated),
			postsCreated[0].Forum,
		)
		if err != nil {
			log.Println("Can't increase forum posts: ", err)
		}
	}

	vals = []interface{}{}
	querySql.Reset()
	querySql.WriteString("INSERT INTO ForumUser(slug, nickname) VALUES ")
	for i, post := range postsCreated {
		if i != 0 {
			querySql.WriteString(",")
		}
		querySql.WriteString(fmt.Sprintf("($%d, $%d)", (i*2)+1, (i*2)+2))
		vals = append(vals, post.Forum, post.Author)
	}
	querySql.WriteString(" ON CONFLICT DO NOTHING")
	_, err = tx.Exec(querySql.String(), vals...)
	if err != nil {
		log.Println("Can't insert into forumUser: ", err)
	}

	err = tx.Commit()
	if err != nil {
		log.Println("After commit post create: ", err)
		return
	}

	tools.ObjectResponce(ctx, http.StatusCreated, postsCreated)
	return
}

func GetThread(ctx *fasthttp.RequestCtx) {
	slugOrId := fmt.Sprintf("%v", ctx.UserValue("slug_or_id"))

	db := database.Connection

	thread, err := tools.GetThreadBySlugOrID(db, slugOrId)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Thread with slug_or_id '%s' not found", slugOrId)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	tools.ObjectResponce(ctx, http.StatusOK, thread)
	return
}

func UpdateThread(ctx *fasthttp.RequestCtx) {
	slugOrId := fmt.Sprintf("%v", ctx.UserValue("slug_or_id"))

	db := database.Connection

	thread, err := tools.GetThreadBySlugOrID(db, slugOrId)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Thread with slug_or_id '%s' not found", slugOrId)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	newThread := models.Thread{}
	body := ctx.Request.Body()
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

	tools.ObjectResponce(ctx, http.StatusOK, thread)
	return
}

func GetThreadPosts(ctx *fasthttp.RequestCtx) {
	slugOrId := fmt.Sprintf("%v", ctx.UserValue("slug_or_id"))

	db := database.Connection

	thread, err := tools.GetThreadBySlugOrID(db, slugOrId)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Thread with slug_or_id '%s' not found", slugOrId)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	limit := ctx.QueryArgs().GetUintOrZero("limit")
	since := ctx.QueryArgs().GetUintOrZero("since")
	desc, _ := strconv.ParseBool(string(ctx.QueryArgs().Peek("desc")))
	sort := string(ctx.QueryArgs().Peek("sort"))

	if limit == 0 {
		limit = 100
	}

	sqlQuery := tools.GetQueryForThreadPosts(int(since), desc, sort)

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

	tools.ObjectResponce(ctx, http.StatusOK, posts)
	return
}

func VoteThread(ctx *fasthttp.RequestCtx) {
	slugOrId := fmt.Sprintf("%v", ctx.UserValue("slug_or_id"))

	db := database.Connection

	thread, err := tools.GetThreadBySlugOrID(db, slugOrId)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("Thread with slug_or_id '%s' not found", slugOrId)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	vote := models.Vote{}
	body := ctx.Request.Body()
	_ = vote.UnmarshalJSON(body)

	if !tools.IsUserExists(db, vote.Nickname) {
		e := models.Error{Message: fmt.Sprintf("User with nickname '%s' not found", vote.Nickname)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
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
	tools.ObjectResponce(ctx, http.StatusOK, thread)
	return
}
