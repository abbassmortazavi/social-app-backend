package utils

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func WriteJson(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		return
	}
}
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJson(w, status, map[string]string{"error": message})
}

func GetPathParamInt(r *http.Request, key string) (int, error) {
	return strconv.Atoi(r.PathValue(key))
}
