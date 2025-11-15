package db

import (
	"database/sql"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"path"
	"time"
)

const (
	dbFileName      = "water-level.db"
	createTablesSQL = `
    CREATE TABLE IF NOT EXISTS readings (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        reading INTEGER NOT NULL,
        timestamp  INTEGER NOT NULL
    );
    CREATE TABLE IF NOT EXISTS state (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        state TEXT NOT NULL CHECK (state IN ('FILL','EMPTY')),
        timestamp INTEGER NOT NULL
    );`
)

type StateDB struct {
	db *sql.DB
}

// State represents the reservoir state.
type State string

const (
	StateFill  State = "FILL"
	StateEmpty State = "EMPTY"
)

func OpenStateDB() (*StateDB, error) {
	dbPath, err := getDatabasePath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	if err := initializeTables(db); err != nil {
		db.Close()
		return nil, err
	}

	log.Println("Readings and state database are ready")
	return &StateDB{db: db}, nil
}

func getDatabasePath() (string, error) {
	homedir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(homedir, dbFileName), nil
}

func initializeTables(db *sql.DB) error {
	_, err := db.Exec(createTablesSQL)
	return err
}

func (rdb *StateDB) Close() {
	_ = rdb.db.Close()
}

func (rdb *StateDB) InsertReading(level int) error {
	_, err := rdb.db.Exec("INSERT INTO readings (reading, timestamp) VALUES (?, ?)", level, time.Now().Unix())
	return err
}

func (rdb *StateDB) IsReadingsEmpty() bool {
	var count int
	err := rdb.db.QueryRow("SELECT COUNT(*) FROM readings").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	return count == 0
}

func (rdb *StateDB) GetLastReading() (int, error) {
	var reading int
	err := rdb.db.QueryRow("SELECT reading FROM readings ORDER BY id DESC LIMIT 1").Scan(&reading)
	if err != nil {
		return 0, err
	}
	return reading, nil
}

// InsertState inserts a new state row with the current timestamp.
func (rdb *StateDB) InsertState(state State) error {
	_, err := rdb.db.Exec("INSERT INTO state (state, timestamp) VALUES (?, ?)", string(state), time.Now().Unix())
	return err
}

// IsStateEmpty reports whether the state table has no rows.
func (rdb *StateDB) IsStateEmpty() bool {
	var count int
	err := rdb.db.QueryRow("SELECT COUNT(*) FROM state").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	return count == 0
}

// GetLastState returns the last inserted state.
func (rdb *StateDB) GetLastState() (State, int64, error) {
	var s string
	var timestamp int64
	err := rdb.db.QueryRow("SELECT state, timestamp FROM state ORDER BY id DESC LIMIT 1").Scan(&s, &timestamp)
	if err != nil {
		return "", 0, err
	}
	return State(s), timestamp, nil
}
