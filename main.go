package main

import (
	"log"
	"os"
	"strconv"

	"malki.codes/linjante/server"
)

func main() {

	var port int = 8080

	portRaw := os.Getenv("PORT")
	if portRaw != "" {
		newPort, err := strconv.Atoi(portRaw)
		if err != nil {
			log.Fatal(err)
		}

		port = newPort
	}

	log.Printf("Starting server on port %d", port)

	err := server.RunServer(port)
	if err != nil {
		log.Fatal(err)
	}
}
