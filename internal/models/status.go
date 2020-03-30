package models

type Status struct {
	UserQuantity   int64 `json:"user"`
	ForumQuantity  int64 `json:"forum"`
	ThreadQuantity int64 `json:"thread"`
	PostQuantity   int64 `json:"post"`
}
