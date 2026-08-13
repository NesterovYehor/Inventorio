package handlers

import (
	"net/http"

	"github.com/NesterovYehor/Inventorio/internal/database"
	"github.com/NesterovYehor/Inventorio/ui"
)

// The Handler struct only holds what the web layer needs.
type Handler struct {
	db *database.DB
}

// New creates the handler and injects the dependencies.
func New(db *database.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// RegisterRoutes sets up the URLs and attaches them to the Handler's methods.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.handleIndex)
	mux.HandleFunc("GET /properties", h.handlerProperties)
	mux.Handle("GET /static/", ui.StaticHandler())

	return mux
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/properties", http.StatusSeeOther)
}

func (h *Handler) handlerProperties(w http.ResponseWriter, r *http.Request) {
	ui.Render(w, r, "properties", nil)

}
