package main

import (
	"context"
	"fmt"
	pb "github.com/moustik06/healthchecker/gen/go/health"
	"github.com/moustik06/healthchecker/internal/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	port        = ":50051"
	metricsPort = ":9091"
)

func main() {
	fmt.Println("Lancement du worker...")

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Printf("Serveur de métriques démarré sur %s\n", metricsPort)
		if err := http.ListenAndServe(metricsPort, nil); err != nil {
			log.Fatalf("Le serveur de métriques a échoué: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Échec de l'écoute: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		s := grpc.NewServer()
		pb.RegisterHealthCheckerServer(s, &worker.Server{})

		log.Printf("Serveur gRPC démarré et à l'écoute sur %s", port)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Échec du démarrage du serveur: %v", err)
		}
	}()
	<-stop
	log.Println("Arrêt du serveur...")
	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := lis.Close(); err != nil {
		log.Fatalf("Erreur lors de l'arrêt du serveur: %v", err)
	}
	log.Println("Serveur arrêté.")
}
