package main

import (
	"database/sql"
	"fmt"
	"time"
)

type TestLaunchEntity struct {
	Id             int64     `column:"launch_id"`
	Branch         string    `column:"branch"`
	Label          string    `column:"label"`
	CreateDate     time.Time `column:"creation_date"`
	FailedTestsNum int
}

func (tle *TestLaunchEntity) String() string {
	return fmt.Sprintf("TestLaunchEntity[Id=%v, Branch=%v, CreateDate=%v]",
		tle.Id, tle.Branch, tle.CreateDate)
}

type TestCaseEntity struct {
	Id           int64  `column:"test_case_id"`
	Name         string `column:"name"`
	Package      string `column:"package"`
	ClassName    string `column:"class_name"`
	Status       string `column:"status"`
	TestLaunchId int64  `column:"parent_launch_id"`
	//	Md5Hash      string `column:"md5_hash"`

	FailureInfo *FailureEntity
}

func (tce *TestCaseEntity) String() string {
	return fmt.Sprintf("TestCaseEntity[Id=%v, Name=%v, ClassName=%v, Status=%v, TestLaunchId=%v]",
		tce.Id, tce.Name, tce.ClassName, tce.Status, tce.TestLaunchId)
}

type TestFullInfoEntity struct {
	Id           int64  `column:"test_case_id"`
	Name         string `column:"name"`
	Package      string `column:"package"`
	ClassName    string `column:"class_name"`
	Status       string `column:"status"`
	TestLaunchId int64  `column:"parent_launch_id"`
	//Md5Hash      string `column:"md5_hash"`

	Branch     string    `column:"branch"`
	CreateDate time.Time `column:"creation_date"`

	FailureId sql.NullInt64 `column:"test_case_failure_id"`
}

type FailureEntity struct {
	Id      int64  `column:"test_case_failure_id"`
	Message string `column:"failure_message"`
	Type    string `column:"failure_type"`
	Text    string `column:"failure_text"`
}

func (fe *FailureEntity) String() string {
	return fmt.Sprintf("FailureEntity[Id=%v, Message=%v, Type=%v, Status=%v, Text=%v]",
		fe.Id, fe.Message, fe.Type, fe.Text)
}

type PackageEntity struct {
	Package         string `column:"package"`
	TestsNum        int    `column:"tests_num"`
	FailedTestsNum  int
	PassedTestsNum  int
	SkippedTestsNum int
}

type UserEntity struct {
	UserId    int64  `column:"user_id"`
	Login     string `column:"login"`
	Password  string `column:"password"`
	IsActive  bool   `column:"is_active"`
	FirstName string `column:"first_name"`
	LastName  string `column:"last_name"`
}
