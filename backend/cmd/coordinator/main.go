package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/christianmz565/microphoto/internal/coordinator"
	"github.com/christianmz565/microphoto/pkg/client/metrics"
	"github.com/christianmz565/microphoto/pkg/client/minio"
	"github.com/christianmz565/microphoto/pkg/client/redis"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("Microphoto Coordinator starting...")

	cfg := coordinator.NewConfig()

	m, err := metrics.InitMetrics("coordinator")
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}
	metrics.StartMetricsServer(cfg.MetricsPort)

	rClient, err := redis.NewClient(cfg.RedisAddr)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}

	mClient, err := minio.NewClient(cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioSSL)
	if err != nil {
		log.Fatalf("Failed to initialize minio: %v", err)
	}

	err = mClient.EnsureBucket(context.Background(), coordinator.BucketName)
	if err != nil {
		log.Fatalf("Failed to ensure bucket: %v", err)
	}

	err = mClient.SetupLifecyclePolicy(context.Background(), coordinator.BucketName)
	if err != nil {
		log.Printf("Warning: Failed to setup lifecycle policy: %v", err)
	}

	orch := coordinator.NewOrchestrator(rClient, mClient, m)
	handler := coordinator.NewHTTPHandler(orch, m, cfg.MaxUploadSize)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	log.Printf("Coordinator listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, corsMiddleware(mux)))
}
