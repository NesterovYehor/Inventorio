package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/NesterovYehor/Inventorio/internal/models"
)

func (h *Handler) handleProperties(w http.ResponseWriter, r *http.Request) {
	names, err := h.db.GetAllItemsNames(r.Context())
	if err != nil {
		s := fmt.Sprintf("Failed to add new property:%v", err)
		http.Error(w, s, http.StatusInternalServerError)
		log.Println(s)
	}
	header := models.Header{}
	header.Names = names
	rows, err := h.db.GetAllPropertyRows(r.Context())
	if err != nil {
		s := fmt.Sprintf("Failed to add new property:%v", err)
		http.Error(w, s, http.StatusInternalServerError)
		log.Println(s)
		return
	}
	pp := models.PropertiesPage{
		Header: header,
		Rows:   rows,
	}
	h.render.Content(w, r, "properties", pp)
}

func (h *Handler) handleAddProperty(w http.ResponseWriter, r *http.Request) {
	id, err := h.db.AddProperty(r.Context())
	if err != nil {
		s := fmt.Sprintf("Failed to add new property:%v", err)
		http.Error(w, s, http.StatusInternalServerError)
		log.Println(s)
		return
	}
	pn, err := h.db.GetNeedsByPropertyID(r.Context(), id)
	if err != nil {
		s := fmt.Sprintf("Failed to add new property:%v", err)
		http.Error(w, s, http.StatusInternalServerError)
		log.Println(s)
		return
	}
	row := models.DefaultPropertyRow(id, pn)
	h.render.Component(w, r, "property-row", row)
}

func (h *Handler) handleUpdatePropertyName(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 0, 64)
	if err != nil {
		http.Error(w, "No id fond in url", http.StatusBadRequest)
		log.Printf("No id fond in url")
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse url:%v", err), http.StatusBadRequest)
		log.Printf("Failed to parse url:%v", err)
		return
	}

	if name := r.FormValue("name"); name != "" {
		if err := h.db.UpdatePropertyNameById(r.Context(), id, name); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update property name: %v", err), http.StatusInternalServerError)
			log.Printf("Failed to update property name: %v", err)
			return
		}
	}
}

func (h *Handler) handleUpdatePropertyNeed(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 0, 64)
	if err != nil {
		http.Error(w, "No id fond in url", http.StatusBadRequest)
		return
	}
	itemId, err := strconv.ParseInt(r.PathValue("itemId"), 0, 64)
	if err != nil {
		http.Error(w, "No id fond in url", http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("Failed to parse url:%v", err), http.StatusBadRequest)
		return
	}

	if q := r.FormValue("quantity"); q != "" {
		newVal, _ := strconv.Atoi(q)
		if err := h.db.UpdatePropertyNeedById(r.Context(), id, itemId, newVal); err != nil {
			http.Error(w, fmt.Sprintf("Failed to update property need: %v", err), http.StatusInternalServerError)
			return
		}
	}
}
