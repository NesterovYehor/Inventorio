package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
)

func (h *Handler) handleItems(w http.ResponseWriter, r *http.Request) {
	items, err := h.db.GetAllItems(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get items:%v", err), http.StatusInternalServerError)
		log.Printf("Failed to get items:%v", err)
		return
	}
	h.render.Content(w, r, "storage", items)
}

func (h *Handler) handleAddItem(w http.ResponseWriter, r *http.Request) {
	item, err := h.db.CreateNewItem(r.Context())
	if err != nil {
		s := fmt.Sprintf("Failed to add new item:%v", err)
		http.Error(w, s, http.StatusInternalServerError)
		log.Println(s)
		return
	}
	h.render.Component(w, r, "item-row", item)
}

func (h *Handler) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 0, 64)
	if err != nil {
		http.Error(w, "No id fond in url", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse url:%v", err), http.StatusBadRequest)
		return
	}

	if name := r.FormValue("name"); name != "" {
		err = h.db.UpdateItemField(r.Context(), "name", name, id)
	}
	if q := r.FormValue("quantity"); q != "" {
		quantity, _ := strconv.Atoi(q)
		err = h.db.UpdateItemField(r.Context(), "quantity", quantity, id)
	}
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update item: %v", err), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 0, 64)
	if err != nil {
		http.Error(w, "No id fond in url", http.StatusBadRequest)
		log.Println("No id fond in url")
		return
	}
	if err := h.db.DeleteItemById(r.Context(), id); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete item: %v", err), http.StatusBadRequest)
		log.Println(err)
		return
	}
}
