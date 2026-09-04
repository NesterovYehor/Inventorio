package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/NesterovYehor/Inventorio/internal/models"
)

func (h *Handler) handeleAddArrival(w http.ResponseWriter, r *http.Request) {
	propID, _ := strconv.Atoi(r.FormValue("property_id"))
	date := r.FormValue("arrival_date")
	id, err := h.db.CreateArrival(r.Context(), propID, date)
	if err != nil {
		http.Error(w, "Failed to create new arrival", http.StatusInternalServerError)
		log.Printf("Failed to create new arrival: %v", err)
		return
	}

	arrival, err := h.db.GetArrivalByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get arrival data", http.StatusInternalServerError)
		log.Printf("Failed to get arrival data: %v", err)
		return

	}

	h.render.Component(w, r, "arrival_row", arrival)
}

func (h *Handler) handleNewArrivalMenu(w http.ResponseWriter, r *http.Request) {
	properties, err := h.db.GetAllProperties(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get properties data: %v", err), http.StatusInternalServerError)
		log.Println(err)
	}
	h.render.Component(w, r, "arrival_modal", models.ArrivalModal{
		Properties: properties,
	})
}

func (h *Handler) handleArrivals(w http.ResponseWriter, r *http.Request) {
	arrivals, err := h.db.GetAllArrivals(r.Context())
	if err != nil {
		http.Error(w, "Failed to get arrivals", http.StatusInternalServerError)
		log.Printf("Failed to get arrivals: %v", err)
		return
	}

	h.render.Content(w, r, "arrivals", arrivals)
}

func (h *Handler) handleUpdateArrivalStatus(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	status := r.FormValue("status")

	if err := h.db.UpdateArrivalStatus(r.Context(), id, status); err != nil {
		http.Error(w, "Failed to update arrival status", http.StatusInternalServerError)
		log.Printf("Failed to update arrival status: %v", err)
		return
	}
	arrival, err := h.db.GetArrivalByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get arrival data", http.StatusInternalServerError)
		log.Printf("Failed to get arrival data: %v", err)
		return
	}

	h.render.Component(w, r, "arrival_row", arrival)
}

func (h *Handler) handleArrivalUpdateMenu(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	arrival, err := h.db.GetArrivalByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get arrival data", http.StatusInternalServerError)
		log.Printf("Failed to get arrival data: %v", err)
		return
	}
	properties, err := h.db.GetAllProperties(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get properties data: %v", err), http.StatusInternalServerError)
		log.Println(err)
	}
	h.render.Component(w, r, "arrival_modal", models.ArrivalModal{
		Arrival:    arrival,
		Properties: properties,
	})
}

func (h *Handler) handleUpdateArrivalDetails(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	propID, _ := strconv.Atoi(r.FormValue("property_id"))
	date := r.FormValue("arrival_date")

	if err := h.db.UpdateArrivalDetails(r.Context(), id, propID, date); err != nil {
		http.Error(w, "Failed to update arrival details", http.StatusInternalServerError)
		log.Printf("Failed to update arrival details: %v", err)
		return
	}

	arrival, err := h.db.GetArrivalByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to get arrival data", http.StatusInternalServerError)
		log.Printf("Failed to get arrival data: %v", err)
		return
	}
	h.render.Component(w, r, "arrival_row", arrival)
}

func (h *Handler) handleDeleteArrival(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if err := h.db.DeleteArrivalByID(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete arrival", http.StatusInternalServerError)
		log.Printf("Failed to delete arrival: %v", err)
		return
	}
}
