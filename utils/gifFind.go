package utils

import (
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
)

func GifFind() (GifName string) {
	gifPath := filepath.Join(os.Args[1:]...)

	if strings.HasSuffix(gifPath, ".gif") {
		if _, err := os.Stat(gifPath); err == nil {
			return gifPath
		}
	}

	if gifPath == "" || !isDirectory(gifPath) {
		if envPath := os.Getenv("MONOLISA_DIR"); envPath != "" {
			gifPath = envPath
		} else {
			gifPath, _ = os.Getwd()
		}
	}

	files, err := filepath.Glob(filepath.Join(gifPath, "*.gif"))
	if err != nil || len(files) == 0 {
		log.Fatal(err)
	}

	GifName = files[rand.Intn(len(files))]
	return GifName
}

func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}
