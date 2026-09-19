package main

import (
	"log"
	"math/rand"
	"monolisa/render"
	"path/filepath"
)

func main() {
	files, err := filepath.Glob("*.gif")
	if err != nil {
		log.Fatal(err)
	}
	render.Gif(files[rand.Intn(len(files))])
}
