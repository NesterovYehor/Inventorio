package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"time"

	"github.com/NesterovYehor/Inventorio/internal/models"
	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schemaSQL string

type DB struct {
	conn *sql.DB
}

func Connect(path string) (*DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to database: %w", err)
	}
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("Failed to ping database: %w", err)
	}
	db := &DB{conn: conn}
	if err := db.initSchema(context.Background()); err != nil {
		return nil, fmt.Errorf("Failed to create schema: %w", err)
	}
	return db, nil
}

func (db *DB) initSchema(ctx context.Context) error {
	if _, err := db.conn.ExecContext(ctx, schemaSQL); err != nil {
		return err
	}
	return nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) AddProperty(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)

	defer cancel()

	result, err := db.conn.ExecContext(ctx, `INSERT INTO properties DEFAULT VALUES;`)
	if err != nil {
		return 0, fmt.Errorf("Failed to add new property: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Failed to get last property id: %w", err)
	}

	query := `
	INSERT INTO property_needs (property_id, item_id, quantity)
	SELECT ?, id, 0
	FROM items
	`
	if _, err := db.conn.ExecContext(ctx, query, id); err != nil {
		return 0, fmt.Errorf("Failed to add property needs: %w", err)
	}

	return id, nil
}

func (db *DB) GetAllProperties(ctx context.Context) ([]models.Property, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, "SELECT * FROM properties")
	if err != nil {
		return nil, err
	}

	ps := []models.Property{}
	for rows.Next() {
		p := models.Property{}
		rows.Scan(
			&p.ID,
			&p.Name,
		)
		ps = append(ps, p)
	}
	if rows.Err() != nil {
		return nil, err
	}
	return ps, nil
}

func (db *DB) GetAllPropertyRows(ctx context.Context) ([]models.PropertyRow, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)

	defer cancel()

	query := `
	SELECT *
	FROM properties as p
	LEFT  OUTER JOIN property_needs as pn
	ON pn.property_id = p.id
	`

	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	prs := []models.PropertyRow{}
	prsMap := make(map[int64]int)
	prsMapBool := make(map[int64]bool)

	for rows.Next() {
		var pr models.PropertyRow
		var pn models.PropertyNeed
		rows.Scan(
			&pr.Property.ID,
			&pr.Property.Name,
			&pn.PropertyID,
			&pn.ItemID,
			&pn.Quantity,
		)
		if prsMapBool[pr.Property.ID] {
			targetIdx := prsMap[pr.Property.ID]
			prs[targetIdx].PropertyNeed = append(prs[targetIdx].PropertyNeed, pn)
		} else {
			pr.PropertyNeed = append(pr.PropertyNeed, pn)
			prsMapBool[pr.Property.ID] = true
			prsMap[pr.Property.ID] = len(prs)
			prs = append(prs, pr)
		}
	}
	if rows.Err() != nil {
		return nil, err
	}

	return prs, nil
}

func (db *DB) UpdatePropertyNameById(ctx context.Context, id int64, name string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)

	defer cancel()

	query := `
	UPDATE properties 
	SET name = ?
	WHERE id = ?
	`
	result, err := db.conn.ExecContext(ctx, query, name, id)
	if err != nil {
		return fmt.Errorf("Failed to update property: %w", err)
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Error durring update property: %w", err)
	}

	if rowAffected == 0 {
		return fmt.Errorf("Property not found")
	}
	return nil
}

func (db *DB) UpdatePropertyNeedById(ctx context.Context, propId, itemId int64, newVal int) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)

	defer cancel()

	query := `
	UPDATE property_needs
	SET quantity = ?
	WHERE property_id = ? and item_id = ?
	`
	result, err := db.conn.ExecContext(ctx, query, newVal, propId, itemId)
	if err != nil {
		return fmt.Errorf("Failed to update property's need: %w", err)
	}

	rowAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("Error durring update property's need: %w", err)
	}

	if rowAffected == 0 {
		return fmt.Errorf("Property's need not found")
	}
	return nil

}

func (db *DB) GetNeedsByPropertyID(ctx context.Context, propertyID int64) ([]models.PropertyNeed, error) {
	query := `
		SELECT property_id, item_id, quantity 
		FROM property_needs 
		WHERE property_id = ?
		ORDER BY item_id
		;
	`
	rows, err := db.conn.QueryContext(ctx, query, propertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to query property needs: %w", err)
	}
	defer rows.Close()

	var needs []models.PropertyNeed
	for rows.Next() {
		var need models.PropertyNeed
		if err := rows.Scan(&need.PropertyID, &need.ItemID, &need.Quantity); err != nil {
			return nil, fmt.Errorf("failed to scan property need: %w", err)
		}
		needs = append(needs, need)
	}

	return needs, rows.Err()
}

func (db *DB) AddNewItem(ctx context.Context) (models.Item, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)

	defer cancel()
	var item models.Item

	var id int
	err := db.conn.QueryRowContext(ctx, `INSERT INTO items DEFAULT VALUES RETURNING id;`).Scan(&id)
	if err != nil {
		return item, fmt.Errorf("Failed to add new item: %w", err)
	}

	query := `
	INSERT INTO property_needs (property_id, item_id, quantity)
	SELECT id, ?, 0
	FROM properties
	`
	if _, err := db.conn.ExecContext(ctx, query, id); err != nil {
		return item, fmt.Errorf("Failed to add property needs: %w", err)
	}
	item.ID = int(id)
	item.Quantity = 0
	item.Name = ""

	return item, nil
}

func (db *DB) UpdateItemField(ctx context.Context, column string, value any, id int64) error {
	allowedColumns := map[string]bool{
		"name":     true,
		"quantity": true,
	}

	if !allowedColumns[column] {
		return fmt.Errorf("invalid column name: %s", column)
	}

	query := fmt.Sprintf("UPDATE items SET %s = ? WHERE id = ?;", column)
	_, err := db.conn.ExecContext(ctx, query, value, id)
	return err
}

func (db *DB) GetAllItems(ctx context.Context) ([]models.Item, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, "SELECT id, name, quantity FROM items ORDER BY id")

	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Quantity); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %w", err)
	}

	return items, nil
}

func (db *DB) GetAllItemsNames(ctx context.Context) ([]models.ItemName, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	rows, err := db.conn.QueryContext(ctx, "SELECT id, name FROM items ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var names []models.ItemName
	for rows.Next() {
		var name models.ItemName
		if err := rows.Scan(&name.ID, &name.Name); err != nil {
			return nil, fmt.Errorf("failed to scan item's name: %w", err)
		}
		names = append(names, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating items: %w", err)
	}

	return names, nil
}

func (db *DB) DeleteItemById(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()
	query := "DELETE FROM items WHERE id = ?;"
	result, err := db.conn.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete item with id %d: %w", id, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected after deleting item with id %d: %w", id, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("item with id %d not found", id)
	}

	return nil
}

func (db *DB) AddPropertyToOrder(ctx context.Context, prName, checkInDate string) (models.OrderPropertyRow, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	query := `
	INSERT INTO order_properties (order_id, property_id, arrival_date)
	SELECT orders.id, properties.id, ?
	FROM orders
	CROSS JOIN properties
	WHERE orders.is_draft = true 
	  AND properties.name = ?
	RETURNING id, property_id;
`

	var row models.OrderPropertyRow
	err := db.conn.QueryRowContext(ctx, query, checkInDate, prName).Scan(&row.ID, &row.PropertyID)
	if err != nil {
		return row, err
	}
	return row, nil
}

func (db *DB) GetOrderRequirements(ctx context.Context, orderID int) ([]models.CalculatorRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Short, clean query using subqueries
	query := `
		SELECT 
			i.id,
			i.name,
			COALESCE((
				SELECT SUM(pn.quantity) 
				FROM property_needs pn
				JOIN order_properties op ON op.property_id = pn.property_id
				WHERE op.order_id = ? AND pn.item_id = i.id
			), 0) AS need_qty,
			COALESCE((
				SELECT ol.extra_qty 
				FROM order_lines ol 
				WHERE ol.order_id = ? AND ol.item_id = i.id
			), 0) AS extra_qty,
			i.quantity AS have_stock
		FROM items i;
	`

	rows, err := db.conn.QueryContext(ctx, query, orderID, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query requirements: %w", err)
	}
	defer rows.Close()

	var results []models.CalculatorRow
	for rows.Next() {
		var r models.CalculatorRow
		if err := rows.Scan(&r.Item.ID, &r.Item.Name, &r.Need, &r.Extra, &r.Have); err != nil {
			return nil, fmt.Errorf("failed to scan calculator row: %w", err)
		}

		// Handle the shortfall & order math cleanly in Go
		if gap := r.Need - r.Have; gap > 0 {
			r.Gap = gap
		}
		if qty := (r.Need + r.Extra) - r.Have; qty > 0 {
			r.OrderQty = qty
		}

		results = append(results, r)
	}

	return results, rows.Err()
}
func (db *DB) AddNewOrder(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	query := `
		INSERT INTO orders (is_draft) VALUES (true)
		ON CONFLICT (is_draft) WHERE is_draft = true
		DO UPDATE SET is_draft = true
		RETURNING id;
	`

	var id int
	err := db.conn.QueryRowContext(ctx, query).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (db *DB) GetSelectedProperties(ctx context.Context, orderID int) ([]models.OrderPropertyRow, error) {
	query := `
	SELECT p.id, p.name, op.id, COALESCE(op.arrival_date, '') 
		FROM properties p
		INNER JOIN order_properties op ON p.id = op.property_id
		WHERE op.order_id = ?;
	`
	rows, err := db.conn.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var props []models.OrderPropertyRow
	for rows.Next() {
		var p models.OrderPropertyRow
		if err := rows.Scan(&p.PropertyID, &p.Name, &p.ID, &p.ArrivalDate); err != nil {
			return nil, err
		}
		props = append(props, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return props, nil
}

func (db *DB) UpdateOrderExtraValue(ctx context.Context, orderID, itemID, value int) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	// MUST BE AN INSERT ... ON CONFLICT (UPSERT)
	query := `
		INSERT INTO order_lines (order_id, item_id, extra_qty) 
		VALUES (?, ?, ?)
		ON CONFLICT (order_id, item_id) 
		DO UPDATE SET extra_qty = excluded.extra_qty;
	`

	if _, err := db.conn.ExecContext(ctx, query, orderID, itemID, value); err != nil {
		return fmt.Errorf("failed to upsert order extra value: %w", err)
	}

	return nil
}

func (db *DB) GetOrderItemRequirement(ctx context.Context, orderID, itemID int) (models.CalculatorRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Added WHERE i.id = ? at the very end
	query := `
		SELECT 
			i.id,
			i.name,
			COALESCE((
				SELECT SUM(pn.quantity) 
				FROM property_needs pn
				JOIN order_properties op ON op.property_id = pn.property_id
				WHERE op.order_id = ? AND pn.item_id = i.id
			), 0) AS need_qty,
			COALESCE((
				SELECT ol.extra_qty 
				FROM order_lines ol 
				WHERE ol.order_id = ? AND ol.item_id = i.id
			), 0) AS extra_qty,
			i.quantity AS have_stock
		FROM items i
		WHERE i.id = ?;
	`

	var r models.CalculatorRow

	if err := db.conn.QueryRowContext(ctx, query, orderID, orderID, itemID).Scan(
		&r.Item.ID,
		&r.Item.Name,
		&r.Need,
		&r.Extra,
		&r.Have,
	); err != nil {
		return r, fmt.Errorf("failed to query single item requirement: %w", err)
	}

	if gap := r.Need - r.Have; gap > 0 {
		r.Gap = gap
	}
	if qty := (r.Need + r.Extra) - r.Have; qty > 0 {
		r.OrderQty = qty
	}

	return r, nil
}

func (db *DB) DeleteOrderProperty(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
	DELETE FROM order_properties
		WHERE id = ?;
	`
	if _, err := db.conn.ExecContext(ctx, query, id); err != nil {
		return err
	}
	return nil
}
