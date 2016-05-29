package main

import (
	"jutra/session"
)

type RenderObject struct {
	User *session.UserRenderInfo
	Data interface{}
}

type HttpErrDTO struct {
	Code    int
	Message string
}

type ViewLaunchDTO struct {
	LaunchId int64
	Label    string
	BranchId int64
	Tests    []*TestCaseEntity

	FailedTestsNum  int
	PassedTestsNum  int
	SkippedTestsNum int
}

type ViewPackageDTO struct {
	LaunchId int64
	Package  string
	Tests    []*TestCaseEntity
}

type PackagesDTO struct {
	LaunchId int64
	Packages []*PackageEntity
}

type LaunchesDiffDTO struct {
	LaunchId1            int64
	LaunchId2            int64
	AddedTests           []*TestCaseEntity
	RemovedTests         []*TestCaseEntity
	PassedToFailedTests  []*TestCaseEntity
	PassedToSkippedTests []*TestCaseEntity
	FailedToPassedTests  []*TestCaseEntity
	FailedToSkippedTests []*TestCaseEntity
	SkippedToFailedTests []*TestCaseEntity
	SkippedToPassedTests []*TestCaseEntity
}

type DbManagmentRO struct {
	DbInfo    DatabaseInfo
	ActionErr error
}
type DatabaseInfo struct {
	DbFileName string
	DbFileSize int64
}

type MainPageRO struct {
	Projects []*ProjectEntity
}
