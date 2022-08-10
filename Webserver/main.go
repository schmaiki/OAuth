package main

import (
	"log"
	"net/http"
)

func main() {
	Server := http.NewServeMux()
	RequestHandler := http.RedirectHandler("http://localhost", 307)

	Server.Handle("/", RequestHandler)
	log.Print("Server gestartet")

	err := http.ListenAndServe(":5000", Server)
	if err != nil {
		return
	}

}
