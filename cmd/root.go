package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/juststeveking/package-api/pkg/server"
	"github.com/spf13/cobra"
)

var (
	port string

	rootCmd = &cobra.Command{
		Use:   "api",
		Short: "Run the Package API.",
		Run:   runServer,
	}
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&port, "port", "p", "3000", "Port to run the server on")
}

func runServer(cmd *cobra.Command, args []string) {
	srv := server.NewServer()

	go func() {
		fmt.Printf("Running on port :%s\n", port)
		if err := srv.ListenAndServe(port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Could not listen on port %s: %v\n", port, err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	if err := srv.Shutdown(context.Background()); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exiting")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
