package main

import (
	"net/http"             // powers the main api
        "regexp"               // for parsing URIs
        "log"
        "reflect"              // for processing router handlers
)

type Handler struct {
    Method string
    Uri *regexp.Regexp
    Handler reflect.Value
}

type Server struct {
    Handlers []Handler
}

type WebContext struct {
    Header http.Header
    Request *http.Request
    conn http.ResponseWriter
}

func NewServer () Server {
    var srv Server
    srv.Handlers = make([]Handler, 0, 250)
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
    //log.Printf("Looking for match for %s %s (%d possible matches)", targetMethod, targetUri, len(srv.Handlers))
    for i := 0; i < len(srv.Handlers); i++ {
        thisMethod := srv.Handlers[i].Method
        thisUri := srv.Handlers[i].Uri
        //log.Printf("...checking %s %s", thisMethod, thisUri.String())

        // if we find a match...
        if thisMethod == targetMethod && thisUri.MatchString(targetUri) {
            // unravel URL parameters and map somewhere in the WebContext
            matchedParams := thisUri.FindStringSubmatch(targetUri)

            // create the args
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

    if ! match {
        // return 404
        json := `{ code: 404, msg: "Resource does not exist." }`
        response.WriteHeader(404)
        response.Write([]byte(json))
    } // FIXME: return 404 if there is no match
}

// addRoute is an internal function that adds a new function handler
func (srv *Server) addRoute (method string, uri string, handler interface{}) {
    re, err := regexp.Compile("^" + uri + "$")
    if err != nil {
        log.Fatalf("Error in route regular expression: %s", uri)
        return
    }

    handlerValue := reflect.ValueOf(handler)
    handlerType := handlerValue.Type()

    if handlerType.Kind() != reflect.Func {
        log.Fatalf("Handler must be a function for route %s %s . %s", method, uri, handlerType.Kind().String())
        return
    }

    if handlerType.NumIn() == 0 ||
       handlerType.In(0).Kind() != reflect.Ptr ||
       handlerType.In(0).Elem() != reflect.TypeOf(WebContext{}) {
        log.Fatalf("Handler function must take a *WebContext as its first parameter in route %s %s", method, uri)
        return
    }

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

