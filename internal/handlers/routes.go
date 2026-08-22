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
	mux.HandleFunc("GET /properties", h.handleProperties)
	mux.HandleFunc("POST /properties", h.handleAddProperty)
	mux.HandleFunc("PATCH /properties/{id}", h.handleUpdatePropertyName)
	mux.HandleFunc("PATCH /properties/{id}/items/{itemId}", h.handleUpdatePropertyNeed)
	mux.HandleFunc("GET /items", h.handleItems)
	mux.HandleFunc("POST /items", h.handleAddItem)
	mux.HandleFunc("PATCH /items/{id}", h.handleUpdateItem)
	mux.HandleFunc("DELETE /items/{id}", h.handleDeleteItem)
	mux.HandleFunc("GET /orders/new", h.HandelNewOrder)
	mux.HandleFunc("POST /orders/properties", h.HandleAddPropertyToOrder)
	mux.Handle("GET /static/", ui.StaticHandler())

	return mux
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/items", http.StatusSeeOther)
}
