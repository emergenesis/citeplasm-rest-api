package main

import (
	"net/http"             // powers the main api
        "regexp"               // for parsing URIs
        "log"
        "reflect"              // for processing router handlers
)

// Handler represents a function handler for a specified URI. Server's Get,
// Put, Post, and Delete methods create and store Handlers from the information
// provided and use those Handlers to respond to matching HTTP requests.
type Handler struct {
    // Method is the HTTP request method for which this Handler can respond.
    Method string

    // Uri is a regular expression of the URIs to which this Handler can
    // respond.
    Uri *regexp.Regexp

    // Handler is a reflect.Value representation of a function to handle the
    // request.
    Handler reflect.Value
}

// Server represents the HTTP server responsible for processing URIs by a set
// of handlers.
type Server struct {

    // Handlers is an array of registered Handlers the server can use to
    // respond to an HTTP request.
    Handlers []Handler
}

// WebContext represents the context under which a particular Handler is invoked.
type WebContext struct {

    // Header represents the HTTP response headers.
    Header http.Header

    // Request represents the HTTP request, including its header and body.
    Request *http.Request

    // conn is an internal construct used by WebContext functions for rendering
    // or manipulating the response.
    conn http.ResponseWriter
}

// NewServer creates a new HTTP Server.
func NewServer () Server {
    // create a new server, allowing a maximum of 250 URI handlers.
    var srv Server
    srv.Handlers = make([]Handler, 0, 250)
    return srv
}

/************************** Server functions *****************************/

// Start initiates the server
func (srv *Server) Start (host string) {
    log.Printf("Started Citeplasm API on %s\n", host)
    log.Fatal(http.ListenAndServe(host, srv))
}

// ServeHTTP implements http.Handler's ServeHTTP function and is responsible
// for processing all requests to the server.
func (srv *Server) ServeHTTP (response http.ResponseWriter, request *http.Request) {
    // create some convenience variables for use in comparing the actual
    // request to the various request handlers.
    targetMethod := request.Method
    targetUri := request.URL.Path
    targetQuery := request.URL.RawQuery

    // log the request
    if len(request.URL.RawQuery) == 0 {
        log.Println(targetMethod + " " + targetUri)
    } else {
        log.Println(targetMethod + " " + targetUri + "?" + targetQuery)
    }

    // generate the WebContext object
    ctx := WebContext{response.Header(), request, response}

    // set the default headers
    ctx.Header.Set("Content-type", "application/json")

    // search srv.Handlers for a compatible match
    match := false
    for i := 0; i < len(srv.Handlers); i++ {
        // create some convenience variables for comparing this handler to the
        // actual request
        thisMethod := srv.Handlers[i].Method
        thisUri := srv.Handlers[i].Uri

        // if we find a match...
        if thisMethod == targetMethod && thisUri.MatchString(targetUri) {
            // unravel URL parameters and map somewhere in the WebContext
            matchedParams := thisUri.FindStringSubmatch(targetUri)

            // create the args to pass to the handler function
            var args []reflect.Value

            // we always include the context
            args = append(args, reflect.ValueOf(&ctx))

            for _, arg := range matchedParams[1:] {
                args = append(args, reflect.ValueOf(arg))
            }

            // call the function
            // FIXME: this should be called safely
            srv.Handlers[i].Handler.Call(args)

            // we have a match, so we're done
            match = true
            break
        }
    }

    // if there was no matching route, we should return a 404 error
    if ! match {
        err404 := MessageError{404, "Resource does not exist."}
        response.WriteHeader(404)
        response.Write(err404.Json())
    }
}

// addRoute is an internal function that adds a new function handler
func (srv *Server) addRoute (method string, uri string, handler interface{}) {
    // we are only going to store the compiled URI regex
    re, err := regexp.Compile("^" + uri + "$")
    if err != nil {
        log.Fatalf("Error in route regular expression: %s", uri)
        return
    }

    // get the reflect.Type and reflect.Value of the handler
    handlerValue := reflect.ValueOf(handler)
    handlerType := handlerValue.Type()

    // log a fatal error if handler is not a function
    if handlerType.Kind() != reflect.Func {
        log.Fatalf("Handler must be a function for route %s %s . %s", method, uri, handlerType.Kind().String())
        return
    }

    // ensure the handler takes at least 1 arg, that is a ptr, to a WebContext
    if handlerType.NumIn() == 0 ||
       handlerType.In(0).Kind() != reflect.Ptr ||
       handlerType.In(0).Elem() != reflect.TypeOf(WebContext{}) {
        log.Fatalf("Handler function must take a *WebContext as its first parameter in route %s %s", method, uri)
        return
    }

    // create the handler and add it to the server's set of handlers
    h := Handler{method, re, handlerValue}
    srv.Handlers = append(srv.Handlers, h)

}

// Get adds a new handler for a GET request to the specified URI.
func (srv *Server) Get (uri string, handler interface{}) {
    srv.addRoute("GET", uri, handler)
}

// Post adds a new handler for a POST request to the specified URI.
func (srv *Server) Post (uri string, handler interface{}) {
    srv.addRoute("POST", uri, handler)
}

// Put adds a new handler for a PUT request to the specified URI.
func (srv *Server) Put (uri string, handler interface{}) {
    srv.addRoute("PUT", uri, handler)
}

// Delete adds a new handler for a DELETE request to the specified URI.
func (srv *Server) Delete (uri string, handler interface{}) {
    srv.addRoute("DELETE", uri, handler)
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

