package api

import (
	api "filevault/api/controllers"
	"net/http"
)

func InitializeRouteMultiplexer() *http.ServeMux {
	mux := http.NewServeMux()

	// Routes
	mux.HandleFunc("/status", api.GetStatus)
	mux.HandleFunc("/register", dummy)
	mux.HandleFunc("/login", dummy)
	mux.HandleFunc("/logout", dummy)
	mux.HandleFunc("/users/me", dummy)
	mux.HandleFunc("/files", dummy)
	mux.HandleFunc("/files/:id", dummy)

	return mux
}

func dummy(w http.ResponseWriter, r *http.Request) {

}
