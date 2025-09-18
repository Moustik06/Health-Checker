package pool

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

/**
 * A quoi va servir plusieurs workers ?
 * - Gérer plusieurs requêtes HTTP en parallèle
 * - Améliorer la latence globale
 * - Augmenter le débit (nombre de requêtes traitées par seconde)
 * - Répartir la charge sur le serveur cible
 * - Gérer les pics de trafic

 * De quoi a besoin un worker ?
 * - D'une URL à vérifier -> un job
 * - Pouvoir fournir un résultat -> Meme Result que protobuf ?
 * - D'un client HTTP pour faire la requête
 */

type Job struct {
	URL string
}

type Result struct {
	URL      string
	Status   string // "success" ou "failure"
	ErrorMsg string // Message d'erreur si failure
}

type MetricsProvider interface {
	IncChecksTotal(status string)
	ObserveCheckDuration(duration time.Duration)
}

type WorkerPool struct {
	numWorkers int
	jobs       chan Job
	results    chan Result
	wg         sync.WaitGroup
	httpClient *http.Client
	metrics    MetricsProvider
}

func New(numWorkers int, clientTimeout time.Duration, metrics MetricsProvider) (*WorkerPool, error) {
	pool := &WorkerPool{
		numWorkers: numWorkers,
		jobs:       make(chan Job, numWorkers*5), // x5 pour avoir assez de buffer si on envoie beaucoup de jobs d'un coup
		results:    make(chan Result, numWorkers*5),
		httpClient: &http.Client{
			Timeout: clientTimeout,
		},
		metrics: metrics,
	}

	return pool, nil
}

func (wp *WorkerPool) worker(ctx context.Context, id int) {
	defer wp.wg.Done()
	for {
		select {
		case <-ctx.Done():
			log.Println("Worker s'arrête car le contexte a été annulé. ( timeout ou annulation )")
			return
		case job, ok := <-wp.jobs:
			if !ok {
				return
			}
			start := time.Now() // Pour mesurer la durée de la requête
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, job.URL, nil)

			if err != nil {
				wp.metrics.IncChecksTotal("failure")
				wp.results <- Result{URL: job.URL, Status: "failure", ErrorMsg: err.Error()}
				continue // On passe au job suivant
			}
			resp, err := wp.httpClient.Do(req)
			duration := time.Since(start) // Durée de la requête

			wp.metrics.ObserveCheckDuration(duration)

			if err != nil {
				log.Printf("ERREUR de vérification pour %s: %v", job.URL, err)
				wp.metrics.IncChecksTotal("failure")
				wp.results <- Result{URL: job.URL, Status: "failure", ErrorMsg: err.Error()}
				continue // On passe au job suivant
			}
			resp.Body.Close()
			log.Printf("Vérification réussie pour %s en %v", job.URL, duration)
			wp.metrics.IncChecksTotal("success")
			wp.results <- Result{URL: job.URL, Status: "success"}

		}
	}
}

// Run /** démarre le pool de workers en lançant les goroutines. */
func (wp *WorkerPool) Run(ctx context.Context) {
	for i := 0; i < wp.numWorkers; i++ {
		wp.wg.Add(1)
		go wp.worker(ctx, i)
	}
}
func (wp *WorkerPool) GenerateJobs(urls []string) {
	for _, url := range urls {
		wp.jobs <- Job{URL: url}
	}
	close(wp.jobs) // On ferme le canal des jobs une fois tous les jobs envoyés
}

func (wp *WorkerPool) Wait() {
	wp.wg.Wait()      // On attend que tous les workers aient terminé
	close(wp.results) // On ferme le canal des résultats
}
func (wp *WorkerPool) Results() <-chan Result {
	return wp.results
}
