package db

import (
	"database/sql"
	"encoding/csv"
	"os"
	"time"
	"welcome-to-the-jungle/pkg/models"
)

type DBClient struct {
	DB *sql.DB
}

func Connect(databaseURL string) (DBClient, error) {
	// Connect to the database
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return DBClient{}, err
	}

	// Check if the connection is successful
	err = db.Ping()
	if err != nil {
		return DBClient{}, err
	}
	dbClient := DBClient{
		DB: db,
	}
	return dbClient, nil
}

func (db *DBClient) InitSchema() error {
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

	_, err := db.DB.Exec(schema)
	return err
}

func (db *DBClient) LoadCSV(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '\t'
	reader.FieldsPerRecord = -1

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare("INSERT INTO clients(uid, birthday, sex, name) VALUES($1, $2, $3, $4)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		// check if is better to skip or not
		if len(record) < 4 {
			continue
		}

		uid := record[0]
		birthday := record[1]
		sex := record[2]
		name := record[3]

		_, err = stmt.Exec(uid, birthday, sex, name)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (db *DBClient) GetClientByAge(start time.Time, end time.Time) ([]models.Client, error) {
	rows, err := db.DB.Query("SELECT uid, birthday, sex, name FROM clients WHERE birthday BETWEEN $1 AND $2 ORDER BY name", start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.Client

	for rows.Next() {
		var client models.Client
		if err := rows.Scan(&client.UID, &client.Birthday, &client.Sex, &client.Name); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil

}

func (db *DBClient) SearchClientByName(name string) ([]models.Client, error) {
	rows, err := db.DB.Query("SELECT uid, birthday, sex, name FROM clients WHERE name ILIKE '%' || $1 || '%'", name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []models.Client

	for rows.Next() {
		var client models.Client
		if err := rows.Scan(&client.UID, &client.Birthday, &client.Sex, &client.Name); err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil
}

func (db *DBClient) GetClientByUID(uid int64) (models.Client, error) {
	var client models.Client

	err := db.DB.QueryRow("SELECT uid, birthday, sex, name FROM clients WHERE uid = $1", uid).Scan(&client.UID, &client.Birthday, &client.Sex, &client.Name)
	if err != nil {
		return models.Client{}, err
	}

	return client, nil
}
