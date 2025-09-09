package worker

import (
	"context"
	pb "github.com/moustik06/healthchecker/gen/go/health"
	"github.com/moustik06/healthchecker/internal/worker/pool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type Server struct {
	pb.UnimplementedHealthCheckerServer
}

func (s *Server) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	log.Printf("Appel gRPC reçu pour Check urls : %d\n", len(req.Urls))

	workerPool, err := pool.New(5, time.Second*10)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "échec de la création du pool: %v", err)
	}

	workerPool.Run(ctx)
	go workerPool.GenerateJobs(req.Urls)
	go workerPool.Wait()
	response := &pb.HealthCheckResponse{
		Results: make(map[string]*pb.HealthCheckResult),
	}

	for result := range workerPool.Results() {
		log.Printf("Résultat pour %s : %s \n", result.URL, result.Status)

		checkResult := &pb.HealthCheckResult{}
		if result.Status == "success" {
			checkResult.Status = pb.CheckStatus_CHECK_STATUS_OK
		} else {
			checkResult.Status = pb.CheckStatus_CHECK_STATUS_ERROR
			checkResult.ErrorMessage = result.ErrorMsg
		}
		response.Results[result.URL] = checkResult
	}

	log.Println("Toutes les vérifications sont terminées.")
	return response, nil
}
