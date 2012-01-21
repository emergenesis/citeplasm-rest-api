package main

import (
    "godis"
    "encoding/json"
    "os"            // for env vars
    "strconv"       // for converting string to int and int64 to string
    "log"
    "reflect"       // saving objects to db
)

// Marshal translates a Go type into a JSON byte array.
func Marshal( T interface{} ) []byte {
        // for development and testing, we'll use a prettier output
	j, _ := json.MarshalIndent(T, "", "    ")
        return j
}

// Resource is a generic reference to a resource represented by the API.
type Resource struct {
        // Label is the friendly name for the resource.
	Label string `json:"label"`

        // Uri is the URI for this resource within this API. It is also a
        // unique identifier.
	Uri   string `json:"uri"`
}

// NewResource creates a new Resource from a struct that implements DbObject.
func NewResource(obj DbObject) Resource {
    var r Resource
    r.Label = obj.Label()
    // FIXME: there's probably a better way to integrate the API version.
    r.Uri = "/v1.0" + obj.Uri()

    return r
}

// MessageSuccess represents a successful request for a resultset.
type MessageSuccess struct {

        // Msg is the human-readable response, often just "success"
	Msg     string     `json:"msg"`

        // Results is an array of Resources associated with this message.
	Results []Resource `json:"results"`
}

// Json provides the JSON version of the MessageSuccess in a byte array.
func (msg *MessageSuccess) Json() []byte {
	return Marshal(msg)
}

// MessageError represents a transaction that could not be fulfilled.
type MessageError struct {

        // Code is the error code indicating the error.
	Code    int    `json:"code"`

        // Message is the human-readable error explaining what went wrong.
	Message string `json:"msg"`
}

// Json provides the JSON version of the MessageError in a byte array.
func (msg *MessageError) Json() []byte {
	return Marshal(msg)
}

// DbConnect returns a connection to the Redis database. Connection details can
// be provided through the CITEPLASM_REDIS_ADDR, CITEPLASM_REDIS_DB, and
// CITEPLASM_REDIS_PWD environment variables.
func DbConnect () *godis.Client {
    addr := os.Getenv("CITEPLASM_REDIS_ADDR")
    db := os.Getenv("CITEPLASM_REDIS_DB")
    pw := os.Getenv("CITEPLASM_REDIS_PWD")

    // default db if not provided
    if db == "" {
        db = "0"
    }

    // conver the DB to an integer value, error out if not possible
    dbi, err := strconv.Atoi(db)
    if err != nil {
        log.Fatal("Environment variable CITEPLASM_REDIS_DB must be an integer.")
    }

    return godis.New(addr, dbi, pw)
}

// DbObject is the basic interface for all objects that persist to the database.
type DbObject interface {
    // GetKey returns the database key for the object, e.g. "user:1234" or "prov:5678"
    GetKey() string

    // Db returns a pointer to the database connection.
    Db() *godis.Client

    // Id returns the unique identifier of this object.
    Id() string

    // Label returns the human-friendly name of this object, used in Resource values.
    Label() string

    // Uri returns the API address of the object.
    Uri() string
}

// SaveHashes pushes one or more documents as a hash to the database. All
// variables must have types that implement DbObject.
// FIXME: support transactions
func SaveHashes(objs ...DbObject) error {
    // run through all provided DbObjects and persist them
    for i := 0; i < len(objs); i++ {
        // convenience representation of the current DbObject
        obj := objs[i]

        // get the DB key for this hash
        key := obj.GetKey()

        // get the reflect.Value and reflect.Type of the obj
        oVal := reflect.ValueOf(obj)
        oTyp := reflect.TypeOf(obj)

        // if it's a pointer to a value, resolve to the value and type itself,
        // rather than pointer
        if oTyp.Kind() == reflect.Ptr {
            oTyp = oTyp.Elem()
            oVal = oVal.Elem()
        }

        // cycle through all the fields and insert them into the hash in the db
        oFieldCount := oTyp.NumField()
        for i := 0; i < oFieldCount; i++ {
            // get the field Value and Type
            fv := oVal.Field(i)
            ft := oTyp.Field(i)

            // if it's not a string, we don't care
            // FIXME: convertible types like int should be included as well
            if fv.Type().Kind() != reflect.String {
                continue
            }

            // if the field has no value or if it's the Identifier field, skip it
            if fv.String() != "" && ft.Name != "Identifier" {
                obj.Db().Hset(key, ft.Name, fv.String())
            }
        }

        // insert into the index so it can be found without knowing its key
        // the list is at "idx:Type" (e.g. idx:User) and new value is "Id|Label" (e.g. "1234|johnsmith")
        idxKeyName := "idx:" + oTyp.Name()
        idxKeyValue := obj.Id() + "|" + obj.Label()
        _, err := obj.Db().Lpush(idxKeyName, idxKeyValue)
        if err != nil {
            return err
        }
    }

    // there was obviously no error, so return nil
    return nil
}

