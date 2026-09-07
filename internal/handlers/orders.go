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
	order, err := h.getOrderData(r.Context(), id, start, end)
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
	row, err := h.calculateItemRow(r.Context(), orderID, itemID)
	if err != nil {
		http.Error(w, "failed to get updated row", http.StatusInternalServerError)
		log.Printf("failed to get updated row: %v", err)
		return
	}
	h.render.Component(w, r, "calculator_row", row)
}

func (h *Handler) HandleCofirmOrder(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	start, end, err := h.db.GetOrderRange(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get order date range", http.StatusInternalServerError)
		log.Printf("failed to get order date range: %v", err)
		return
	}
	order, err := h.getOrderData(r.Context(), id, start, end)
	if err != nil {
		http.Error(w, "failed to get order data", http.StatusInternalServerError)
		log.Printf("failed to get order data: %v", err)
		return
	}

	if err := h.db.UpdateAllItems(r.Context(), order.Rows); err != nil {
		http.Error(w, "failed to update items", http.StatusInternalServerError)
		log.Printf("failed to update items: %v", err)
		return
	}

	if err := h.db.UpdateOrderStatus(r.Context(), id); err != nil {
		http.Error(w, "failed to confirm order", http.StatusInternalServerError)
		log.Printf("failed to confirm order: %v", err)
		return
	}
	log.Println(h.db.GetOrderRange(r.Context(), id))
}

func (h *Handler) HandleOrder(w http.ResponseWriter, r *http.Request) {
	isDraft := r.FormValue("draft")
	id, _ := strconv.Atoi(r.PathValue("id"))
	log.Println(id)

	start, end, err := h.db.GetOrderRange(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to get order range", http.StatusInternalServerError)
		log.Printf("failed to get order range: %v", err)
		return
	}

	order, err := h.getOrderData(r.Context(), id, start, end)
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
func (h *Handler) getOrderData(ctx context.Context, id int, start, end string) (models.OrderView, error) {

	items, err := h.db.GetAllItems(ctx)
	if err != nil {
		return models.OrderView{}, fmt.Errorf("failed to get items list: %w", err)
	}

	extras, err := h.db.GetOrderExtras(ctx, id)
	if err != nil {
		return models.OrderView{}, fmt.Errorf("failed to get extras values of items: %w", err)
	}

	needs, err := h.db.GetItemNeedsInDateRange(ctx, start, end)
	if err != nil {
		return models.OrderView{}, fmt.Errorf("failed to get needs values of order: %w", err)
	}

	rows := []models.CalculatorRow{}
	for _, item := range items {
		need := needs[item.ID]
		have := item.Quantity
		gap := max(need-have, 0)
		extr := extras[item.ID]
		total := max((need+extr)-have, 0)
		row := models.CalculatorRow{
			Item: models.ItemName{
				ID:   item.ID,
				Name: item.Name,
			},
			Need:     need,
			Have:     have,
			Gap:      gap,
			Extra:    extr,
			OrderQty: total,
		}
		rows = append(rows, row)
	}

	return models.OrderView{
		ID:        id,
		Rows:      rows,
		StartDate: start,
		EndDate:   end,
	}, nil
}

func (h *Handler) calculateItemRow(ctx context.Context, orderID, itemID int) (models.CalculatorRow, error) {
	start, end, err := h.db.GetOrderRange(ctx, orderID)
	if err != nil {
		return models.CalculatorRow{}, fmt.Errorf("failed to get order range dates: %w", err)
	}
	item, err := h.db.GetItemByID(ctx, itemID)
	if err != nil {
		return models.CalculatorRow{}, fmt.Errorf("failed to get item: %w", err)
	}

	extra, err := h.db.GetSingleItemExtra(ctx, orderID, itemID)
	if err != nil {
		return models.CalculatorRow{}, fmt.Errorf("failed to get extra: %w", err)
	}

	var need int
	if start != "" && end != "" {
		need, err = h.db.GetSingleItemNeed(ctx, itemID, start, end)
		if err != nil {
			return models.CalculatorRow{}, fmt.Errorf("failed to get need: %w", err)
		}
	}

	have := item.Quantity
	gap := max(need-have, 0)
	total := max((need+extra)-have, 0)

	return models.CalculatorRow{
		Item: models.ItemName{
			ID:   item.ID,
			Name: item.Name,
		},
		Need:     need,
		Have:     have,
		Gap:      gap,
		Extra:    extra,
		OrderQty: total,
	}, nil
}
