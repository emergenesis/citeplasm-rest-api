package main

import (
	"crypto/hmac"     // for authentication generation
	"crypto/md5"      // for authentication generation
	"encoding/base64" // for authentication generation
	"encoding/json"   // marshal/unmarshal json
	"fmt"             // printing errors, etc.
	"gospec"          // powers the specifications
	. "gospec"        // ditto
	"hash"            // for authentication generation
	"io/ioutil"       // parsing response bodies
	"net/http"        // used to run queries against the main server
	"time"            // Date header
)

// ProcessedResponse is a simple container for handling responses.
type ProcessedResponse struct {
	Header http.Header
	Code   int
	Body   string
}

// LoadFixtures flushes the database and loads the test data
func LoadFixtures() error {
    // connect to the DB
    c := DbConnect()

    // create some incrementers
    c.Set("nxUserId", 1000)
    c.Set("nxProvId", 1000)
    c.Set("nxRsrcId", 1000)
    c.Set("nxTextId", 1000)

    // create a few dummy providers
    p1 := NewProvider(c, "National Library of Medicine")
    p2 := NewProvider(c, "FactCheck.org")
    p3 := NewProvider(c, "OpenLibrary.org")
    SaveHashes(p1, p2, p3)

    return nil
}

func FlushDb() error {
    // connect to the DB
    c := DbConnect()

    // flush its contents
    if err := c.Flushall(); err != nil {
        return err
    }

    return nil
}

// createRequest is a basic http.NewRequest wrapper with error handling.
func createRequest(method string, url string) *http.Request {
	request, err := http.NewRequest("GET", "http://localhost:9999"+url, nil)
	if err != nil {
		panic(fmt.Sprintf("Bug in test: cannot construct http.Request from method=GET, url=%q: %s", url, err))
	}
	currentTime := time.Now().UTC().Format(time.RFC1123)
	request.Header.Add("Accept", "application/json")
	request.Header.Add("Date", currentTime)

	return request
}

// do is a basic wrapper for http.Client.Do with error handling.
func do(request *http.Request) *http.Response {
	client := new(http.Client)
	response, err := client.Do(request)
	if err != nil {
		panic(fmt.Sprintf("Bug in test: cannot run request: GET %s\nError: %v", request.URL.Raw, err))
	}

	return response
}

// getResponseBody is a convenience function for processing response bodies
// from io.Reader to string.
func getResponseBody(response *http.Response) string {
	body_raw, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(fmt.Sprintf("Bug in test: cannot read body from response. Error: %v", err))
	}

	return string(body_raw)
}

// GetRequest simply performs a GET on the specified URL with the Content-type
// set as "application/json".
func GetRequest(url string) ProcessedResponse {
	request := createRequest("GET", url)
	response := do(request)
	body := getResponseBody(response)

	return ProcessedResponse{response.Header, response.StatusCode, body}
}

// CreateSignature generates a GDS authentication signature
func CreateSignature(verb string, body string, date string, uri string) string {
	var (
		signature string
		bodyHash  hash.Hash
		sigHmac   hash.Hash
	)

	// do an MD5 hash of the body
	bodyHash = md5.New()
	bodyHash.Write([]byte(body))

	// compute the signature value
	signature += verb + "\n"
	signature += string(bodyHash.Sum(nil)) + "\n"
	signature += date + "\n"
	signature += uri + "\n"

	// create the hmac
	sigHmac = hmac.NewSHA1([]byte("password"))
	sigHmac.Write([]byte(signature))

	return base64.StdEncoding.EncodeToString(sigHmac.Sum(nil))
}

// GetRequestWithAuth performs a GET on the specified URL with the Content-type
// set as "application/json" and a correct Authorization header.
func GetRequestWithAuth(uri string) ProcessedResponse {
	request := createRequest("GET", uri)
	signature := CreateSignature("GET", "", request.Header.Get("Date"), uri)
	request.Header.Add("Authorization", "GDS username:"+signature)
	response := do(request)
	body := getResponseBody(response)

	return ProcessedResponse{response.Header, response.StatusCode, body}
}

// MainSpec is the master specification test for the REST server.
func MainSpec(c gospec.Context) {
	c.Specify("GET /", func() {
		response := GetRequest("/")

		c.Specify("returns a status code of 200", func() {
			c.Expect(response.Code, Equals, 200)
		})

		c.Specify("returns a list of available resources", func() {
			var msg MessageSuccess
			err := json.Unmarshal([]byte(response.Body), &msg)
			c.Expect(err, Equals, nil)
			c.Expect(msg.Msg, Equals, "success")
			c.Expect(len(msg.Results), Equals, 2)
		})
	})

	c.Specify("GET /providers", func() {

		c.Specify("returns 401 unauthorized when Authorization is not provided", func() {
			response := GetRequest("/providers")
			c.Expect(response.Code, Equals, 401)
			c.Expect(response.Header.Get("WWW-Authenticate"), Not(IsNil))
		})

		c.Specify("returns 401 unauthorized when Authorization does not contain two arguments", func() {
			request := createRequest("GET", "/providers")
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
			request := createRequest("GET", "/providers")
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
			request := createRequest("GET", "/providers")
			request.Header.Add("Authorization", "GDS onetwothreefour")
			response := do(request)
			body := getResponseBody(response)
			var msg MessageError
			json.Unmarshal([]byte(body), &msg)

			c.Expect(response.StatusCode, Equals, 401)
			c.Expect(response.Header.Get("WWW-Authenticate"), Not(IsNil))
			c.Expect(msg.Code, Equals, 401)
		})

		c.Specify("returns 401 unauthorized when key is not a valid username", func() {
			request := createRequest("GET", "/providers")
			request.Header.Add("Authorization", "GDS baduser:signature")
			response := do(request)
			body := getResponseBody(response)
			var msg MessageError
			json.Unmarshal([]byte(body), &msg)

			c.Expect(response.StatusCode, Equals, 401)
			c.Expect(response.Header.Get("WWW-Authenticate"), Not(IsNil))
			c.Expect(msg.Code, Equals, 401)
		})

		c.Specify("returns 401 unauthorized when the signature is not valid", func() {
			request := createRequest("GET", "/providers")
			request.Header.Add("Authorization", "GDS username:signature")
			response := do(request)
			body := getResponseBody(response)
			var msg MessageError
			json.Unmarshal([]byte(body), &msg)

			c.Expect(response.StatusCode, Equals, 401)
			c.Expect(response.Header.Get("WWW-Authenticate"), Not(IsNil))
			c.Expect(msg.Code, Equals, 401)
		})

		c.Specify("returns a list of providers when valid credentials are provided", func() {
		        response := GetRequestWithAuth("/providers")
			c.Expect(response.Code, Equals, 200)

                        var msg MessageSuccess
			err := json.Unmarshal([]byte(response.Body), &msg)
			c.Expect(err, Equals, nil)
			c.Expect(msg.Msg, Equals, "success")
			c.Expect(len(msg.Results), Equals, 3)
		})
	})
}
