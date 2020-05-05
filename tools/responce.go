package tools

import (
	"fmt"
	"github.com/valyala/fasthttp"
)

func ObjectResponce(ctx *fasthttp.RequestCtx, status int, body interface{ MarshalJSON() ([]byte, error) }) {
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.Response.SetStatusCode(status)

	marshalBody, err := body.MarshalJSON()
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx.Response.SetBody(marshalBody)
}
