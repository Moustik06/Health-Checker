package worker

import (
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	pb "github.com/moustik06/healthchecker/gen/go/health"
	"github.com/moustik06/healthchecker/internal/worker/pool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"time"
)

type server struct {
	pb.UnimplementedHealthCheckerServer
	redisClient *redis.Client
}

func NewGRPCServer(rdb *redis.Client) *grpc.Server {
	s := grpc.NewServer()
	srv := &server{redisClient: rdb}
	pb.RegisterHealthCheckerServer(s, srv)
	return s
}

func (s *server) Check(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	log.Printf("Appel gRPC reçu pour vérifier %d URLs", len(req.Urls))

	var urlsToActuallyCheck []string
	cachedResults := make(map[string]*pb.HealthCheckResult)
	finalResponse := &pb.HealthCheckResponse{Results: make(map[string]*pb.HealthCheckResult)}
	const cacheTTL = 15 * time.Second

	for _, u := range req.Urls {
		cacheKey := "healthcheck:" + u
		cachedValue, err := s.redisClient.Get(ctx, cacheKey).Result()

		if err == redis.Nil {
			CacheMisses.Inc()
			urlsToActuallyCheck = append(urlsToActuallyCheck, u)
		} else if err != nil {
			log.Printf("ERREUR Redis Get pour '%s': %v", u, err)
			CacheMisses.Inc()
			urlsToActuallyCheck = append(urlsToActuallyCheck, u)
		} else {
			CacheHits.Inc()
			var cachedRes pb.HealthCheckResult
			if err := json.Unmarshal([]byte(cachedValue), &cachedRes); err == nil {
				cachedResults[u] = &cachedRes
			} else {
				log.Printf("ERREUR unmarshal cache pour '%s': %v", u, err)
				CacheMisses.Inc()
				urlsToActuallyCheck = append(urlsToActuallyCheck, u)
			}
		}
	}
	log.Printf("Cache hits: %d, Cache misses: %d", len(cachedResults), len(urlsToActuallyCheck))

	if len(urlsToActuallyCheck) > 0 {

		metricsProvider := NewPrometheusMetricsProvider()
		workerPool, err := pool.New(10, time.Second*10, metricsProvider)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "échec de la création du pool: %v", err)
		}

		workerPool.Run(ctx)
		go workerPool.GenerateJobs(urlsToActuallyCheck)
		go workerPool.Wait()

		for result := range workerPool.Results() {
			checkResult := &pb.HealthCheckResult{}
			if result.Status == "success" {
				checkResult.Status = pb.CheckStatus_CHECK_STATUS_OK
			} else {
				checkResult.Status = pb.CheckStatus_CHECK_STATUS_ERROR
				checkResult.ErrorMessage = result.ErrorMsg
			}

			finalResponse.Results[result.URL] = checkResult

			resJSON, err := json.Marshal(checkResult)
			if err == nil {
				cacheKey := "healthcheck:" + result.URL
				if err := s.redisClient.Set(ctx, cacheKey, resJSON, cacheTTL).Err(); err != nil {
					log.Printf("ERREUR Redis Set pour '%s': %v", result.URL, err)
				}
			}
		}
	}

	for url, result := range cachedResults {
		finalResponse.Results[url] = result
	}

	return finalResponse, nil
}
