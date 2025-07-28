package main

import (
	api "filevault/api/routes"
	"log"
	"net/http"
)

func main(){
	mux := api.InitializeRouteMultiplexer()

	log.Print("Starting http server on port 6000")
	log.Fatal(http.ListenAndServe(":6000", mux))
}