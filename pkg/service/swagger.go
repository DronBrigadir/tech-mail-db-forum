package service

import "net/http"

func SwaggerUI() http.Handler {
	return http.StripPrefix("/", http.FileServer(http.Dir("../common/swagger-ui")))
}
