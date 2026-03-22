package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"time"

	"gallery/internal/server"
)

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		cmd = "xdg-open"
		args = []string{url}
	}

	if len(args) == 0 {
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	gallery, err := server.New(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Errore: %v\n", err)
		os.Exit(1)
	}

	http.HandleFunc("/", gallery.HandleIndex)
	http.HandleFunc("/api/files", gallery.HandleFiles)
	http.HandleFunc("/api/search", gallery.HandleSearch)
	http.HandleFunc("/api/folders", gallery.HandleFolders)
	http.HandleFunc("/raw/", gallery.HandleRaw)
	http.HandleFunc("/thumb/", gallery.HandleThumb)

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	url := fmt.Sprintf("http://localhost:%s", port)

	fmt.Printf("🖼️  Galleria avviata su %s\n", url)
	fmt.Printf("📁 Cartella: %s\n", gallery.RootDir())
	fmt.Println("Premi Ctrl+C per uscire")

	// Open browser after a short delay to ensure server is ready
	go func() {
		time.Sleep(500 * time.Millisecond)
		if err := openBrowser(url); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Impossibile aprire il browser: %v\n", err)
		}
	}()

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Errore server: %v\n", err)
		os.Exit(1)
	}
}
