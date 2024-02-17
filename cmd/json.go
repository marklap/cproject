package main

import (
	"encoding/json"
	"io"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error error `json:"error"`
}

// WriteJSON writes a value to the writer as JSON
func WriteJSON(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteJSONErrorWithStatus writes the error with specific status.
func WriteJSONErrorWithStatus(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	WriteJSON(w, ErrorResponse{err})
}

// WriteJSONServerError writes the error to the writer as JSON with an internal server error response code.
func WriteJSONServerError(w http.ResponseWriter, err error) {
	WriteJSONErrorWithStatus(w, err, http.StatusInternalServerError)
}

// WriteJSONBadRequest writes the error to the writer as JSON with a bad request response code.
func WriteJSONBadRequest(w http.ResponseWriter, err error) {
	WriteJSONErrorWithStatus(w, err, http.StatusBadRequest)
}
