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
	j, _ := json.MarshalIndent(T, "", "    ")
        return j
}

// Resource is a generic reference to a resource.
type Resource struct {
	Label string `json:"label"`
	Uri   string `json:"uri"`
}

// NewResource creates a new Resource from a DbObject
func NewResource(obj DbObject) Resource {
    var r Resource
    r.Label = obj.Label()
    r.Uri = "/v1.0" + obj.Uri()

    return r
}

// MessageSuccess represents a successful request for a resultset.
type MessageSuccess struct {
	Msg     string     `json:"msg"`
	Results []Resource `json:"results"`
}

// Json provides the JSON version of the MessageError in a byte array.
func (msg *MessageSuccess) Json() []byte {
	return Marshal(msg)
}

// MessageError represents a transaction that could not be fulfilled.
type MessageError struct {
	Code    int    `json:"code"`
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

    dbi, err := strconv.Atoi(db)
    if err != nil {
        log.Fatal("Environment variable CITEPLASM_REDIS_DB must be an integer.")
    }

    return godis.New(addr, dbi, pw)
}

// DbObject is the basic interface for all persistent objects
type DbObject interface {
    GetKey() string
    Db() *godis.Client
    Id() string
    Label() string
    Uri() string
}

// SaveHashes pushes one or more documents as a hash to the database
func SaveHashes(objs ...DbObject) error {
    for i := 0; i < len(objs); i++ {
        obj := objs[i]

        // get the DB key for this hash
        key := obj.GetKey()

        // get the reflect.Value and reflect.Type of the obj
        oVal := reflect.ValueOf(obj)
        oTyp := reflect.TypeOf(obj)

        // if it's a pointer to a value, resolve to the value itself
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
            if fv.Type().Kind() != reflect.String {
                continue
            }

            // if the field has no value or if it's the Identifier field, skip it
            if fv.String() != "" && ft.Name != "Identifier" {
                obj.Db().Hset(key, ft.Name, fv.String())
            }
        }

        // insert into the index
        // the list is at "idx:Type" and new value is "Label|Id"
        idxKeyName := "idx:" + oTyp.Name()
        idxKeyValue := obj.Id() + "|" + obj.Label()
        obj.Db().Lpush(idxKeyName, idxKeyValue)
    }

    return nil
}

