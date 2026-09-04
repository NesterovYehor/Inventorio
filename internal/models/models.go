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

type Order struct {
	ID          int
	IsDraft     bool
	ConfirmDate string
}

type Arrival struct {
	ID          int
	Name        string
	ArrivalDate string
	Status      string
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

type CalculatorRow struct {
	Item     ItemName // Makes it cleaner: row.Item.ID and row.Item.Name
	Need     int
	Have     int
	Gap      int
	Extra    int
	OrderQty int // Much clearer than EndNumber
}

type OrderView struct {
	ID        int
	StartDate string
	EndDate   string
	Rows      []CalculatorRow
}

type ArrivalModal struct {
	Arrival    Arrival
	Properties []Property
}

func DefaultPropertyRow(id int64, pn []PropertyNeed) *PropertyRow {
	return &PropertyRow{
		Property: Property{
			ID:   id,
			Name: "Unnamed Property",
		},
		PropertyNeed: pn,
	}
}

func EmptyOrderView(items []Item, id int) OrderView {
	order := OrderView{
		ID:   id,
		Rows: []CalculatorRow{},
	}
	for _, i := range items {
		row := CalculatorRow{
			Item: ItemName{
				ID:   i.ID,
				Name: i.Name,
			},
			Have: i.Quantity,
		}
		order.Rows = append(order.Rows, row)
	}
	return order
}
