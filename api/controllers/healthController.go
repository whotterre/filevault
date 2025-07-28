package api

import (
	"net/http"
)
// Checks whether the api is resting (pun intended) 
func GetStatus(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	w.Write([]byte(`{"status": "OK" }`))
}

