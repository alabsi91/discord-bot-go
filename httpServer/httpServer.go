package httpServer

import (
	"context"
	"discord-bot/utils"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var Log = utils.Log

var (
	server   *http.Server
	wg       sync.WaitGroup
	shutdown = make(chan struct{})
)

// Initialize initializes and starts the HTTP server in a non-blocking way.
func Initialize() {
	config := utils.GetAppConfig()

	mux := http.NewServeMux()

	// Define your handlers
	mux.HandleFunc("GET "+config.Http.Routes.Video+"{fileName}", videoHandler)

	// Create the server
	server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Http.Host, config.Http.Port),
		Handler: mux,
	}

	// Start the server in a goroutine so that it doesn't block
	wg.Add(1)

	go func() {
		defer wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			Log.Fatal("\nError starting the http server:", err.Error())
			Log.Debug(Log.Level.Fatal, err.Error())
		}
	}()
}

// Close shuts down the HTTP server gracefully without shutting down the whole program.
func Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if server != nil {
		if err := server.Shutdown(ctx); err != nil {
			Log.Error("\nServer Shutdown Failed:", err.Error())
			Log.Debug(Log.Level.Error, err.Error())
			return
		}
		close(shutdown)
	}
}

func videoHandler(w http.ResponseWriter, r *http.Request) {
	config := utils.GetAppConfig()

	videoPathWithEscape := r.PathValue("fileName")
	videoPathWithoutEscape, err := url.PathUnescape(videoPathWithEscape)
	if err != nil {
		http.Error(w, "Could not unescape video path", http.StatusInternalServerError)
		Log.Error("\nError unescaping video path:", err.Error())
		Log.Debug(Log.Level.Error, err.Error())
		return
	}
	videoPath := filepath.Join(config.Torrent.DownloadDir, videoPathWithoutEscape)

	// Open the video file
	file, err := os.Open(videoPath)
	if err != nil {
		http.Error(w, "Could not open video file", http.StatusInternalServerError)
		Log.Error("\nError opening video file:", err.Error())
		Log.Debug(Log.Level.Error, err.Error())
		return
	}
	defer file.Close()

	// Get file size
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, "Could not get file info", http.StatusInternalServerError)
		Log.Error("\nError getting file info:", err.Error())
		Log.Debug(Log.Level.Error, err.Error())
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", fileInfo.Name()))
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}
