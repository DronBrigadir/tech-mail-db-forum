package tools

import (
	"fmt"
	"net/http"
)

func ObjectResponce(w http.ResponseWriter, status int, body interface{ MarshalJSON() ([]byte, error) }) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	marshalBody, err := body.MarshalJSON()
	if err != nil {
		fmt.Println(err)
		return
	}

	_, _ = w.Write(marshalBody)
}
