package gateway

import (
	"context"
	"encoding/json"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"time"

	pb "github.com/moustik06/healthchecker/gen/go/health"
)

/**
 * Que va recevoir le handler  sur /check ?
 * - Une liste d'URL à checker
 */

type CheckRequest struct {
	URLs []string `json:"urls"`
}

type Handler struct {
	client pb.HealthCheckerClient
}

func NewHandler(conn *grpc.ClientConn) *Handler {
	client := pb.NewHealthCheckerClient(conn)
	return &Handler{client: client}
}
func (h *Handler) CheckURLs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var req CheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Corps de la requête invalide", http.StatusBadRequest)
		return
	}

	// Timeout de 30 secondes pour la requête HTTP
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	resp, err := h.client.Check(ctx, &pb.HealthCheckRequest{Urls: req.URLs})
	if err != nil {
		http.Error(w, "Erreur lors de l'appel au service worker: "+err.Error(), http.StatusInternalServerError)
		log.Fatal("Erreur lors de l'appel au service worker: ", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp.Results); err != nil {
		http.Error(w, "Erreur lors de l'encodage de la réponse: "+err.Error(), http.StatusInternalServerError)
		log.Fatal("Erreur lors de l'encodage de la réponse: ", err)
		return
	}
	log.Println("Réponse envoyée avec succès")

}
