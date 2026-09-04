package database

import (
	"context"
	"fmt"
	"time"

	"github.com/NesterovYehor/Inventorio/internal/models"
)

func (db *DB) CreateProperty(ctx context.Context) (int64, error) {
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
