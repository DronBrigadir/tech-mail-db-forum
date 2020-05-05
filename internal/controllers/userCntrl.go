package controllers

import (
	"fmt"
	"github.com/dronbrigadir/tech-mail-db-forum/internal/database"
	"github.com/dronbrigadir/tech-mail-db-forum/internal/models"
	"github.com/dronbrigadir/tech-mail-db-forum/tools"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
)

func UserCreate(ctx *fasthttp.RequestCtx) {
	nickname := fmt.Sprintf("%v", ctx.UserValue("nickname"))

	user := models.User{}
	body := ctx.Request.Body()
	_ = user.UnmarshalJSON(body)
	user.Nickname = nickname

	db := database.Connection

	if _, err := db.Exec(
		"INSERT INTO Users (nickname, fullname, about, email) VALUES ($1, $2, $3, $4)",
		user.Nickname,
		user.Fullname,
		user.About,
		user.Email,
	); err != nil {
		existingUsers, _ := tools.GetUsersByEmailOrNickname(db, user.Email, nickname)

		tools.ObjectResponce(ctx, http.StatusConflict, existingUsers)
		return
	}

	newUser, _ := tools.GetUserByNickname(db, nickname)

	tools.ObjectResponce(ctx, http.StatusCreated, newUser)
	return
}

func GetUser(ctx *fasthttp.RequestCtx) {
	nickname := fmt.Sprintf("%v", ctx.UserValue("nickname"))

	db := database.Connection

	user, err := tools.GetUserByNickname(db, nickname)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("User with nickname '%s' not found", nickname)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	tools.ObjectResponce(ctx, http.StatusOK, user)
	return
}

func UpdateUser(ctx *fasthttp.RequestCtx) {
	nickname := fmt.Sprintf("%v", ctx.UserValue("nickname"))

	db := database.Connection

	user, err := tools.GetUserByNickname(db, nickname)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("User with nickname '%s' not found", nickname)}
		tools.ObjectResponce(ctx, http.StatusNotFound, e)
		return
	}

	newUser := models.User{}
	body := ctx.Request.Body()
	_ = newUser.UnmarshalJSON(body)

	_, err = db.Exec(
		"UPDATE Users SET "+
			"fullname = COALESCE(NULLIF($1, ''), fullname), "+
			"about = COALESCE(NULLIF($2, ''), about), "+
			"email = COALESCE(NULLIF($3, ''), email) "+
			"WHERE nickname = $4",
		newUser.Fullname,
		newUser.About,
		newUser.Email,
		nickname,
	)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("User with email '%s' already exists", newUser.Email)}
		tools.ObjectResponce(ctx, http.StatusConflict, e)
		return
	}

	user, err = tools.GetUserByNickname(db, nickname)
	if err != nil {
		log.Println(err)
		return
	}

	tools.ObjectResponce(ctx, http.StatusOK, user)
	return
}
