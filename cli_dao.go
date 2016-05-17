package main

import (
	"database/sql"
	"log"
)

func (dao *DaoService) PersistLaunch(launchInfo ParsedLaunchInfo) error {
	connection, err := OpenDbConnection()
	if err != nil {
		return nil
	}
	defer closeDb(connection)

	transaction, err := connection.Begin()
	if err != nil {
		return err
	}

	branchId, err := dao.provideBranchId(transaction, launchInfo.Project, launchInfo.Branch)
	if err != nil {
		transaction.Rollback()
		return err
	}

	res, err := transaction.Exec("INSERT INTO test_launches(parent_branch_id, creation_date, label, test_num, failed_num, skipped_num, passed_num) values(?, ?, ?, ?, ?, ?, ?)",
		branchId, launchInfo.LaunchTime, launchInfo.Label, launchInfo.OveralNum, launchInfo.FailedNum, launchInfo.SkippedNum, launchInfo.PassedNum)
	if err != nil {
		transaction.Rollback()
		return err
	}
	launchId, err := res.LastInsertId()
	if err != nil {
		transaction.Rollback()
		return err
	}

	testStmt, err := transaction.Prepare("INSERT INTO test_cases(name, package, class_name, md5_hash, status, parent_launch_id) values(?, ?, ?, ?, ?, ?)")
	if err != nil {
		transaction.Rollback()
		return err
	}

	failureStmt, err := transaction.Prepare("INSERT INTO test_case_failures(failure_type, failure_message, failure_text, parent_test_case_id) values(?, ?, ?, ?)")
	if err != nil {
		transaction.Rollback()
		return err
	}

	for _, test := range launchInfo.Tests {
		res, err := testStmt.Exec(test.Name, test.Package, test.ClassName, test.Md5Hash, test.Status, launchId)
		if err != nil {
			transaction.Rollback()
			return err
		}
		if test.Failure != nil {
			testId, err := res.LastInsertId()
			if err != nil {
				transaction.Rollback()
				return err
			}
			_, failureAddErr := failureStmt.Exec(test.Failure.Type, test.Failure.Message, test.Failure.Text, testId)
			if failureAddErr != nil {
				transaction.Rollback()
				return failureAddErr
			}
		}
	}

	if commitErr := transaction.Commit(); commitErr != nil {
		log.Printf("Unable to commit new launch INSERT. Reason: %v\n", commitErr)
	}

	return nil
}

func (dao *DaoService) provideBranchId(tx *sql.Tx, projectName string, branchName string) (int64, error) {
	projectId, err := dao.provideProjectId(tx, projectName)
	if err != nil {
		return 0, err
	}

	rows, err := tx.Query("SELECT branch_id FROM project_branches WHERE parent_project_id = ? AND branch_name = ?", projectId, branchName)
	if err != nil {
		return 0, err
	}

	var branchId int64 = 0
	for rows.Next() {
		rows.Scan(&branchId)
	}

	if branchId == 0 {
		res, err := tx.Exec("INSERT INTO project_branches (parent_project_id, branch_name) VALUES (?, ?)", projectId, branchName)
		if err != nil {
			return 0, err
		}
		branchId, err = res.LastInsertId()
		if err != nil {
			return 0, err
		}
	}

	return branchId, nil
}

func (*DaoService) provideProjectId(tx *sql.Tx, projectName string) (int64, error) {
	// Search for project with given project name
	rows, err := tx.Query("SELECT project_id FROM test_projects WHERE project_name = ?", projectName)
	if err != nil {
		return 0, err
	}

	var projectId int64 = 0
	for rows.Next() {
		rows.Scan(&projectId)
	}

	// If such project is absent, then create it
	if projectId == 0 {
		res, err := tx.Exec("INSERT INTO test_projects (project_name) VALUES (?)", projectName)
		if err != nil {
			return 0, err
		}
		projectId, err = res.LastInsertId()
		if err != nil {
			return 0, err
		}
	}
	return projectId, nil
}
