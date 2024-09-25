package httpServer

import (
	"encoding/json"
	"net/http"
	"strconv"

	"AIChallenge/internal/usecase/news"
	"github.com/gorilla/mux"
)

type HTTPHandler struct {
	UseCase *usecase.NewsUseCase
}

func NewHTTPHandler(useCase *usecase.NewsUseCase) *HTTPHandler {
	return &HTTPHandler{
		UseCase: useCase,
	}
}

func (h *HTTPHandler) GetLatestNewsHandler(w http.ResponseWriter, r *http.Request) {

	k, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil || k <= 0 {
		k = 10
	}

	newsList, err := h.UseCase.GetLatestNews(k)
	if err != nil {
		http.Error(w, "Failed to get news", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err = json.NewEncoder(w).Encode(newsList)
	if err != nil {
		http.Error(w, "Failed to encode news to JSON", http.StatusInternalServerError)
		return
	}

}

func (h *HTTPHandler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/news", http.StatusSeeOther)
}

func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/", h.HomeHandler).Methods("GET")

	router.HandleFunc("/news", h.GetLatestNewsHandler).Methods("GET")
}
