package main

import (
	"fmt"
	"strings"
	"time"
)

type TestLaunchEntity struct {
	Id             int64     `column:"launch_id"`
	Branch         string    `column:"branch"`
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
	ClassName    string `column:"class_name"`
	Status       string `column:"status"`
	TestLaunchId int64  `column:"parent_launch_id"`

	FailureInfo *FailureEntity
}

func (tce *TestCaseEntity) GetPackage() string {
	return tce.ClassName[0:strings.LastIndex(tce.ClassName, ".")]
}

func (tce *TestCaseEntity) GetClassName() string {
	return tce.ClassName[strings.LastIndex(tce.ClassName, ".")+1:]
}

func (tce *TestCaseEntity) String() string {
	return fmt.Sprintf("TestCaseEntity[Id=%v, Name=%v, ClassName=%v, Status=%v, TestLaunchId=%v]",
		tce.Id, tce.Name, tce.ClassName, tce.Status, tce.TestLaunchId)
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
