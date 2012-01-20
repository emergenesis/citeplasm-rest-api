package main

import (
    "encoding/json"
)

type Resource struct {
	Label string `json:"label"`
	Uri   string `json:"uri"`
}

type MessageSuccess struct {
	Msg     string     `json:"msg"`
	Results []Resource `json:"results"`
}

type MessageError struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
}

func (msg *MessageError) Json() []byte {
	j, _ := json.MarshalIndent(msg, "", "    ")
	return j
}
func main() {

    var server = NewServer()

	// redirect / to current version (e.g. /v1.0)
	server.Get("/", func(ctx *WebContext) {
		ctx.Redirect(301, "/v1.0")
	})

        server.Get("/v1.0", func(ctx *WebContext) {
		ctx.Header.Set("Content-type", "application/json")
		providers := Resource{"providers", "/v1.0/providers"}
		resources := Resource{"resources", "/v1.0/resources"}
		msg := MessageSuccess{"success", []Resource{providers, resources}}

		b, err := json.MarshalIndent(msg, "", "    ")
		if err == nil {
			ctx.Write(b)
		} else {
			//ctx.Abort(500, "internal error: "+err.String())
		}
	})

	// TODO: GET /users
	// TODO: POST /users

	// TODO: GET /users/id
	// TODO: PUT /users/id
	// TODO: DELETE /users/id

	// GET /users/id/texts
	server.Get("/v1.0/users/(.+)/texts", func(ctx *WebContext) {
		ctx.Header.Set("Content-type", "application/json")
		if ! IsAuthenticated(ctx) {
			return
		}
	})

	// TODO: POST /users/id/texts

	// TODO: GET /users/id/texts/id
	// TODO: PUT /users/id/texts/id
	// TODO: DELETE /users/id/texts/id

	// TODO: GET /users/id/resources
	// TODO: POST /users/id/resources

	// TODO: GET /users/id/resources/id
	// TODO: PUT /users/id/resources/id
	// TODO: DELETE /users/id/resources/id

        server.Start(":9999")
}
