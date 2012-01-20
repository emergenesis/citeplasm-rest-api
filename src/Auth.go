package main

import (
	"crypto/hmac"     // for authentication verification
	"crypto/md5"      // for authentication verification
	"encoding/base64" // for authentication verification
	"hash"            // for authentication verification
	"io/ioutil"       // to read request bodies
	"strings"         // to parse request headers
)

func IsAuthenticated(ctx *WebContext) bool {
	var error MessageError
	authHeader := ctx.Request.Header.Get("Authorization")

	// ensure header was provided
	if authHeader == "" {
		error = MessageError{401, "You must authenticate prior to accessing this resource."}
	} else {
		// parse header to ensure GDS key:value
		authFields := strings.Fields(authHeader)
		if len(authFields) != 2 {
			error = MessageError{401, "The Authenticate header must be of the form 'GDS username:signature'."}
		} else {
			// ensure appropriate auth method
			if authFields[0] != "GDS" {
				error = MessageError{401, "The Authenticate header must be of the form 'GDS username:signature'."}
			} else {
				// parse key:value
				keyValue := strings.Split(authFields[1], ":")
				if len(keyValue) != 2 {
					error = MessageError{401, "The Authenticate header must be of the form 'GDS username:signature'."}
				} else {
					// ensure key exists and is valid user
					// for now, we use a static username and password
					if keyValue[0] != "username" {
						error = MessageError{401, "The Authenticate header did not contain a valid user."}
					} else {
						// validate value is as expected
						var (
							signature string
							bodyHash  hash.Hash
							sigHmac   hash.Hash
						)

						body, _ := ioutil.ReadAll(ctx.Request.Body)

						// do an MD5 hash of the body
						bodyHash = md5.New()
						bodyHash.Write(body)

						// compute the signature value
						signature += ctx.Request.Method + "\n"
						signature += string(bodyHash.Sum(nil)) + "\n"
						signature += ctx.Request.Header.Get("Date") + "\n"
						signature += ctx.Request.URL.Path + "\n"

						// create the hmac
						sigHmac = hmac.NewSHA1([]byte("password"))
						sigHmac.Write([]byte(signature))

						correctHash := base64.StdEncoding.EncodeToString(sigHmac.Sum(nil))

						if keyValue[1] != correctHash {
							error = MessageError{401, "The Authenticate header did not contain a valid signature."}
						} else {
							// TODO validate date is current to within 15min
						}
					}
				}
			}
		}
	}

	if error.Code != 0 {
		ctx.Header.Set("WWW-Authenticate", "GDS realm=\"http://api.citeplasm.com/v1.0\"")
		ctx.Abort(401, error.Json())
		return false
	}

	return true
}


