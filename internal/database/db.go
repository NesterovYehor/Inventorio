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
		WHERE property_id = ?;
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

	result, err := db.conn.ExecContext(ctx, `INSERT INTO items DEFAULT VALUES;`)
	if err != nil {
		return item, fmt.Errorf("Failed to add new item: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return item, fmt.Errorf("Failed to get last item id: %w", err)
	}

	query := `
	INSERT INTO property_needs (property_id, item_id, quantity)
	SELECT ?, id, 0
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

	rows, err := db.conn.QueryContext(ctx, "SELECT id, name, quantity FROM items")
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

	rows, err := db.conn.QueryContext(ctx, "SELECT id, name FROM items")
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
	query = "DELETE FROM property_needs WHERE item_id = ?;"
	result, err = db.conn.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete item with id %d: %w", id, err)
	}

	return nil
}
