package main

// main is the entry point to the REST API server.
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
		ctx.Write(msg.Json())
	})

        // GET /providers
	server.Get("/v1.0/providers", func(ctx *WebContext) {
		ctx.Header.Set("Content-type", "application/json")
		if ! IsAuthenticated(ctx) {
			return
		}

                db := DbConnect()
                providers := GetProviders(db)
                msg := MessageSuccess{"success", providers}
                ctx.Write(msg.Json())
	})

	// TODO: GET /users
	// TODO: POST /users

	// TODO: GET /users/id
	// TODO: PUT /users/id
	// TODO: DELETE /users/id

	// TODO: GET /users/id/texts
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
