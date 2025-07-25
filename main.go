package main

import (
	"log"
	"os"
	"strconv"

	"malki.codes/linjante/server"
	"malki.codes/linjante/words"
)

func main() {
	words, err := words.LoadWords()
	if err != nil {
		log.Fatal(err)
	}

	var port int = 8080

	portRaw := os.Getenv("PORT")
	if portRaw != "" {
		port, err = strconv.Atoi(portRaw)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Printf("Starting server on port %d", port)

	server.RunServer(port, words)
	if err != nil {
		log.Fatal(err)
	}
}
