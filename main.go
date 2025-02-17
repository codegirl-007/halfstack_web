package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/russross/blackfriday/v2"
)

var tmpl = template.Must(template.ParseFiles("template.html"))

func renderBlog(w http.ResponseWriter, r *http.Request) {
	// Redirect "/blog" to "/blog/" for consistency.
	if r.URL.Path == "/blog" {
		http.Redirect(w, r, "/blog/", http.StatusMovedPermanently)
		return
	}

	// Trim the "/blog/" prefix to get the file path.
	filename := strings.TrimPrefix(r.URL.Path, "/blog/")
	if filename == "" {
		filename = "index.md" // Default markdown file.
	}
	// Construct the full path to the file in the "blog" directory.
	filepath := fmt.Sprintf("blog/%s.md", strings.ReplaceAll(filename, "/", string(os.PathSeparator)))
	fmt.Println("Loading file:", filepath)

	content, err := os.ReadFile(filepath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	htmlContent := blackfriday.Run(content, blackfriday.WithNoExtensions())
	err = tmpl.Execute(w, struct{ Content template.HTML }{Content: template.HTML(htmlContent)})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func renderShow(w http.ResponseWriter, r *http.Request) {
	// Redirect "/episodes" to "/episodes/" for consistency.
	if r.URL.Path == "/episodes" {
		http.Redirect(w, r, "/episodes/", http.StatusMovedPermanently)
		return
	}

	// Trim the "/episodes/" prefix to get the file path.
	filename := strings.TrimPrefix(r.URL.Path, "/episodes/")
	if filename == "" {
		filename = "index.md" // Default markdown file.
	}
	// Construct the full path to the file in the "episodes" directory.
	filepath := fmt.Sprintf("episodes/%s.md", strings.ReplaceAll(filename, "/", string(os.PathSeparator)))
	fmt.Println("Loading file:", filepath)

	content, err := os.ReadFile(filepath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	htmlContent := blackfriday.Run(content, blackfriday.WithNoExtensions())
	err = tmpl.Execute(w, struct{ Content template.HTML }{Content: template.HTML(htmlContent)})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func renderHome(w http.ResponseWriter, r *http.Request) {
	filename := strings.TrimPrefix(r.URL.Path, "/")
	if filename == "" {
		filename = "index" // Default markdown file.
	}
	// Construct the full path to the file in the "pages" directory.
	filepath := fmt.Sprintf("pages/%s.md", strings.ReplaceAll(filename, "/", string(os.PathSeparator)))
	fmt.Println("Loading file:", filepath)

	content, err := os.ReadFile(filepath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	htmlContent := blackfriday.Run(content, blackfriday.WithNoExtensions())
	err = tmpl.Execute(w, struct{ Content template.HTML }{Content: template.HTML(htmlContent)})
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func main() {
	assetsFileServer := http.FileServer(http.Dir("assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", assetsFileServer))

	// Use patterns ending with "/" to capture subpaths.
	http.HandleFunc("/blog/", renderBlog)
	http.HandleFunc("/episodes/", renderShow)
	http.HandleFunc("/", renderHome)
	port := 8080
	fmt.Printf("Server is running on http://localhost:%d\n", port)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Server is running on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	sig := <-sigChan
	log.Printf("Received signal: %s, shutting down...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Default().Print("Server gracefully stopped")
}
