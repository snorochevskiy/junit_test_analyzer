package main

type HttpErrDTO struct {
	Code    int
	Message string
}

type ViewLaunchDTO struct {
	LaunchId int
	Label    string
	Branch   string
	Tests    []*TestCaseEntity

	FailedTestsNum  int
	PassedTestsNum  int
	SkippedTestsNum int
}

type ViewPackageDTO struct {
	LaunchId int
	Package  string
	Tests    []*TestCaseEntity
}

type PackagesDTO struct {
	LaunchId int
	Packages []*PackageEntity
}

type LaunchesDiffDTO struct {
	LaunchId1            int
	LaunchId2            int
	AddedTests           []*TestCaseEntity
	RemovedTests         []*TestCaseEntity
	PassedToFailedTests  []*TestCaseEntity
	PassedToSkippedTests []*TestCaseEntity
	FailedToPassedTests  []*TestCaseEntity
	FailedToSkippedTests []*TestCaseEntity
	SkippedToFailedTests []*TestCaseEntity
	SkippedToPassedTests []*TestCaseEntity
}
