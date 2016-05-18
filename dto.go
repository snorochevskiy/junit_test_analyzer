package main

type RenderObject struct {
	User *UserRenderInfo
	Data interface{}
}

type UserRenderInfo struct {
	LoggedIn bool
	Details  *UserEntity
}

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
