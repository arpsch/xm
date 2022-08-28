package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"golang.org/x/sys/unix"

	api "github.com/arpsch/xm/api/http"
	"github.com/arpsch/xm/comp"
	"github.com/arpsch/xm/store"
)

// InitAndRun initializes the server and runs it
func InitAndRun(dataStore store.DataStore) error {
	ctx := context.Background()

	appl, err := comp.NewApp(
		dataStore,
	)

	if err != nil {
		log.Fatalf("server setup encounterd a fatal error, stopping :%v", err)
		return err
	}

	router := api.NewRouter(appl)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Printf("starting the server, Listening on :8080\n")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, unix.SIGINT, unix.SIGTERM)
	<-quit

	log.Print("Server shutting down")

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxWithTimeout); err != nil {
		log.Fatal("error when shutting down the server ", err)
	}

	log.Print("Server exited")
	return nil
}
