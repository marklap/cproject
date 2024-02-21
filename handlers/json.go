// JSON utility functions for handlers.
package handlers

import (
	"encoding/json"
	"io"
	"net/http"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error error `json:"error"`
}

// WriteJSONWithIndent writes a value to the writer as JSON with a specific indent setting.
func WriteJSONWithIndent(w io.Writer, v interface{}, indentPrefix, indentIndent string) error {
	enc := json.NewEncoder(w)
	enc.SetIndent(indentPrefix, indentIndent)
	return enc.Encode(v)
}

// WriteJSONCompact writes a value to the writer as compact JSON
func WriteJSONCompact(w io.Writer, v interface{}) error {
	return WriteJSONWithIndent(w, v, "", "")
}

// WriteJSON writes a value to the writer as JSON with readable indent.
func WriteJSON(w io.Writer, v interface{}) error {
	return WriteJSONWithIndent(w, v, "", "  ")
}

// WriteJSONErrorWithStatus writes the error with specific status.
func WriteJSONErrorWithStatus(w http.ResponseWriter, err error, status int) {
	w.WriteHeader(status)
	WriteJSON(w, &ErrorResponse{err})
}

// WriteJSONServerError writes the error to the writer as JSON with an internal server error response code.
func WriteJSONServerError(w http.ResponseWriter, err error) {
	WriteJSONErrorWithStatus(w, err, http.StatusInternalServerError)
}

// WriteJSONBadRequest writes the error to the writer as JSON with a bad request response code.
func WriteJSONBadRequest(w http.ResponseWriter, err error) {
	WriteJSONErrorWithStatus(w, err, http.StatusBadRequest)
}
