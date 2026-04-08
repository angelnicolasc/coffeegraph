// Package server provides the HTTP + WebSocket server that powers the
// browser-based graph visualizer (coffeegraph visualize).
package server

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"strconv"
	"time"
)

// VisualizeServer serves graph.html, /graph.json and WebSocket /ws.
type VisualizeServer struct {
	ProjectRoot string
	port        int
	handler     *GraphHandler
}

// NewVisualizeServer creates a server for the given project.
func NewVisualizeServer(projectRoot string) *VisualizeServer {
	return &VisualizeServer{
		ProjectRoot: projectRoot,
		handler:     NewGraphHandler(projectRoot),
	}
}

// Addr returns the host:port the server is listening on.
func (s *VisualizeServer) Addr() string {
	return net.JoinHostPort("127.0.0.1", strconv.Itoa(s.port))
}

// ListenAndServe picks a port from 7777–7779, starts the file watcher
// for live graph updates, and blocks until ctx is cancelled.
func (s *VisualizeServer) ListenAndServe(ctx context.Context, graphHTML fs.FS) error {
	var ln net.Listener
	var err error
	for _, p := range []int{7777, 7778, 7779} {
		ln, err = net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(p)))
		if err == nil {
			s.port = p
			break
		}
	}
	if ln == nil {
		return fmt.Errorf("could not open server on ports 7777-7779")
	}

	mux := http.NewServeMux()

	// Serve graph.html at /.
	htmlFS, err := fs.Sub(graphHTML, ".")
	if err != nil {
		return err
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		f, ferr := htmlFS.Open("graph.html")
		if ferr != nil {
			http.Error(w, ferr.Error(), http.StatusInternalServerError)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.Copy(w, f)
	})

	// Register JSON + WebSocket handlers.
	s.handler.Register(mux)

	srv := &http.Server{Handler: mux}

	// Graceful shutdown on context cancellation.
	go func() {
		<-ctx.Done()
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutCtx)
	}()

	// Start graph file watcher for live WebSocket pushes.
	s.handler.StartWatching(ctx)

	return srv.Serve(ln)
}
