package main

import (
	"github.com/robertkozin/x/htmgo"
	"log"
	"os"
	"path/filepath"
)

func main() {
	gofile, ok1 := os.LookupEnv("GOFILE")
	gopackage, ok2 := os.LookupEnv("GOPACKAGE")

	if !ok1 || !ok2 {
		log.Fatal("GOFILE or GOPACKAGE environment variables not set")
	}

	dir := filepath.Dir(gofile)
	out := filepath.Join(dir, "html.gen.go")

	if err := htmgo.Generate2(dir, out, gopackage); err != nil {
		log.Fatal(err)
	}
}
