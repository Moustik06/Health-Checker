package main

import (
	"context"
	"fmt"
	"github.com/moustik06/healthchecker/internal/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	workerAddr = "health-worker:50051"
	httpPort   = ":8080"
)

func main() {
	fmt.Println("Lancement de la gateway...")

	conn, err := grpc.NewClient(workerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("Impossible de se connecter au service worker: ", err)
	}
	defer conn.Close()

	log.Println("Connecté au service worker gRPC sur ", workerAddr)

	handler := gateway.NewHandler(conn)
	mainMux := http.NewServeMux()
	mainMux.HandleFunc("/check", handler.CheckURLs)

	handlerFinal := gateway.RecoveryMiddleware(mainMux)
	server := &http.Server{
		Addr:    httpPort,
		Handler: handlerFinal,
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := server.ListenAndServe(); err != nil {
			fmt.Printf("Échec du démarrage du serveur HTTP: %v\n", err)
		}
	}()
	<-stop
	log.Println("Arrêt du serveur HTTP...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Erreur lors de l'arrêt du serveur HTTP: %v", err)
	}
	log.Println("Serveur HTTP arrêté.")

}
