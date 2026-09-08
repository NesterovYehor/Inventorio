package database

import (
	"context"
	"fmt"
	"time"

	"github.com/NesterovYehor/Inventorio/internal/models"
)

func (db *DB) CreateArrival(ctx context.Context, propertyID int, checkInDate string) (int, error) {
	query := `INSERT INTO arrivals (property_id, arrival_date) VALUES (?, ?) RETURNING id;`

	var id int
	err := db.conn.QueryRowContext(ctx, query, propertyID, checkInDate).Scan(&id)
	return id, err
}

func (db *DB) GetArrivalByID(ctx context.Context, id int) (models.Arrival, error) {
	query := `
	SELECT arrivals.id, arrivals.arrival_date, arrivals.status, properties.name 
	FROM arrivals
	JOIN properties ON arrivals.property_id = properties.id
	WHERE arrivals.id = ?;`

	var arrival models.Arrival
	err := db.conn.QueryRowContext(ctx, query, id).Scan(
		&arrival.ID,
		&arrival.ArrivalDate,
		&arrival.Status,
		&arrival.Name,
	)
	return arrival, err
}

func (db *DB) GetAllArrivals(ctx context.Context) ([]models.Arrival, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	query := `
		SELECT 
    	arrivals.id, 
    	arrivals.arrival_date, 
    	arrivals.status, 
    	properties.name 
		FROM arrivals
		JOIN properties ON arrivals.property_id = properties.id
		ORDER BY arrivals.arrival_date ASC
`

	var arrivals []models.Arrival
	rows, err := db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var arrival models.Arrival
		rows.Scan(
			&arrival.ID,
			&arrival.ArrivalDate,
			&arrival.Status,
			&arrival.Name,
		)
		arrivals = append(arrivals, arrival)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return arrivals, nil
}

func (db *DB) UpdateArrivalStatus(ctx context.Context, id int, status string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	query := `
		UPDATE arrivals
    SET status = ?
		WHERE id = ?;
	`
	_, err := db.conn.ExecContext(ctx, query, status, id)
	return err
}

func (db *DB) UpdateArrivalDetails(ctx context.Context, id, propID int, date string) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	query := `
		UPDATE arrivals
    SET property_id = ?, arrival_date = ?
		WHERE id = ?;
	`
	_, err := db.conn.ExecContext(ctx, query, propID, date, id)
	return err
}

func (db *DB) DeleteArrivalByID(ctx context.Context, id int) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	query := `
		DELETE FROM arrivals WHERE id = ?;
	`
	_, err := db.conn.ExecContext(ctx, query, id)
	return err
}

func (db *DB) ApplyArrivalNeeds(ctx context.Context, arrivalID int, deduct bool) error {
	operator := "+"
	if deduct {
		operator = "-"
	}

	// while keeping the arrivalID parameterized for security.
	query := fmt.Sprintf(`
	UPDATE items 
	SET quantity = items.quantity %s pn.quantity
	FROM property_needs pn
	JOIN arrivals a ON a.property_id = pn.property_id
	WHERE items.id = pn.item_id 
	AND a.id = ?;
	`, operator)

	_, err := db.conn.ExecContext(ctx, query, arrivalID)
	return err
}
