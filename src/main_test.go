package main

import (
    "gospec"
    . "gospec"
    "http"
    "fmt"
    "io/ioutil"
    "json"
)

func getRequest( url string ) ( code int, body string ) {
    client := new(http.Client)
    request, err := http.NewRequest( "GET", url, nil )
    if err != nil {
        panic(fmt.Sprintf("Bug in test: cannot construct http.Request from method=GET, url=%q, body=%#v: %s", url, body, err))
    }
    request.Header.Add("Accept", "application/json")

    response, err := client.Do(request)
    if err != nil {
        panic(fmt.Sprintf("Bug in test: cannot run request: GET %s\nError: %v", url, err))
    }

    body_raw, _ := ioutil.ReadAll(response.Body)
    if err != nil {
        panic(fmt.Sprintf("Bug in test: cannot read body from response. Error: %v", err))
    }

    return response.StatusCode, string(body_raw)
}

func MainSpec(c gospec.Context) {
    c.Specify("GET /v1.0", func() {
        statusCode, responseBody := getRequest ("http://localhost:9999/v1.0")

        c.Specify("returns a status code of 200", func () {
            c.Expect(statusCode, Equals, 200)
        })

        c.Specify("returns a list of available resources", func () {
            var msg MessageSuccess
            err := json.Unmarshal([]byte(responseBody), &msg)
            c.Expect(err, Equals, nil)
            c.Expect(msg.Msg, Equals, "success")
            c.Expect(len(msg.Results), Equals, 2)
        })
    })
}

