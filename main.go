package main

import (
	"log"
	"os"

	"github.com/murphybytes/shapes/search"
)

func main() {
	g, err := readGrid(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatalf("Program exited %q", err)
	}
	s, err := search.New(g)
	if err != nil {
		log.Fatalf("search returned error %q", err)
	}
	s.Print(os.Stdout)
}
