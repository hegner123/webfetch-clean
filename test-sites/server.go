package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	port := flag.Int("port", 8080, "Port to run the server on")
	dir := flag.String("dir", ".", "Directory to serve files from")
	flag.Parse()

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		log.Fatalf("Error resolving directory: %v", err)
	}

	if _, err := os.Stat(absDir); os.IsNotExist(err) {
		log.Fatalf("Directory does not exist: %s", absDir)
	}

	fs := http.FileServer(http.Dir(absDir))

	handler := loggingMiddleware(corsMiddleware(fs))

	addr := fmt.Sprintf(":%d", *port)

	fmt.Printf("Starting test site server...\n")
	fmt.Printf("Serving directory: %s\n", absDir)
	fmt.Printf("Server running at: http://localhost%s\n", addr)
	fmt.Printf("\nAvailable test sites:\n")
	fmt.Printf("  - http://localhost%s/heavy-ads/\n", addr)
	fmt.Printf("  - http://localhost%s/deep-nesting/\n", addr)
	fmt.Printf("  - http://localhost%s/large-content/\n", addr)
	fmt.Printf("  - http://localhost%s/malformed-html/\n", addr)
	fmt.Printf("  - http://localhost%s/script-heavy/\n", addr)
	fmt.Printf("  - http://localhost%s/navigation-heavy/\n", addr)
	fmt.Printf("  - http://localhost%s/minimal-content/\n", addr)
	fmt.Printf("  - http://localhost%s/modern-blog/\n", addr)
	fmt.Printf("  - http://localhost%s/nextjs-ssr/\n", addr)
	fmt.Printf("  - http://localhost%s/vue-app/\n", addr)
	fmt.Printf("  - http://localhost%s/svelte-app/\n", addr)
	fmt.Printf("\nNote: nextjs-ssr, vue-app, and svelte-app require JavaScript execution\n")
	fmt.Printf("      and will return empty/skeleton HTML only (expected behavior)\n")
	fmt.Printf("\nPress Ctrl+C to stop\n\n")

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
