package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/moustik06/healthchecker/internal/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const (
	port        = ":50051"
	metricsPort = ":9091"
	redisAddr   = "redis:6379"
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

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatalf("Impossible de se connecter à Redis sur %s: %v", redisAddr, err)
	}
	log.Printf("Connecté à Redis sur %s", redisAddr)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Échec de l'écoute: %v", err)
	}

	grpcServer := worker.NewGRPCServer(rdb)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		fmt.Printf("Démarrage du serveur gRPC sur %s\n", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Le serveur gRPC a échoué: %v", err)
		}
	}()

	<-stop

	log.Println("Arrêt du serveur gRPC...")
	grpcServer.GracefulStop()
	log.Println("Serveur gRPC arrêté.")
}
