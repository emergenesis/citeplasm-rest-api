package main

import (
    "web"
    "json"
)

type Resource struct {
    Label string `json:"label"`
    Uri string `json:"uri"`
}

type MessageSuccess struct {
    Msg string `json:"msg"`
    Results []Resource `json:"results"`
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
        //return "Citeplasm API version 1"
    })

    web.Run("0.0.0.0:9999")
}
