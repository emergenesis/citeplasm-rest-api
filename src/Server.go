package main

import (
	"net/http"             // powers the main api
        "regexp"               // for parsing URIs
        "log"
)

type Handler struct {
    Method string
    Uri *regexp.Regexp
    Function func(*WebContext)
}

type Server struct {
    HandlerFuncs []Handler
}

type WebContext struct {
    Header http.Header
    Request *http.Request
    conn http.ResponseWriter
}

func NewServer () Server {
    var srv Server
    srv.HandlerFuncs = make([]Handler, 0, 250)
    return srv
}

/************************** Server functions *****************************/

// Start initiates the server
func (srv *Server) Start (host string) {
    //http.Handle("/(.*)", srv)
    log.Printf("Started Citeplasm API on %s\n", host)
    log.Fatal(http.ListenAndServe(host, srv))
}

// ServeHTTP implements http.Handler's ServeHTTP function and is responsible
// for processing all requests to the server.
func (srv *Server) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // generate the WebContext object
    // TODO unravel URL parameters and map somewhere in the WebContext
    ctx := WebContext{response.Header(), request, response}

    // search srv.HandlerFuncs for a compatible match
    match := false
    log.Printf("Looking for match for %s %s (%d possible matches)", request.Method, request.URL.Path, len(srv.HandlerFuncs))
    for i := 0; i < len(srv.HandlerFuncs); i++ {
        log.Printf("...checking %s %s", srv.HandlerFuncs[i].Method, srv.HandlerFuncs[i].Uri.String())
        if srv.HandlerFuncs[i].Method == request.Method && srv.HandlerFuncs[i].Uri.MatchString(request.URL.Path) {
            match = true
            srv.HandlerFuncs[i].Function(&ctx)
            break
        }
    }

    if ! match {
        // return 404
        json := `{ code: 404, msg: "Resource does not exist." }`
        response.WriteHeader(404)
        response.Write([]byte(json))
    }
}

// addRoute is an internal function that adds a new function handler
func (srv *Server) addRoute (method string, uri string, fx func(*WebContext)) {
    re, err := regexp.Compile("^" + uri + "$")
    if err != nil {
        log.Fatal("Error in route regular expression: %q\n", uri)
        return
    }

    h := Handler{method, re, fx}
    srv.HandlerFuncs = append(srv.HandlerFuncs, h)
}

// Get adds a new handler for a GET request to the specified URI.
func (srv *Server) Get (uri string, fx func(*WebContext)) {
    srv.addRoute("GET", uri, fx)
}

// Post adds a new handler for a POST request to the specified URI.
func (srv *Server) Post (uri string, fx func(*WebContext)) {
    srv.addRoute("POST", uri, fx)
}

// Put adds a new handler for a PUT request to the specified URI.
func (srv *Server) Put (uri string, fx func(*WebContext)) {
    srv.addRoute("PUT", uri, fx)
}

// Delete adds a new handler for a DELETE request to the specified URI.
func (srv *Server) Delete (uri string, fx func(*WebContext)) {
    srv.addRoute("DELETE", uri, fx)
}

/************************** WebContext functions *****************************/

// Write adds content to the HTTP response body. If no response code has been
// set, 200 OK is automatically used.
func (ctx *WebContext) Write (body []byte) {
    ctx.conn.Write(body)
}

// Redirect sets the response code indicated and provides a Location header to
// instruct the client to redirect.
func (ctx *WebContext) Redirect ( code int, uri string ) {
    http.Redirect(ctx.conn, ctx.Request, uri, code)
}

// Abort ends the request with an status code and an optional body.
func (ctx *WebContext) Abort ( code int, body []byte ) {
    ctx.conn.WriteHeader(code)
    ctx.Write(body)
}

