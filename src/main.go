package main

// main is the entry point to the REST API server.
func main() {

    var server = NewServer()

        server.Get("/", func(ctx *WebContext) {
                // return a set of available resources
		ctx.Header.Set("Content-type", "application/json")
		providers := Resource{"providers", "/v1.0/providers"}
		resources := Resource{"resources", "/v1.0/resources"}
		msg := MessageSuccess{"success", []Resource{providers, resources}}
		ctx.Write(msg.Json())
	})

        // GET /providers
	server.Get("/providers", func(ctx *WebContext) {
                // ensure the user properly authenticated
		if ! IsAuthenticated(ctx) {
			return
		}

                // connect to the DB
                db := DbConnect()

                // fetch a list of providers
                providers := GetProviders(db)

                // create a response message for the providers and write it out
                msg := MessageSuccess{"success", providers}
                ctx.Write(msg.Json())
	})

        // TODO: POST /providers

        // TODO: GET /providers/id
        // TODO: PUT /providers/id
        // TODO: DELETE /providers/id

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

        // start the server on all addresses on port 9999
        server.Start(":9999")
}
