package controllers

import (
	"net/http"
)

func SwaggerUI() http.Handler {
	return http.StripPrefix("/swaggerui/", http.FileServer(http.Dir("./third-party/swagger-ui")))
}

func SwaggerApi(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./api/swagger.yml")
}
