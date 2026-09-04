package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/NesterovYehor/Inventorio/internal/models"
)

func (h *Handler) HandelNewOrder(w http.ResponseWriter, r *http.Request) {
	id, err := h.db.CreateNewOrder(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create new order: %v", err), http.StatusInternalServerError)
		log.Println(err)
		return
	}

	order, err := h.getOrderData(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get order data: %v", err), http.StatusInternalServerError)
		log.Println(err)

	}

	h.render.Content(w, r, "draft_order", order)
}

func (h *Handler) HandleOrdersList(w http.ResponseWriter, r *http.Request) {
	orders, err := h.db.GetAllOrders(r.Context())
	if err != nil {
		http.Error(w, "Failed to get all orders", http.StatusInternalServerError)
		log.Printf("Failed to get all orders %v", err)
		return
	}

	h.render.Content(w, r, "orders", orders)
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
	h.render.Component(w, r, "calculator_row", row)
}

func (h *Handler) HandleRemoveOrderProperty(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.FormValue("id"))
	orderID, _ := strconv.Atoi(r.FormValue("order_id"))

	if err := h.db.DeleteOrderProperty(r.Context(), id); err != nil {
		http.Error(w, "failed to delete propety from order", http.StatusInternalServerError)
		log.Printf("failed to delete propety from order: %v", err)
		return
	}
	rows, err := h.db.GetOrderRequirements(r.Context(), orderID)
	if err != nil {
		http.Error(w, "failed to get updated req rows", http.StatusInternalServerError)
		log.Printf("failed to get updated req rows: %v", err)
		return
	}
	h.render.Component(w, r, "calculator_tbody_oob", rows)
}

func (h *Handler) HandleCofirmOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	if err := h.db.UpdateOrderStatus(r.Context(), id); err != nil {
		http.Error(w, "failed to confirm order", http.StatusInternalServerError)
		log.Printf("failed to confirm order: %v", err)
		return
	}
}

func (h *Handler) HandleOrder(w http.ResponseWriter, r *http.Request) {
	isDraft := r.FormValue("draft")
	id, _ := strconv.Atoi(r.PathValue("id"))

	order, err := h.getOrderData(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get order", http.StatusInternalServerError)
		log.Printf("failed to get order: %v", err)
		return
	}

	if isDraft == "true" {
		h.render.Content(w, r, "draft_order", order)
		return
	}

	h.render.Content(w, r, "confirmed_order", order)
}

// This is a helper on the Handler, not the DB!
func (h *Handler) getOrderData(ctx context.Context, orderID int) (models.OrderView, error) {
	rows, err := h.db.GetOrderRequirements(ctx, orderID)
	if err != nil {
		return models.OrderView{}, fmt.Errorf("failed to get requirements: %w", err)
	}

	ps, err := h.db.GetAllProperties(ctx)
	if err != nil {
		return models.OrderView{}, fmt.Errorf("failed to list properties: %w", err)
	}

	selectedProp, err := h.db.GetSelectedProperties(ctx, orderID)
	if err != nil {
		return models.OrderView{}, fmt.Errorf("failed to list selected properties: %w", err)
	}

	return models.OrderView{
		ID:                 orderID,
		AllProperties:      ps,
		SelectedProperties: selectedProp,
		Rows:               rows,
	}, nil
}
