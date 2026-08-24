package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/NesterovYehor/Inventorio/internal/models"
	"github.com/NesterovYehor/Inventorio/ui"
)

func (h *Handler) HandelNewOrder(w http.ResponseWriter, r *http.Request) {
	id, err := h.db.AddNewOrder(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create new order: %v", err), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	rows, err := h.db.GetOrderRequirements(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get new order: %v", err), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	ps, err := h.db.GetAllProperties(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list of all properties %v", err), http.StatusInternalServerError)
		log.Printf("Failed to list of all properties %v \n", err)
		return
	}
	selectedProp, err := h.db.GetSelectedProperties(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list of selected properties %v", err), http.StatusInternalServerError)
		log.Printf("Failed to list of selected properties %v", err)
		return
	}
	log.Println(id)
	ui.RenderContent(w, r, "calculator", models.Calculator{
		ID:                 id,
		AllProperties:      ps,
		SelectedProperties: selectedProp,
		Rows:               rows,
	})
}

func (h *Handler) HandleAddPropertyToOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	name := r.FormValue("property_name")
	dateStr := r.FormValue("property_date")

	var date *time.Time
	if dateStr != "" {
		parsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "Invalid date format", http.StatusBadRequest)
			log.Println("Invalid date format")
			return
		}

		date = &parsed
	}
	if err := h.db.AddPropertyToOrder(r.Context(), name, date); err != nil {
		http.Error(w, "faild to add property in to order", http.StatusInternalServerError)
		log.Printf("faild to add property in to order: %v/n", err)
		return
	}

	rows, err := h.db.GetOrderRequirements(r.Context(), id)
	if err != nil {
		http.Error(w, "faild to get calculated data", http.StatusInternalServerError)
		log.Printf("faild to get calculated data: %v/n", err)
		return
	}
	ui.RenderComponent(w, "property_item", models.Property{Name: name})

	ui.RenderComponent(w, "calculator_tbody_oob", rows)
}

func (h *Handler) HandleUpateExtraValue(w http.ResponseWriter, r *http.Request) {
	itemID, _ := strconv.Atoi(r.PathValue("id"))

	value, _ := strconv.Atoi(r.FormValue("extra"))
	orderID, _ := strconv.Atoi(r.FormValue("id"))

	if err := h.db.UpdateOrderExtraValue(r.Context(), orderID, itemID, value); err != nil {
		http.Error(w, "failed to update extra value", http.StatusInternalServerError)
		log.Printf("failed to update extra value: %v/n", err)
		return
	}
	row, err := h.db.GetOrderItemRequirement(r.Context(), orderID, itemID)
	if err != nil {
		http.Error(w, "failed to get updated row", http.StatusInternalServerError)
		log.Printf("failed to get updated row: %v", err)
		return
	}
	ui.RenderComponent(w, "calculator_row", row)
}

func (h *Handler) HandleRemoveOrderProperty(w http.ResponseWriter, r *http.Request) {
}
