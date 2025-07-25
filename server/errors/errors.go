package errors

import (
	"encoding/json"
	"net/http"
)

func HandleUserError(w http.ResponseWriter, err string) {
	w.WriteHeader(400)
	response, _ := json.Marshal(map[string]string{"error": err})

	w.Write(response)
}

func HandleRateLimitError(w http.ResponseWriter) {
	w.WriteHeader(429)
	response, _ := json.Marshal(map[string]string{"error": "too many requests"})

	w.Write(response)
}

func HandleServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(500)
	response, _ := json.Marshal(map[string]string{"error": err.Error()})

	w.Write(response)
}
