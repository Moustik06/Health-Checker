package main

import (
	"context"
	"fmt"
	pb "github.com/moustik06/healthchecker/gen/go/health"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"time"
)

func main() {
	fmt.Println("Lancement de la gateway...")

	const workerAddr = "localhost:50051"

	conn, err := grpc.NewClient(workerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Échec de la connexion: %v", err)
	}
	defer conn.Close()

	c := pb.NewHealthCheckerClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	urls := []string{"https://google.com", "https://github.com"}
	log.Printf("Envoi de la requête gRPC avec les URLs: %v", urls)

	response, err := c.Check(ctx, &pb.HealthCheckRequest{Urls: urls})
	if err != nil {
		log.Fatalf("L'appel RPC a échoué: %v", err)
	}

	log.Println("Appel gRPC réussi !")
	log.Printf("Réponse reçue: %v", response.Results)
}
