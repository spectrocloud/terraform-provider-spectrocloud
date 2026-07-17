// Command apiServerMock runs the mock Palette API as a standalone process.
//
// It exists for manual testing and any external tooling that expects to
// connect to 127.0.0.1:8088 (positive routes) or :8888 (negative routes)
// without going through `go test`. The Go unit tests do NOT use this
// binary — they import tests/mockApiServer/mockserver directly and start
// the servers in-process from TestMain.
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spectrocloud/terraform-provider-spectrocloud/tests/mockApiServer/mockserver"
)

func main() {
	srv, err := mockserver.Start()
	if err != nil {
		log.Fatalf("mock api server failed to start: %v", err)
	}
	log.Printf("Mock API server listening on https://127.0.0.1:%d (positive) and https://127.0.0.1:%d (negative)",
		mockserver.PositivePort, mockserver.NegativePort)

	// Block until interrupted so the shell script continues to work as a
	// long-running background process. `kill <pid>` triggers graceful stop.
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig

	srv.Stop()
	log.Println("Mock API server stopped")
}
