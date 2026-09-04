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
	items, err := h.db.GetAllItems(r.Context())
	if err != nil {
		http.Error(w, "Failed to get items", http.StatusInternalServerError)
		log.Println(err)
		return

	}
	order := models.EmptyOrderView(items, id)

	h.render.Content(w, r, "draft_order", order)
}

func (h *Handler) handleSetDataRange(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))

	start := r.FormValue("start_data")
	end := r.FormValue("end_data")

	if err := h.db.SetOrderRange(r.Context(), id, start, end); err != nil {
		http.Error(w, "Failed to set date range for order", http.StatusInternalServerError)
		log.Printf("Failed to set date range for order %v", err)
		return
	}
	order, err := h.getOrderData(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get order data: %v", err), http.StatusInternalServerError)
		log.Println(err)
		return
	}
	h.render.Component(w, r, "calculator_update", order)
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
func (h *Handler) getOrderData(ctx context.Context, id int) (models.OrderView, error) {
	rows, err := h.db.GetOrderRequirements(ctx, id)
	if err != nil {
		return models.OrderView{}, fmt.Errorf("failed to get requirements: %w", err)
	}

	start, end, err := h.db.GetOrderRange(ctx, id)
	if err != nil {
		return models.OrderView{}, fmt.Errorf("failed to get date rangbe: %w", err)
	}

	return models.OrderView{
		ID:        id,
		Rows:      rows,
		StartDate: start,
		EndDate:   end,
	}, nil
}
