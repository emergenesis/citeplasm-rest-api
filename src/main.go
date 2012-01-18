package main

import (
    "web"
    "json"
    "strings"
)

type Resource struct {
    Label string `json:"label"`
    Uri string `json:"uri"`
}

type MessageSuccess struct {
    Msg string `json:"msg"`
    Results []Resource `json:"results"`
}

type MessageError struct {
    Code int `json:"code"`
    Message string `json:"msg"`
}

func ( msg *MessageError ) Json () string {
    j, _ := json.MarshalIndent(msg, "", "    ")
    return string(j)
}

func isAuthenticated( ctx *web.Context ) bool {
    var error MessageError
    authHeader := ctx.Request.Headers.Get("Authorization")

    // ensure header was provided
    if authHeader == "" {
        error = MessageError{401,"You must authenticate prior to accessing this resource."}
    } else {
        // parse header to ensure GDS key:value
        authFields := strings.Fields(authHeader)
        if len(authFields) != 2 {
            error = MessageError{401,"The Authenticate header must be of the form 'GDS username:signature'."}
        } else {
            // ensure appropriate auth method
            if authFields[0] != "GDS" {
                error = MessageError{401,"The Authenticate header must be of the form 'GDS username:signature'."}
            } else {
                // parse key:value
                keyValue := strings.Split(authFields[1], ":")
                if len(keyValue) != 2 {
                    error = MessageError{401,"The Authenticate header must be of the form 'GDS username:signature'."}
                } else {
                    // TODO ensure key exists and is valid user
                    // TODO validate value is as expected
                }
            }
        }
    }

    if error.Code != 0 {
        ctx.SetHeader("WWW-Authenticate", "GDS realm=\"http://api.citeplasm.com/v1.0\"", true)
        ctx.Abort(401, error.Json())
        return false
    }

    return true
}

func main() {

    // redirect / to current version (e.g. /v1.0)
    web.Get("/", func ( ctx *web.Context ) {
        ctx.Redirect(301, "/v1.0")
    })

    web.Get("/v1.0", func ( ctx *web.Context ) {
        ctx.SetHeader("Content-type", "application/json", true)
        providers := Resource{"providers", "/v1.0/providers"}
        resources := Resource{"resources", "/v1.0/resources"}
        msg := MessageSuccess{"success", []Resource{providers,resources}}

        b, err := json.MarshalIndent(msg, "", "    ")
        if err == nil {
            ctx.Write(b)
        } else {
            ctx.Abort(500, "internal error: " + err.String())
        }
    })

    // TODO: GET /users
    // TODO: POST /users

    // TODO: GET /users/id
    // TODO: PUT /users/id
    // TODO: DELETE /users/id

    // GET /users/id/texts
    web.Get("/v1.0/users/(.+)/texts", func(ctx *web.Context, user string) {
        ctx.SetHeader("Content-type", "application/json", true)
        if ! isAuthenticated(ctx) {
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

    web.Run("0.0.0.0:9999")
}
