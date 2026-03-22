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

func killExistingServer(port string) bool {
	url := fmt.Sprintf("http://localhost:%s/api/shutdown", port)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

func waitForServer(port string, timeout time.Duration) bool {
	url := fmt.Sprintf("http://localhost:%s/api/files", port)
	client := &http.Client{Timeout: 500 * time.Millisecond}
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

func main() {
	rootDir := "."
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	url := fmt.Sprintf("http://localhost:%s", port)
	shutdownUrl := url + "/api/shutdown"

	// Check if a server is already running
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(shutdownUrl)
	if err == nil && resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		fmt.Println("🔴 Server già in esecuzione, lo stoppo...")
		
		// Wait for the server to shut down
		time.Sleep(500 * time.Millisecond)
		
		// Verify it's down
		for i := 0; i < 20; i++ {
			_, err := client.Get(url)
			if err != nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	gallery, err := server.New(rootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Errore: %v\n", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", gallery.HandleIndex)
	mux.HandleFunc("/api/files", gallery.HandleFiles)
	mux.HandleFunc("/api/search", gallery.HandleSearch)
	mux.HandleFunc("/api/folders", gallery.HandleFolders)
	mux.HandleFunc("/api/shutdown", gallery.HandleShutdown)
	mux.HandleFunc("/raw/", gallery.HandleRaw)
	mux.HandleFunc("/thumb/", gallery.HandleThumb)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	
	gallery.SetServer(srv)

	fmt.Printf("🖼️  Galleria avviata su %s\n", url)
	fmt.Printf("📁 Cartella: %s\n", gallery.RootDir())
	fmt.Println("Premi Ctrl+C per uscire")

	// Open browser after a short delay to ensure server is ready
	go func() {
		// Wait for server to be ready
		if waitForServer(port, 5*time.Second) {
			if err := openBrowser(url); err != nil {
				fmt.Fprintf(os.Stderr, "⚠️  Impossibile aprire il browser: %v\n", err)
			}
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "Errore server: %v\n", err)
		os.Exit(1)
	}
}
