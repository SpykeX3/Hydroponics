package db

import (
	"database/sql"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"path"
	"time"
)

type ReadingsDB struct {
	db *sql.DB
}

func OpenReadingsDB() (*ReadingsDB, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	readingsPath := path.Join(homedir, "readings.db")

	db, err := sql.Open("sqlite", readingsPath)
	if err != nil {
		log.Fatal(err)
	}
	sqlStmt := `
    CREATE TABLE IF NOT EXISTS readings (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        reading INTEGER NOT NULL,
        timestamp  INTEGER NOT NULL
    );
    `
	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}
	log.Println("Readings database is ready")
	return &ReadingsDB{db: db}, nil
}

func (rdb *ReadingsDB) Close() {
	_ = rdb.db.Close()
}

func (rdb *ReadingsDB) InsertReading(level int) error {
	_, err := rdb.db.Exec("INSERT INTO readings (reading, timestamp) VALUES (?, ?)", level, time.Now().Unix())
	if err != nil {
		return err
	}
	return nil
}

func (rdb *ReadingsDB) IsEmpty() bool {
	var count int
	err := rdb.db.QueryRow("SELECT COUNT(*) FROM readings").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	return count == 0
}

func (rdb *ReadingsDB) GetLastReading() (int, error) {
	var reading int
	err := rdb.db.QueryRow("SELECT reading FROM readings ORDER BY id DESC LIMIT 1").Scan(&reading)
	if err != nil {
		return 0, err
	}
	return reading, nil
}
