package handlers

import (
	"log"
	"net/http"

	"github.com/NesterovYehor/Inventorio/internal/database"
	"github.com/NesterovYehor/Inventorio/ui"
)

// The Handler struct only holds what the web layer needs.
type Handler struct {
	db     *database.DB
	render *ui.Renderer
}

// New creates the handler and injects the dependencies.
func New(db *database.DB, render *ui.Renderer) *Handler {
	return &Handler{
		db:     db,
		render: render,
	}
}

// RegisterRoutes sets up the URLs and attaches them to the Handler's methods.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", h.handleIndex)
	mux.HandleFunc("GET /change-language", h.handleChangeLanguage)
	mux.HandleFunc("GET /properties", h.handleProperties)
	mux.HandleFunc("POST /properties", h.handleAddProperty)
	mux.HandleFunc("PATCH /properties/{id}", h.handleUpdatePropertyName)
	mux.HandleFunc("PATCH /properties/{id}/items/{itemId}", h.handleUpdatePropertyNeed)
	mux.HandleFunc("GET /items", h.handleItems)
	mux.HandleFunc("POST /items", h.handleAddItem)
	mux.HandleFunc("PATCH /items/{id}", h.handleUpdateItem)
	mux.HandleFunc("DELETE /items/{id}", h.handleDeleteItem)
	mux.HandleFunc("GET /orders", h.HandleOrdersList)
	mux.HandleFunc("GET /orders/{id}", h.HandleOrder)
	mux.HandleFunc("GET /orders/new", h.HandelNewOrder)
	mux.HandleFunc("POST /orders/{id}", h.handleSetDataRange)
	mux.HandleFunc("POST /orders/confirm/{id}", h.HandleCofirmOrder)
	mux.HandleFunc("PATCH /orders/items/{id}", h.HandleUpateExtraValue)
	mux.HandleFunc("GET /arrivals", h.handleArrivals)
	mux.HandleFunc("GET /arrivals/new", h.handleNewArrivalMenu)
	mux.HandleFunc("POST /arrivals", h.handeleAddArrival)
	mux.HandleFunc("PATCH /arrivals/{id}/status", h.handleUpdateArrivalStatus)
	mux.HandleFunc("GET /arrivals/{id}/details/update", h.handleArrivalUpdateMenu)
	mux.HandleFunc("PATCH /arrivals/{id}/details", h.handleUpdateArrivalDetails)
	mux.HandleFunc("DELETE /arrivals/{id}", h.handleDeleteArrival)
	mux.Handle("GET /static/", ui.StaticHandler())

	return LanguageMiddleware(mux)
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/items", http.StatusSeeOther)
}

func (h *Handler) handleChangeLanguage(w http.ResponseWriter, r *http.Request) {
	lang := r.FormValue("lang")
	if lang == "" {
		http.Error(w, "No language passed", http.StatusBadRequest)
		log.Println("No language passed")
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "lang",
		Value:    lang,
		Path:     "/",
		MaxAge:   31536000,
		HttpOnly: true,
		Secure:   false,
	})
	w.Header().Add("HX-Refresh", "true")
}
