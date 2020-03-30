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
)

func UserCreate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nickname := vars["nickname"]

	user := models.User{}
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
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
		existingUser, _ := tools.GetUserByNickname(db, nickname)

		tools.ObjectResponce(w, http.StatusConflict, existingUser)
		return
	}

	newUser, _ := tools.GetUserByNickname(db, nickname)

	tools.ObjectResponce(w, http.StatusCreated, newUser)
	return
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nickname := vars["nickname"]

	db := database.Connection

	user, err := tools.GetUserByNickname(db, nickname)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("User with nickname '%s' not found", nickname)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	tools.ObjectResponce(w, http.StatusOK, user)
	return
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	nickname := vars["nickname"]

	db := database.Connection

	user, err := tools.GetUserByNickname(db, nickname)
	if err != nil {
		e := models.Error{Message: fmt.Sprintf("User with nickname '%s' not found", nickname)}
		tools.ObjectResponce(w, http.StatusNotFound, e)
		return
	}

	newUser := models.User{}
	body, _ := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
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
		log.Println(err)
		e := models.Error{Message: fmt.Sprintf("User with email '%s' already exists", newUser.Email)}
		tools.ObjectResponce(w, http.StatusConflict, e)
		return
	}

	user, err = tools.GetUserByNickname(db, nickname)
	if err != nil {
		log.Println(err)
		return
	}

	tools.ObjectResponce(w, http.StatusOK, user)
	return
}
