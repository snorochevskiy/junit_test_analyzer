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
