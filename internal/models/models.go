package models

// Property represents an apartment (e.g., "CARRIÓ")
type Property struct {
	ID   int64
	Name string
}

// Item represents a physical thing in your central storage (e.g., "TOALLA grande")
type Item struct {
	ID       int
	Name     string
	Quantity int // How many you currently have in the central warehouse
}

// ItemName represents a name of a physical thing
// in your central storage (e.g., "TOALLA grande") with out Quantity value
type ItemName struct {
	ID   int
	Name string
}

// PropertyNeed maps how many of a specific item a specific property requires
type PropertyNeed struct {
	PropertyID int
	ItemID     int
	Quantity   int // E.g., This property needs 6 big towels
}

// ---------------------------------------------------------
// VIEW MODELS (Used for HTMX templates, not stored directly in DB)
// ---------------------------------------------------------

type PropertyRow struct {
	Property     Property
	PropertyNeed []PropertyNeed
}

type Header struct {
	Names []ItemName
}

type PropertiesPage struct {
	Header Header
	Rows   []PropertyRow
}

func DefaulPropertyRow(id int64, pn []PropertyNeed) *PropertyRow {
	return &PropertyRow{
		Property: Property{
			ID:   id,
			Name: "Unnamed Property",
		},
		PropertyNeed: pn,
	}
}
