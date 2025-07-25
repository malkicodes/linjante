package errors

import (
	"encoding/json"
	"net/http"
)

func HandleNotFoundError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	response, _ := json.Marshal(map[string]string{"error": "Not Found"})

	w.Write(response)
}

func HandleUserError(w http.ResponseWriter, err string) {
	w.WriteHeader(http.StatusBadRequest)
	response, _ := json.Marshal(map[string]string{"error": err})

	w.Write(response)
}

func HandleRateLimitError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusTooManyRequests)
	response, _ := json.Marshal(map[string]string{"error": "too many requests"})

	w.Write(response)
}

func HandleServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	response, _ := json.Marshal(map[string]string{"error": err.Error()})

	w.Write(response)
}
