package db

import (
	"context"
	"log"
	"os"
	"time"
	"welcome-to-the-jungle/pkg/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBClient struct {
	db *pgxpool.Pool
}

func Connect(databaseURL string) (*DBClient, error) {
	db, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	return &DBClient{db: db}, nil
}

func (dbClient *DBClient) InitSchema() error {
	schema := `
    CREATE TABLE IF NOT EXISTS clients (
        uid BIGINT PRIMARY KEY,
        birthday TIMESTAMP,
        sex CHAR(1),
        name TEXT
    );
    CREATE INDEX IF NOT EXISTS idx_clients_birthday ON clients(birthday);
    CREATE INDEX IF NOT EXISTS idx_clients_name ON clients(name);
    `

	_, err := dbClient.db.Exec(context.Background(), schema)
	return err
}

func (dbClient *DBClient) LoadCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	conn, err := dbClient.db.Acquire(context.Background())
	if err != nil {
		return err
	}
	defer conn.Release()

	copySql := "COPY clients(uid, birthday, sex, name) FROM STDIN WITH (FORMAT csv, DELIMITER ',', HEADER true)"
	_, err = conn.Conn().PgConn().CopyFrom(context.Background(), file, copySql)
	if err != nil {
		return err
	}

	log.Println("Data loaded successfully using COPY")
	return nil
}

func (dbClient *DBClient) GetClientByUID(uid int64) (models.Client, error) {
	var client models.Client
	err := dbClient.db.QueryRow(context.Background(), "SELECT uid, birthday, sex, name FROM clients WHERE uid=$1", uid).Scan(&client.UID, &client.Birthday, &client.Sex, &client.Name)
	return client, err
}

func (dbClient *DBClient) GetClientByAgeWithPagination(start, end time.Time, limit, offset int) ([]models.Client, error) {
	var clients []models.Client
	query := `SELECT uid, birthday, sex, name FROM clients WHERE birthday BETWEEN $1 AND $2 ORDER BY birthday LIMIT $3 OFFSET $4`
	rows, err := dbClient.db.Query(context.Background(), query, start, end, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var client models.Client
		if err := rows.Scan(&client.UID, &client.Birthday, &client.Sex, &client.Name); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clients, nil
}

func (dbClient *DBClient) SearchClientByName(name string) ([]models.Client, error) {
	rows, err := dbClient.db.Query(context.Background(), "SELECT uid, birthday, sex, name FROM clients WHERE name ILIKE '%' || $1 || '%' ORDER BY name", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clients := []models.Client{}
	for rows.Next() {
		var client models.Client
		err := rows.Scan(&client.UID, &client.Birthday, &client.Sex, &client.Name)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil
}

func (dbClient *DBClient) OptimizeDatabase() error {
	queries := []string{

		// Create partitions
		`CREATE TABLE IF NOT EXISTS clients_1970 PARTITION OF clients FOR VALUES FROM ('1970-01-01') TO ('1980-01-01');`,
		`CREATE TABLE IF NOT EXISTS clients_1980 PARTITION OF clients FOR VALUES FROM ('1980-01-01') TO ('1990-01-01');`,
		`CREATE TABLE IF NOT EXISTS clients_1990 PARTITION OF clients FOR VALUES FROM ('1990-01-01') TO ('2000-01-01');`,
		`CREATE TABLE IF NOT EXISTS clients_2000 PARTITION OF clients FOR VALUES FROM ('2000-01-01') TO ('2010-01-01');`,
		`CREATE TABLE IF NOT EXISTS clients_2010 PARTITION OF clients FOR VALUES FROM ('2010-01-01') TO ('2020-01-01');`,
		`CREATE TABLE IF NOT EXISTS clients_2020 PARTITION OF clients FOR VALUES FROM ('2020-01-01') TO ('2030-01-01');`,

		// Create materialized views
		`CREATE MATERIALIZED VIEW IF NOT EXISTS client_age_view AS
		SELECT uid, birthday, sex, name
		FROM clients
		WHERE birthday BETWEEN '1970-01-01' AND '1980-01-01';`,
	}

	for _, query := range queries {
		_, err := dbClient.db.Exec(context.Background(), query)
		if err != nil {
			return err
		}
	}

	log.Println("Database optimized successfully")
	return nil
}
