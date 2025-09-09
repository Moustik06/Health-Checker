package worker

import (
	"context"
	pb "github.com/moustik06/healthchecker/gen/go/health"
	"log"
)

type Server struct {
	pb.UnimplementedHealthCheckerServer
}

func (s *Server) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	log.Printf("Appel gRPC reçu pour Check urls : %d\n", len(req.Urls))
	return &pb.HealthCheckResponse{
		Results: make(map[string]*pb.HealthCheckResult),
	}, nil
}
