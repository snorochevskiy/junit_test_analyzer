package main

import (
	"database/sql"
	"fmt"
	"time"
)

type Comparable interface {
	IsLess(Comparable) bool
}

type SortableSlice []*BranchDetailedInfoEntity

func (s SortableSlice) Len() int {
	return len(s)
}
func (s SortableSlice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortableSlice) Less(i, j int) bool {
	return s[i].IsLess(s[j])
}

type ProjectEntity struct {
	Id          int64  `column:"project_id",json:"id"`
	Name        string `column:"project_name",json:"name"`
	Description string `column:"description",json:"description"`
}

type ProjectBranchEntity struct {
	Id              int64  `column:"branch_id"`
	ParentProjectId int64  `column:"parent_project_id"`
	Name            string `column:"branch_name"`
}

type BranchDetailedInfoEntity struct {
	Id           int64
	BranchName   string
	CreationDate time.Time `column:"creation_date"`
	LastLaunchId int64     `column:"launch_id"`

	LastLaunchFailedNum sql.NullInt64 `column:"failed_num"`

	//LastLauchFailed bool
}

func (this *BranchDetailedInfoEntity) LastLauchFailed() bool {
	if this.LastLaunchFailedNum.Valid && this.LastLaunchFailedNum.Int64 == 0 {
		return false
	}
	return true
}

func (this *BranchDetailedInfoEntity) IsLess(other *BranchDetailedInfoEntity) bool {
	return this.CreationDate.Before(other.CreationDate)
}

type TestLaunchEntity struct {
	Id             int64     `column:"launch_id"`
	BranchId       int64     `column:"parent_branch_id"`
	Label          string    `column:"label"`
	CreateDate     time.Time `column:"creation_date"`
	FailedTestsNum int
}

func (tle *TestLaunchEntity) String() string {
	return fmt.Sprintf("TestLaunchEntity[Id=%v, BranchId=%v, CreateDate=%v]",
		tle.Id, tle.BranchId, tle.CreateDate)
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
