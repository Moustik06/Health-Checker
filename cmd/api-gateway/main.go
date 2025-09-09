package main

import (
	"fmt"
	"github.com/moustik06/healthchecker/internal/gateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
)

const (
	workerAddr = "health-worker:50051"
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
	server := http.NewServeMux()
	server.HandleFunc("/check", handler.CheckURLs)
	if err := http.ListenAndServe(":8080", server); err != nil {
		fmt.Printf("Échec du démarrage du serveur HTTP: %v\n", err)
	}

}
