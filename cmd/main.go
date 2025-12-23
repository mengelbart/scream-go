package main

import (
	"log"

	"github.com/mengelbart/scream-go"
)

func main() {
	tx := scream.NewTx(false)
	log.Printf("got tx: %v", tx)
}
