package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-sqlite3"
)

const DB_FILE_NAME = "persist.db"

var DB_CONNECTION_URL = ConstructDbUrl()

var DB_DRIVER string

type DbUtil struct {
}

var DB_UTIL = DbUtil{}

func ConstructDbUrl() string {

	fileName := calculateFullDbFilePath()

	connectionString := "file:" + fileName
	connectionString += "?cache=shared"
	connectionString += "&mode=rwc"
	connectionString += "&_busy_timeout=2000000"

	return connectionString
}

func calculateFullDbFilePath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	return filepath.Join(dir, DB_FILE_NAME)
}

func initializeDriver() {
	sql.Register(DB_DRIVER, &sqlite3.SQLiteDriver{})
}

func OpenDbConnection() (*sql.DB, error) {
	connection, err := sql.Open(DB_DRIVER, DB_CONNECTION_URL)
	connection.SetMaxIdleConns(0)
	return connection, err
}

func ExecuteSelect(query string, args ...interface{}) (*sql.Rows, error) {
	database, openErr := OpenDbConnection()
	if openErr != nil {
		log.Println("Failed to create the handle")
		return nil, openErr
	}
	defer closeDb(database)

	return database.Query(query, args...)
}

func SelectOneRow(query string, args ...interface{}) *sql.Row {
	database, openErr := OpenDbConnection()
	if openErr != nil {
		log.Println("Failed to create the handle")
		return nil
	}
	defer closeDb(database)

	return database.QueryRow(query, args...)
}

func ExecuteUpdate(query string, args ...interface{}) (sql.Result, error) {
	return ExecuteInsert(query, args...)
}

func ExecuteInsert(query string, args ...interface{}) (sql.Result, error) {
	database, openErr := OpenDbConnection()
	if openErr != nil {
		log.Println("Failed to create the handle")
		return DummyResult{}, openErr
	}
	defer closeDb(database)

	stmt, err := database.Prepare(query)
	if err != nil {
		log.Println(err)
	}

	execResult, execErr := stmt.Exec(args...)
	stmt.Close()

	return execResult, execErr
}

func ExecuteDelete(query string, args ...interface{}) (sql.Result, error) {
	database, openErr := OpenDbConnection()
	if openErr != nil {
		log.Println("Failed to create the handle")
		return DummyResult{}, openErr
	}
	defer closeDb(database)

	fkOnRes, fkOnErr := database.Exec("PRAGMA foreign_keys=ON")
	if fkOnErr != nil {
		log.Println(fkOnErr)
		return fkOnRes, fkOnErr
	}

	tx, txErr := database.Begin()
	if txErr != nil {
		log.Println(txErr)
		return nil, txErr
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		log.Println(err)
		tx.Rollback()
		return nil, err
	}

	rows, execErr := stmt.Exec(args...)
	if execErr != nil {
		log.Printf("Unable execute DELETE statement. Reason: %v \n", execErr)
		tx.Rollback()
		return nil, execErr
	}
	closeErr := stmt.Close()
	if closeErr != nil {
		log.Println(closeErr)
		tx.Rollback()
		return nil, execErr
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		log.Printf("Unable to commit DELETE transaction. Reason: %v \n", commitErr)
		tx.Rollback()
		return nil, commitErr
	}

	return rows, execErr
}

func closeDb(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("Unable to close DB connection. Reason:%v\n", err)
	}
}

func closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		log.Printf("Unable to close Rows. Reason:%v\n", err)
	}
}

func (*DbUtil) vacuum() error {
	database, openErr := OpenDbConnection()
	if openErr != nil {
		log.Println("Failed to create the handle")
		return openErr
	}
	defer closeDb(database)

	_, err := database.Exec("VACUUM")
	return err
}

func ParseSqlite3Date(str string) (time.Time, error) {
	return time.Parse(sqlite3.SQLiteTimestampFormats[0], str)
}

type DummyResult struct{}

func (dr DummyResult) LastInsertId() (int64, error) { return 0, nil }
func (dr DummyResult) RowsAffected() (int64, error) { return 0, nil }
