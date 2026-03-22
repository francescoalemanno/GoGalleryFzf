package main

import (
	"fmt"
	"net/http"
	"os"

	"gallery/internal/server"
)

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

	fmt.Printf("🖼️  Galleria avviata su http://localhost:%s\n", port)
	fmt.Printf("📁 Cartella: %s\n", gallery.RootDir())
	fmt.Println("Premi Ctrl+C per uscire")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Errore server: %v\n", err)
		os.Exit(1)
	}
}
