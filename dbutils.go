package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-sqlite3"
)

var DB_CONNECTION_URL = ConstructDbUrl()

var DB_DRIVER string

func ConstructDbUrl() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatalln(err)
	}
	fileName := filepath.Join(dir, "persist.db")
	return "file:" + fileName + "?cache=shared&mode=rwc&_busy_timeout=50000000"
}

func initializeDriver() {
	sql.Register(DB_DRIVER, &sqlite3.SQLiteDriver{})
}

func OpenDbConnection() (*sql.DB, error) {
	return sql.Open(DB_DRIVER, DB_CONNECTION_URL)
}

func ExecuteSelect(query string, args ...interface{}) (*sql.Rows, error) {
	database, openErr := OpenDbConnection()
	if openErr != nil {
		log.Println("Failed to create the handle")
		return nil, openErr
	}
	defer database.Close()

	return database.Query(query, args...)
}

func SelectOneRow(query string, args ...interface{}) *sql.Row {
	database, openErr := OpenDbConnection()
	if openErr != nil {
		log.Println("Failed to create the handle")
		return nil
	}
	defer database.Close()

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
	defer database.Close()

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
	defer database.Close()

	fkOnRes, fkOnErr := database.Exec("PRAGMA foreign_keys=ON")
	if fkOnErr != nil {
		log.Println(fkOnErr)
		return fkOnRes, fkOnErr
	}

	stmt, err := database.Prepare(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	rows, execErr := stmt.Exec(args...)
	stmt.Close()

	return rows, execErr
}

func ParseSqlite3Date(str string) (time.Time, error) {
	return time.Parse(sqlite3.SQLiteTimestampFormats[0], str)
}

type DummyResult struct{}

func (dr DummyResult) LastInsertId() (int64, error) { return 0, nil }
func (dr DummyResult) RowsAffected() (int64, error) { return 0, nil }
