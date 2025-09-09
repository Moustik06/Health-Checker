package main

import (
	"fmt"
	pb "github.com/moustik06/healthchecker/gen/go/health"
	"github.com/moustik06/healthchecker/internal/worker"
	"google.golang.org/grpc"
	"log"
	"net"
)

const (
	port = ":50051"
)

func main() {
	fmt.Println("Lancement du worker...")

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Échec de l'écoute: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterHealthCheckerServer(s, &worker.Server{})

	log.Printf("Serveur gRPC démarré et à l'écoute sur %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Échec du démarrage du serveur: %v", err)
	}
}
