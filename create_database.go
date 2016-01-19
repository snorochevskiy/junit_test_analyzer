package main

import (
	"log"
)

const DDL_TESTS_LAUNCHES = `
CREATE TABLE IF NOT EXISTS test_launches (
	launch_id integer PRIMARY KEY AUTOINCREMENT,
	branch TEXT,
	creation_date DATE NOT NULL DEFAULT (datetime('now','localtime'))
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
	name TEXT,
	class_name TEXT,
	status TEXT,
	parent_launch_id INTEGER,
	FOREIGN KEY(parent_launch_id) REFERENCES test_launches(launch_id)
)`

const DDL_TEST_CASE_FAILURE = `
CREATE TABLE IF NOT EXISTS test_case_failures (
	test_case_failure_id integer PRIMARY KEY AUTOINCREMENT,
	failure_type TEXT NULL,
	failure_message TEXT NULL,
	failure_text TEXT NULL,
	parent_test_case_id INTEGER,
	FOREIGN KEY(parent_test_case_id) REFERENCES test_cases(test_case_id)
)`

func createDbIfNotExists() {

	database, operErr := OpenDbConnection()
	if operErr != nil {
		log.Println("Failed to create the handle")
	}
	defer database.Close()

	if pingErr := database.Ping(); pingErr != nil {
		log.Println("Failed to keep connection alive")
	}

	_, ddlTestsLaunchesErr := database.Exec(DDL_TESTS_LAUNCHES)
	if ddlTestsLaunchesErr != nil {
		log.Fatal(ddlTestsLaunchesErr)
	}

	_, ddlTestCasesErr := database.Exec(DDL_TEST_CASES)
	if ddlTestCasesErr != nil {
		log.Fatal(ddlTestCasesErr)
	}

	_, ddlTestCaseFailuresErr := database.Exec(DDL_TEST_CASE_FAILURE)
	if ddlTestCaseFailuresErr != nil {
		log.Fatal(ddlTestCaseFailuresErr)
	}

}
