package database

import (
	"context"
	"fmt"
	"time"

	"github.com/NesterovYehor/Inventorio/internal/models"
)

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
				JOIN arrivals op ON op.property_id = pn.property_id
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
func (db *DB) CreateNewOrder(ctx context.Context) (int, error) {
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
		INNER JOIN arrivals op ON p.id = op.property_id
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
				JOIN arrivals op ON op.property_id = pn.property_id
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
	DELETE FROM arrivals
		WHERE id = ?;
	`
	if _, err := db.conn.ExecContext(ctx, query, id); err != nil {
		return err
	}
	return nil
}

func (db *DB) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `
		SELECT * FROM orders
		ORDER BY 
    is_draft DESC,       
    confirm_date DESC;
	`
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	var orders []models.Order

	for rows.Next() {
		var order models.Order
		rows.Scan(
			&order.ID,
			&order.IsDraft,
			&order.ConfirmDate,
		)
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

func (db *DB) UpdateOrderStatus(ctx context.Context, orderID int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 1. Grab the current time and format it as an ISO8601/RFC3339 string
	now := time.Now().Format("2006-01-02")

	// 2. Update the draft status and set the date
	query := `
		UPDATE orders 
		SET is_draft = false, confirm_date = ? 
		WHERE id = ?;
	`

	result, err := db.conn.ExecContext(ctx, query, now, orderID)
	if err != nil {
		return fmt.Errorf("failed to confirm order: %w", err)
	}

	// 3. Optional: Verify the row actually existed
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no draft order found with id %d", orderID)
	}

	return nil
}
