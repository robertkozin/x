package main

import (
	"github.com/robertkozin/x/htmgo"
	"log"
	"os"
)

func main() {
	gofile, ok1 := os.LookupEnv("GOFILE")
	gopackage, ok2 := os.LookupEnv("GOPACKAGE")

	if !ok1 || !ok2 {
		log.Fatal("GOFILE or GOPACKAGE environment variables not set")
	}

	if err := htmgo.Generate(gofile, gopackage); err != nil {
		log.Fatal(err)
	}
}
