package models

import "time"

type Post struct {
	ID       int64     `json:"id,omitempty"`
	Parent   int64     `json:"parent,omitempty"`
	Author   string    `json:"author"`
	Message  string    `json:"message"`
	IsEdited bool      `json:"isEdited"`
	Forum    string    `json:"forum,omitempty"`
	Thread   int64     `json:"thread,omitempty"`
	Created  time.Time `json:"created,omitempty"`
}
