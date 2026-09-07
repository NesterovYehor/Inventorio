package database

import (
	"context"
	"fmt"
	"time"

	"github.com/NesterovYehor/Inventorio/internal/models"
)

func (db *DB) GetItemByID(ctx context.Context, id int) (models.Item, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)

	defer cancel()
	var item models.Item

	err := db.conn.QueryRowContext(ctx, `SELECT * FROM items ;`).Scan(&item.ID, &item.Name, &item.Quantity)
	return item, err
}

func (db *DB) UpdateAllItems(ctx context.Context, rows []models.CalculatorRow) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	tx, err := db.conn.BeginTx(ctx, nil)
	defer tx.Rollback()
	if err != nil {
		return fmt.Errorf("Failed to beguin transaction")
	}
	query := `
	UPDATE items
	SET quantity = ?
	WHERE id = ?;
	`

	for _, row := range rows {
		tx.ExecContext(ctx, query, row.OrderQty+row.Have, row.Item.ID)
	}

	return tx.Commit()

}

func (db *DB) CreateNewItem(ctx context.Context) (models.Item, error) {
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
