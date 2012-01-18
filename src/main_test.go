package main

import (
    "gospec"
    . "gospec"
    "http"
    "fmt"
    "io/ioutil"
    "json"
)

// ProcessedResponse is a simple container for handling responses.
type ProcessedResponse struct {
    Header http.Header
    Code int
    Body string
}

// createRequest is a basic http.NewRequest wrapper with error handling.
func createRequest( method string, url string ) *http.Request {
    request, err := http.NewRequest( "GET", url, nil )
    if err != nil {
        panic(fmt.Sprintf("Bug in test: cannot construct http.Request from method=GET, url=%q: %s", url, err))
    }
    request.Header.Add("Accept", "application/json")

    return request
}

// do is a basic wrapper for http.Client.Do with error handling.
func do( request *http.Request ) *http.Response {
    client := new(http.Client)
    response, err := client.Do(request)
    if err != nil {
        panic(fmt.Sprintf("Bug in test: cannot run request: GET %s\nError: %v", request.URL.Raw, err))
    }

    return response
}

// getResponseBody is a convenience function for processing response bodies
// from io.Reader to string.
func getResponseBody( response *http.Response ) string {
    body_raw, err := ioutil.ReadAll(response.Body)
    if err != nil {
        panic(fmt.Sprintf("Bug in test: cannot read body from response. Error: %v", err))
    }

    return string(body_raw)
}

// GetRequest simply performs a GET on the specified URL with the Content-type
// set as "application/json".
func GetRequest( url string ) ProcessedResponse {
    request := createRequest( "GET", url )
    response := do(request)
    body := getResponseBody(response)

    return ProcessedResponse{response.Header, response.StatusCode, body}
}

// GetRequestWithAuth performs a GET on the specified URL with the Content-type
// set as "application/json" and a correct Authorization header.
func GetRequestWithAuth ( url string ) ProcessedResponse {
    request := createRequest( "GET", url )
    request.Header.Add("Authorization", "GDS username:signature")
    response := do(request)
    body := getResponseBody(response)

    return ProcessedResponse{response.Header, response.StatusCode, body}
}

// MainSpec is the master specification test for the REST server.
func MainSpec(c gospec.Context) {
    c.Specify("GET /v1.0", func() {
        response := GetRequest ("http://localhost:9999/v1.0")

        c.Specify("returns a status code of 200", func () {
            c.Expect(response.Code, Equals, 200)
        })

        c.Specify("returns a list of available resources", func () {
            var msg MessageSuccess
            err := json.Unmarshal([]byte(response.Body), &msg)
            c.Expect(err, Equals, nil)
            c.Expect(msg.Msg, Equals, "success")
            c.Expect(len(msg.Results), Equals, 2)
        })
    })

    c.Specify("GET /v1.0/users/abcde/texts", func() {
        response := GetRequestWithAuth ("http://localhost:9999/v1.0/users/abcde/texts")

        c.Specify("returns 401 unauthorized when Authorization is not provided", func() {
            response := GetRequest ("http://localhost:9999/v1.0/users/abcde/texts")
            c.Expect(response.Code, Equals, 401)
            c.Expect(response.Header.Get("WWW-Authenticate"), Not(IsNil))
        })

        c.Specify("returns 401 unauthorized when Authorization does not contain two arguments", func() {
            request := createRequest( "GET", "http://localhost:9999/v1.0/users/abcde/texts" )
            request.Header.Add("Authorization", "invalid auth header")
            response := do(request)
            body := getResponseBody(response)
            var msg MessageError
            json.Unmarshal([]byte(body), &msg)

            c.Expect(response.StatusCode, Equals, 401)
            c.Expect(response.Header.Get("WWW-Authenticate"), Not(IsNil))
            c.Expect(msg.Code, Equals, 401)
        })

        c.Specify("returns 401 unauthorized when Authorization does not contain GDS", func() {
            request := createRequest( "GET", "http://localhost:9999/v1.0/users/abcde/texts" )
            request.Header.Add("Authorization", "INVALID onetwothreefour")
            response := do(request)
            body := getResponseBody(response)
            var msg MessageError
            json.Unmarshal([]byte(body), &msg)

            c.Expect(response.StatusCode, Equals, 401)
            c.Expect(response.Header.Get("WWW-Authenticate"), Not(IsNil))
            c.Expect(msg.Code, Equals, 401)
        })

        c.Specify("returns 401 unauthorized when Authorization does not have key:signature format", func() {
            request := createRequest( "GET", "http://localhost:9999/v1.0/users/abcde/texts" )
            request.Header.Add("Authorization", "GDS onetwothreefour")
            response := do(request)
            body := getResponseBody(response)
            var msg MessageError
            json.Unmarshal([]byte(body), &msg)

            c.Expect(response.StatusCode, Equals, 401)
            c.Expect(response.Header.Get("WWW-Authenticate"), Not(IsNil))
            c.Expect(msg.Code, Equals, 401)
        })

        c.Specify("returns 200 when valid credentials are provided", func() {
            c.Expect(response.Code, Equals, 200)
        })
    })
}

