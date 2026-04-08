package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/coffeegraph/coffeegraph/internal/project"
	"github.com/coffeegraph/coffeegraph/internal/server"
	"github.com/coffeegraph/coffeegraph/web"
)

// RunVisualize starts the visualization server and opens the browser.
func RunVisualize() error {
	root, err := project.FindRoot("")
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := server.NewVisualizeServer(root)
	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe(ctx, web.FS)
	}()

	// Brief wait for the server to start before opening the browser.
	time.Sleep(400 * time.Millisecond)
	url := "http://" + srv.Addr()
	if err := openBrowser(url); err != nil {
		fmt.Fprintf(os.Stderr, "could not open browser: %v\n", err)
	}
	fmt.Printf("Visualizer running at %s (Ctrl+C to stop)\n", url)

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil {
			return err
		}
	}
	return nil
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}
