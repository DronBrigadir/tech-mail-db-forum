package tools

import (
	"fmt"
	"github.com/dronbrigadir/tech-mail-db-forum/internal/database"
	"github.com/dronbrigadir/tech-mail-db-forum/internal/models"
	"strconv"
)

func GetUserByNickname(db database.TxOrDb, nickname string) (models.User, error) {
	user := models.User{}
	err := db.QueryRow(
		"SELECT nickname, fullname, about, email FROM Users WHERE nickname = $1",
		nickname,
	).Scan(
		&user.Nickname,
		&user.Fullname,
		&user.About,
		&user.Email,
	)

	return user, err
}

func GetUsersByEmailOrNickname(db database.TxOrDb, email, nickname string) (models.Users, error) {
	rows, err := db.Query(
		"SELECT nickname, fullname, about, email FROM Users WHERE email = $1 OR nickname = $2",
		email,
		nickname,
	)
	defer rows.Close()

	users := models.Users{}
	for rows.Next() {
		user := models.User{}
		_ = rows.Scan(&user.Nickname, &user.Fullname, &user.About, &user.Email)
		users = append(users, user)
	}

	return users, err
}

func IsUserExists(db database.TxOrDb, nickname string) bool {
	var tmp string
	err := db.QueryRow("SELECT nickname FROM Users WHERE nickname = $1", nickname).Scan(&tmp)
	return err == nil
}

func GetForumBySlug(db database.TxOrDb, slug string) (models.Forum, error) {
	forum := models.Forum{}
	err := db.QueryRow(
		"SELECT title, forumUser, slug, posts, threads FROM Forum WHERE slug = $1",
		slug,
	).Scan(
		&forum.Title,
		&forum.User,
		&forum.Slug,
		&forum.Posts,
		&forum.Threads,
	)

	return forum, err
}

func IsForumExists(db database.TxOrDb, slug string) bool {
	var tmp string
	err := db.QueryRow("SELECT slug FROM Forum WHERE slug = $1", slug).Scan(&tmp)
	return err == nil
}

func GetThreadBySlug(db database.TxOrDb, slug string) (models.Thread, error) {
	thread := models.Thread{}
	err := db.QueryRow(
		"SELECT id, title, author, forum, message, votes, slug, created FROM Thread WHERE slug = $1",
		slug,
	).Scan(
		&thread.ID,
		&thread.Title,
		&thread.Author,
		&thread.Forum,
		&thread.Message,
		&thread.Votes,
		&thread.Slug,
		&thread.Created,
	)

	return thread, err
}

func GetThreadByID(db database.TxOrDb, id int) (models.Thread, error) {
	thread := models.Thread{}
	err := db.QueryRow(
		"SELECT id, title, author, forum, message, coalesce(votes, 0), coalesce(slug, ''), created FROM Thread WHERE id = $1",
		id,
	).Scan(
		&thread.ID,
		&thread.Title,
		&thread.Author,
		&thread.Forum,
		&thread.Message,
		&thread.Votes,
		&thread.Slug,
		&thread.Created,
	)

	return thread, err
}

func GetThreadBySlugOrID(db database.TxOrDb, slugOrId string) (models.Thread, error) {
	id, err := strconv.Atoi(slugOrId)
	if err != nil {
		return GetThreadBySlug(db, slugOrId)
	}

	return GetThreadByID(db, id)
}

func IsPostExists(db database.TxOrDb, postId int) bool {
	var tmp string
	err := db.QueryRow("SELECT author FROM Post WHERE id = $1", postId).Scan(&tmp)
	return err == nil
}

func IsParentPost(db database.TxOrDb, parentId, threadID int) bool {
	var tmp string
	err := db.QueryRow("SELECT author FROM Post WHERE id = $1 AND thread = $2", parentId, threadID).Scan(&tmp)
	return err == nil
}

func GetPostByID(db database.TxOrDb, id int) (models.Post, error) {
	post := models.Post{}
	err := db.QueryRow(
		"SELECT id, coalesce(parent, 0), author, message, isedited, forum, thread, created FROM Post WHERE id = $1",
		id,
	).Scan(
		&post.ID,
		&post.Parent,
		&post.Author,
		&post.Message,
		&post.IsEdited,
		&post.Forum,
		&post.Thread,
		&post.Created,
	)

	return post, err
}

func GetQueryForThreadPosts(since int, desc bool, sort string) string {
	comparison := ""
	if since != 0 {
		comparisonSign := ">"
		if desc {
			comparisonSign = "<"
		}

		comparison = fmt.Sprintf("and post.id %s %d", comparisonSign, since)
		if sort == "tree" {
			comparison = fmt.Sprintf("and post.path %s (SELECT tree_post.path FROM post AS tree_post WHERE tree_post.id = %d)", comparisonSign, since)
		}
		if sort == "parent_tree" {
			comparison = fmt.Sprintf("and post_roots.path[1] %s (SELECT tree_post.path[1] FROM post AS tree_post WHERE tree_post.id = %d)", comparisonSign, since)
		}
	}

	order := "ASC"
	if desc == true {
		order = "DESC"
	}

	sqlQuery := fmt.Sprintf(
		"SELECT post.author, post.created, post.forum, post.id, post.message, post.thread, coalesce(post.parent, 0) "+
			"FROM post "+
			"WHERE post.thread = $1 %s "+
			"ORDER BY (post.created, post.id) %s "+
			"LIMIT $2",
		comparison,
		order,
	)

	if sort == "tree" {
		sqlQuery = fmt.Sprintf(
			"SELECT post.author, post.created, post.forum, post.id, post.message, post.thread, coalesce(post.parent, 0) "+
				"FROM post "+
				"WHERE post.thread = $1 %s "+
				"ORDER BY (post.path, post.created) %s "+
				"LIMIT $2",
			comparison,
			order,
		)
	} else if sort == "parent_tree" {
		sqlQuery = fmt.Sprintf(
			"SELECT post.author, post.created, post.forum, post.id, post.message, post.thread, coalesce(post.parent, 0) "+
				"FROM post "+
				"WHERE post.thread = $1 AND post.path[1] IN ("+
				"SELECT post_roots.id "+
				"FROM post as post_roots "+
				"WHERE post_roots.id = post_roots.path[1] AND post_roots.thread = post.thread %s "+
				"ORDER BY post_roots.path[1] %s "+
				"LIMIT $2)"+
				"ORDER BY post.path[1] %s, post.path",
			comparison,
			order,
			order,
		)
	}

	return sqlQuery
}
