package main

import (
	"fmt"
	"time"
)

type TestLaunchEntity struct {
	Id             int64     `column:"id"`
	Branch         string    `column:"branch"`
	CreateDate     time.Time `column:"creation_date"`
	FailedTestsNum int
}

func (tle *TestLaunchEntity) String() string {
	return fmt.Sprintf("TestLaunchEntity[Id=%v, Branch=%v, CreateDate=%v]",
		tle.Id, tle.Branch, tle.CreateDate)
}

type TestCaseEntity struct {
	Id           int64  `column:"id"`
	Name         string `column:"name"`
	ClassName    string `column:"class_name"`
	Status       string `column:"status"`
	TestLaunchId int64  `column:"test_launch_id"`
}

func (tce *TestCaseEntity) String() string {
	return fmt.Sprintf("TestCaseEntity[Id=%v, Name=%v, ClassName=%v, Status=%v, TestLaunchId=%v]",
		tce.Id, tce.Name, tce.ClassName, tce.Status, tce.TestLaunchId)
}
