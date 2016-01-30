package main

import (
	"testing"
)

func TestPrepareTestCase(t *testing.T) {
	testCase := TestCase{Name: "test name", FullClassName: "org.test.TestClass", Skipped: nil, Failure: nil}
	prepareTestCase(&testCase)

	if testCase.ClassName != "TestClass" {
		t.Error("Expected 'TestClass', got ", testCase.ClassName)
	}
	if testCase.Package != "org.test" {
		t.Error("Expected 'org.test', got ", testCase.Package)
	}
	if testCase.Md5Hash != "481d3ea5ac22d549674f2660e706d3d7" {
		t.Error("Expected '481d3ea5ac22d549674f2660e706d3d7', got ", testCase.Md5Hash)
	}
	if testCase.Status != TEST_CASE_STATUS_PASSED {
		t.Error("Expected '"+TEST_CASE_STATUS_PASSED+"', got ", testCase.Md5Hash)
	}

}
