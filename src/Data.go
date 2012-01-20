package main

import (
    //"godis"
    "encoding/json"
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

