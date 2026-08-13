package models

// Property represents an apartment (e.g., "CARRIÓ")
type Property struct {
	ID   int
	Name string
}

// Item represents a physical thing in your central storage (e.g., "TOALLA grande")
type Item struct {
	ID      int
	Name    string
	InStock int // How many you currently have in the central warehouse
}

// PropertyNeed maps how many of a specific item a specific property requires
type PropertyNeed struct {
	PropertyID int
	ItemID     int
	Needed     int // E.g., This property needs 6 big towels
}

// ---------------------------------------------------------
// VIEW MODELS (Used for HTMX templates, not stored directly in DB)
// ---------------------------------------------------------

// ShortfallRow represents one row in the HTML calculator table
type ShortfallRow struct {
	ItemName  string
	Needed    int
	InStock   int
	Shortfall int // (Needed - InStock). If InStock >= Needed, this is 0.
}
