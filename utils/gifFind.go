package utils

import (
	"math/rand"
	"os"
	"path/filepath"

	"github.com/charmbracelet/log"
)

func GifFind() (GifName string) {
	gifPath := filepath.Join(os.Args[1:]...)
	if gifPath == "" {
		gifPath, _ = os.Getwd()
	}
	files, err := filepath.Glob(filepath.Join(gifPath, "*.gif"))
	if err != nil || len(files) == 0 {
		log.Fatal(err)
	}

	GifName = files[rand.Intn(len(files))]
	return GifName
}
