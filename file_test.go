/*
Copyright © 2022 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main_test

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-exporter
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/file_test.html

import (
	"os"
	"testing"

	main "github.com/RedHatInsights/insights-results-aggregator-exporter"

	"github.com/stretchr/testify/assert"
)

// mustCreateTemporaryDirectory helper function creates temporary directory
// that will be cleaned up after tests
func mustCreateTemporaryDirectory(t *testing.T) string {
	directory, err := os.MkdirTemp(os.TempDir(), "exporter")
	if err != nil {
		t.Fatal(err)
	}

	return directory
}

// mustReadFile helper function tries to read specified file and return its
// content as a string
func mustReadFile(t *testing.T, filename string) string {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		t.Fatal(err)
	}

	return string(fileContent)
}

// mustDeleteFile helper function tries to delete specified file
func mustDeleteFile(t *testing.T, filename string) {
	err := os.Remove(filename)
	if err != nil {
		t.Fatal(err)
	}
}

// checkFileContent helper function checks if file has the expected content
func checkFileContent(t *testing.T, filename, expected string) {
	content := mustReadFile(t, filename)
	assert.Equal(t, expected, content)
}

// mustRemoveTempDirectory helper function make sure the temporary directory is
// deleted after testing
func mustRemoveTempDirectory(t *testing.T, directory string) {
	err := os.RemoveAll(directory)
	if err != nil {
		t.Fatal(err)
	}
}

// TestStoreTableNamesIntoFileNoWritableFile checks that error is thrown when
// file can not be created
func TestStoreTableNamesIntoFileNoWritableFile(t *testing.T) {
	const filename = ""
	tableNames := []main.TableName{}

	err := main.StoreTableNamesIntoFile(filename, tableNames)
	assert.Error(t, err, "Error should be thrown for empty file name")
}

// TestStoreTableNamesIntoFileEmptyListOfTables check the behaviour if empty
// list of tables is pass into the storeDisabledRulesIntoFile function
func TestStoreTableNamesIntoFileEmptyListOfTables(t *testing.T) {
	directory := mustCreateTemporaryDirectory(t)
	defer mustRemoveTempDirectory(t, directory)

	filename := directory + "tables.csv"
	tableNames := []main.TableName{}

	// just to be sure
	assert.NoFileExists(t, filename, "File must not exist")

	err := main.StoreTableNamesIntoFile(filename, tableNames)
	assert.NoError(t, err, "Error should not be thrown for regular file name")

	// file with exported data must be created
	assert.FileExists(t, filename, "File must be created")

	// check generated file content
	expected := "Table name\n"
	checkFileContent(t, filename, expected)

	// delete temporary file
	mustDeleteFile(t, filename)
}

// TestStoreTableNamesIntoFile check the behaviour of
// storeDisabledRulesIntoFile function
func TestStoreTableNamesIntoFile(t *testing.T) {
	directory := mustCreateTemporaryDirectory(t)
	defer mustRemoveTempDirectory(t, directory)

	filename := directory + "tables.csv"
	tableNames := []main.TableName{
		main.TableName("first"),
		main.TableName("second"),
	}

	// just to be sure
	assert.NoFileExists(t, filename, "File must not exist")

	err := main.StoreTableNamesIntoFile(filename, tableNames)
	assert.NoError(t, err, "Error should not be thrown for regular file name")

	// file with exported data must be created
	assert.FileExists(t, filename, "File must be created")

	// check generated file content
	const expected = "Table name\nfirst\nsecond\n"
	checkFileContent(t, filename, expected)

	// delete temporary file
	mustDeleteFile(t, filename)
}

// TestStoreDisableRulesIntoFile checks that error is thrown when
// file can not be created
func TestStoreDisableRulesIntoFile(t *testing.T) {
	const filename = ""
	disabledRules := []main.DisabledRuleInfo{}

	err := main.StoreDisabledRulesIntoFile(filename, disabledRules)
	assert.Error(t, err, "Error should be thrown for empty file name")
}

// TestStoreDisabledRulesIntoFileEmptyListOfTables check the behaviour if empty
// list of disabled rules is pass into the storeDisabledRulesIntoFile function
func TestStoreDisabledRulesIntoFileEmptyListOfTables(t *testing.T) {
	directory := mustCreateTemporaryDirectory(t)
	defer mustRemoveTempDirectory(t, directory)

	filename := directory + "disabled_rules.csv"
	disabledRules := []main.DisabledRuleInfo{}

	// just to be sure
	assert.NoFileExists(t, filename, "File must not exist")

	err := main.StoreDisabledRulesIntoFile(filename, disabledRules)
	assert.NoError(t, err, "Error should not be thrown for regular file name")

	// file with exported data must be created
	assert.FileExists(t, filename, "File must be created")

	// check generated file content
	expected := "Rule,Count\n"
	checkFileContent(t, filename, expected)

	// delete temporary file
	mustDeleteFile(t, filename)
}

// TestStoreDisabledRulesIntoFile check the behaviour of
// storeDisabledRulesIntoFile function
func TestStoreDisabledRulesIntoFile(t *testing.T) {
	directory := mustCreateTemporaryDirectory(t)
	defer mustRemoveTempDirectory(t, directory)

	filename := directory + "disabled_rules.csv"
	disabledRules := []main.DisabledRuleInfo{
		{"first", 1},
		{"second", 2},
		{"third", 3},
	}

	// just to be sure
	assert.NoFileExists(t, filename, "File must not exist")

	err := main.StoreDisabledRulesIntoFile(filename, disabledRules)
	assert.NoError(t, err, "Error should not be thrown for regular file name")

	// file with exported data must be created
	assert.FileExists(t, filename, "File must be created")

	// check generated file content
	expected := "Rule,Count\nfirst,1\nsecond,2\nthird,3\n"
	checkFileContent(t, filename, expected)

	// delete temporary file
	mustDeleteFile(t, filename)
}
