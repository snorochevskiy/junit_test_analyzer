package main

import (
	"database/sql"
	"log"
)

const DDL_TESTS_LAUNCHES = `
CREATE TABLE IF NOT EXISTS test_launches (
	launch_id integer PRIMARY KEY AUTOINCREMENT,
	branch TEXT,
	label TEXT NULL,
	creation_date DATE NOT NULL
)`

const DDL_TEST_SUITES = `
CREATE TABLE IF NOT EXISTS test_suites (
	test_suite_id integer PRIMARY KEY AUTOINCREMENT,
	name TEXT,
	parent_launch_id INTEGER,
	FOREIGN KEY(parent_launch_id) REFERENCES test_launches(launch_id)
)`

const DDL_TEST_CASES = `
CREATE TABLE IF NOT EXISTS test_cases (
	test_case_id integer PRIMARY KEY AUTOINCREMENT,
	md5_hash TEXT,
	name TEXT,
	package TEXT,
	class_name TEXT,
	status TEXT,
	parent_launch_id INTEGER REFERENCES test_launches(launch_id) ON DELETE CASCADE
)`

const DDL_TEST_CASE_FAILURE = `
CREATE TABLE IF NOT EXISTS test_case_failures (
	test_case_failure_id integer PRIMARY KEY AUTOINCREMENT,
	failure_type TEXT NULL,
	failure_message TEXT NULL,
	failure_text TEXT NULL,
	parent_test_case_id INTEGER REFERENCES test_cases(test_case_id) ON DELETE CASCADE 
)`

const DDL_USERS = `
CREATE TABLE IF NOT EXISTS users (
	user_id integer PRIMARY KEY AUTOINCREMENT,
	login TEXT UNIQUE NOT NULL,
	password TEXT,
	is_active BOOLEAN DEFAULT 0,
	first_name TEXT NULL,
	last_name TEXT NULL
)`

const SQL_REMOVED_ORPHAN_TESTS = `
	DELETE FROM test_cases WHERE parent_launch_id IN (
		SELECT DISTINCT parent_launch_id
		FROM test_cases LEFT JOIN test_launches
		ON parent_launch_id=launch_id
		WHERE launch_id is NULL
	)
`

const DDL_INDEX_TESTS_LAUNCHES_BRANCH = `
	CREATE INDEX ind_test_launches_branch ON test_launches(branch)
`

const DDL_INDEX_TEST_CAST_PARENT_ID = `
	CREATE INDEX ind_test_cases_prnt_id ON test_cases (parent_launch_id)
`

func createDbIfNotExists() {

	database, operErr := OpenDbConnection()
	if operErr != nil {
		log.Println("Failed to create the handle")
	}
	defer database.Close()

	if pingErr := database.Ping(); pingErr != nil {
		log.Fatal("Failed to keep connection alive")
	}

	if _, err := database.Exec(DDL_TESTS_LAUNCHES); err != nil {
		log.Fatal(err)
	}

	if _, err := database.Exec(DDL_TEST_CASES); err != nil {
		log.Fatal(err)
	}

	if _, err := database.Exec(DDL_TEST_CASE_FAILURE); err != nil {
		log.Fatal(err)
	}

	if _, err := database.Exec(DDL_USERS); err != nil {
		log.Fatal(err)
	}
	initUsers(database)

}

func initUsers(database *sql.DB) {
	row := database.QueryRow("SELECT count(*) FROM users")

	var numberOfUsers int
	if err := row.Scan(&numberOfUsers); err != nil {
		log.Fatal(err)
	}

	if numberOfUsers > 0 {
		return
	}

	if _, err := database.Exec("INSERT INTO users(login, password, is_active) VALUES('admin', 'admin', 1)"); err != nil {
		log.Fatal(err)
	}
}
