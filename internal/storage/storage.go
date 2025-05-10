package storage

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/MosinFAM/load-balancer/internal/model"
	_ "github.com/lib/pq"
)

type ClientLimit struct {
	Capacity   int
	RefillRate int
}

type Store struct {
	DB *sql.DB
}

func NewStore(dsn string) (*Store, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
		return nil, fmt.Errorf("failed to open DB: %w", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Could not ping the database: %v", err)
		return nil, fmt.Errorf("failed to ping DB: %w", err)
	}
	return &Store{DB: db}, nil
}

func (s *Store) GetClientLimit(key string) (*model.ClientLimit, error) {
	var cl model.ClientLimit
	err := s.DB.QueryRow(`
		SELECT capacity, refill_rate
		FROM rate_limits
		WHERE client_key = $1
	`, key).Scan(&cl.Capacity, &cl.RefillRate)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query error: %w", err)
	}
	return &cl, nil
}
