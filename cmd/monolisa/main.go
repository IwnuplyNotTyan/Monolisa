package main

import (
	"math/rand"
	"monolisa/render"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

func main() {
	gifPath := filepath.Join(os.Args[1:]...)
	if gifPath == "" {
		gifPath, _ = os.Getwd()
	}
	files, err := filepath.Glob(filepath.Join(gifPath, "*.gif"))
	if err != nil || len(files) == 0 {
		log.Fatal(err)
	} else {
		render.Gif(files[rand.Intn(len(files))])
	}
}
