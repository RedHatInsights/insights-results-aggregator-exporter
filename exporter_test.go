/*
Copyright © 2021, 2022 Red Hat, Inc.

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

// Unit test definitions for functions and methods defined in source file
// exporter.go

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-exporter
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/exporter_test.html

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/tisnik/go-capture"

	main "github.com/RedHatInsights/insights-results-aggregator-exporter"
)

const (
	expectedVersionMessage        = "Insights Results Aggregator Exporter version 1.0"
	expectedAuthorsMessage        = "Pavel Tisnovsky"
	expectedCopyrightMessage      = "Red Hat Inc."
	expectedConfigurationMessage1 = "Driver"
	expectedConfigurationMessage2 = "Username"
	expectedConfigurationMessage3 = "Host"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

// TestShowVersion checks the function showVersion
func TestShowVersion(t *testing.T) {
	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		main.ShowVersion()
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, expectedVersionMessage)
}

// TestShowAuthors checks the function showAuthors
func TestShowAuthors(t *testing.T) {
	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		main.ShowAuthors()
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, expectedAuthorsMessage)
	assert.Contains(t, output, expectedCopyrightMessage)
}

// TestShowConfiguration checks the function ShowConfiguration
func TestShowConfiguration(t *testing.T) {
	// fill in configuration structure
	configuration := main.ConfigStruct{}
	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		main.ShowConfiguration(&configuration)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, expectedConfigurationMessage1)
	assert.Contains(t, output, expectedConfigurationMessage2)
	assert.Contains(t, output, expectedConfigurationMessage3)
}

func checkCapture(t *testing.T, err error) {
	if err != nil {
		t.Fatal("Unable to capture standard output", err)
	}
}

// TestDoSelectedOperationShowVersion checks the function showVersion called
// via doSelectedOperation function
func TestDoSelectedOperationShowVersion(t *testing.T) {
	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}
	cliFlags := main.CliFlags{
		ShowVersion:       true,
		ShowAuthors:       false,
		ShowConfiguration: false,
	}

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		code, err := main.DoSelectedOperation(&configuration, cliFlags, log.Logger)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, expectedVersionMessage)
}

// TestDoSelectedOperationShowAuthors checks the function showAuthors called
// via doSelectedOperation function
func TestDoSelectedOperationShowAuthors(t *testing.T) {
	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       true,
		ShowConfiguration: false,
	}

	// try to call the tested function and capture its output
	output, err := capture.StandardOutput(func() {
		code, err := main.DoSelectedOperation(&configuration, cliFlags, log.Logger)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, expectedAuthorsMessage)
	assert.Contains(t, output, expectedCopyrightMessage)
}

// TestDoSelectedOperationShowConfiguration checks the function
// showConfiguration called via doSelectedOperation function
func TestDoSelectedOperationShowConfiguration(t *testing.T) {
	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: true,
	}

	// try to call the tested function and capture its output
	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		code, err := main.DoSelectedOperation(&configuration, cliFlags, log.Logger)
		assert.Equal(t, code, main.ExitStatusOK)
		assert.Nil(t, err)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, expectedConfigurationMessage1)
	assert.Contains(t, output, expectedConfigurationMessage2)
	assert.Contains(t, output, expectedConfigurationMessage3)
}

// TestDoSelectedOperationCheckS3Connection checks the function
// checkS3Connection called via doSelectedOperation function
func TestDoSelectedOperationCheckS3Connection(t *testing.T) {
	// stub for structures needed to call the tested function
	configuration := main.ConfigStruct{}
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
		CheckS3Connection: true,
	}

	code, err := main.DoSelectedOperation(&configuration, cliFlags, log.Logger)
	assert.Equal(t, code, main.ExitStatusS3Error)
	assert.Error(t, err)
}

// TestPrintTables checks the function printTables
func TestPrintTables(t *testing.T) {
	tables := []main.TableName{
		main.TableName("first"),
		main.TableName("second"),
		main.TableName("third"),
	}

	output, err := capture.ErrorOutput(func() {
		log.Logger = log.Output(zerolog.New(os.Stderr))
		main.PrintTables(tables)
	})

	// check the captured text
	checkCapture(t, err)

	assert.Contains(t, output, "\\\"table\\\":\\\"first\\\"")
	assert.Contains(t, output, "\\\"table\\\":\\\"second\\\"")
	assert.Contains(t, output, "\\\"table\\\":\\\"third\\\"")
}

// TestParseFlags is dummy test for parseFlags function
func TestParseFlags(t *testing.T) {
	flags := main.ParseFlags()
	assert.NotNil(t, flags)
}

// TestPerformDataExportViaDoSelectedOperation checks the function
// performDataExport.
func TestPerformDataExportViaDoSelectedOperation(t *testing.T) {
	// fill in configuration structure w/o specifying S3 connection or DB
	// connection
	configuration := main.ConfigStruct{}

	// default operation is export data
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
		CheckS3Connection: false,
	}

	// the call should fail
	code, err := main.DoSelectedOperation(&configuration, cliFlags, log.Logger)
	assert.Equal(t, code, main.ExitStatusStorageError)
	assert.Error(t, err)
}

// TestCheckS3Connection checks the function CheckS3Connection
func TestCheckS3Connection(t *testing.T) {
	// fill in configuration structure
	// w/o specifying S3 connection
	configuration := main.ConfigStruct{}

	// the call should fail
	code, err := main.CheckS3Connection(&configuration)
	assert.Equal(t, code, main.ExitStatusS3Error)
	assert.Error(t, err)
}

// TestPerformDataExport checks the function performDataExport.
func TestPerformDataExportNoStorage(t *testing.T) {
	// fill in configuration structure w/o specifying S3 connection or DB
	// connection
	configuration := main.ConfigStruct{}

	// default operation is export data
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
		CheckS3Connection: false,
	}

	// the call should fail
	code, err := main.PerformDataExport(&configuration, cliFlags, log.Logger)
	assert.Equal(t, code, main.ExitStatusStorageError)
	assert.Error(t, err)
}

// TestPerformDataExport checks the function performDataExport.
func TestPerformDataExportConfigError(t *testing.T) {
	// fill in configuration structure w/o specifying S3 connection
	// but DB connection is specified
	configuration := main.ConfigStruct{
		main.StorageConfiguration{
			Driver:        "postgres",
			PGUsername:    "user",
			PGPassword:    "password",
			PGHost:        "nowhere",
			PGPort:        1234,
			PGDBName:      "test",
			PGParams:      "",
			LogSQLQueries: true,
		},
		main.S3Configuration{},
		main.LoggingConfiguration{},
		main.SentryConfiguration{},
	}

	// default operation is export data
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
		CheckS3Connection: false,
	}

	// the call should fail, but now because of improper configuration
	code, err := main.PerformDataExport(&configuration, cliFlags, log.Logger)
	assert.Equal(t, code, main.ExitStatusConfigurationError)
	assert.Error(t, err)
}

// TestPerformDataExport checks the function performDataExport.
func TestPerformDataExportToS3(t *testing.T) {
	// fill in configuration structure w/o specifying S3 connection
	// but DB connection is specified
	configuration := main.ConfigStruct{
		main.StorageConfiguration{
			Driver:        "postgres",
			PGUsername:    "user",
			PGPassword:    "password",
			PGHost:        "nowhere",
			PGPort:        1234,
			PGDBName:      "test",
			PGParams:      "",
			LogSQLQueries: true,
		},
		main.S3Configuration{},
		main.LoggingConfiguration{},
		main.SentryConfiguration{},
	}

	// default operation is export data
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
		CheckS3Connection: false,
		Output:            "S3",
	}

	// the call should fail due to inaccessible S3/Minio
	code, err := main.PerformDataExport(&configuration, cliFlags, log.Logger)
	assert.Equal(t, code, main.ExitStatusS3Error)
	assert.Error(t, err)
}

// TestPerformDataExport checks the function performDataExport.
func TestPerformDataExportToFile(t *testing.T) {
	// fill in configuration structure w/o specifying S3 connection
	// but DB connection is specified
	configuration := main.ConfigStruct{
		main.StorageConfiguration{
			Driver:        "postgres",
			PGUsername:    "user",
			PGPassword:    "password",
			PGHost:        "nowhere",
			PGPort:        1234,
			PGDBName:      "test",
			PGParams:      "",
			LogSQLQueries: true,
		},
		main.S3Configuration{},
		main.LoggingConfiguration{},
		main.SentryConfiguration{},
	}

	// default operation is export data
	cliFlags := main.CliFlags{
		ShowVersion:       false,
		ShowAuthors:       false,
		ShowConfiguration: false,
		CheckS3Connection: false,
		Output:            "file",
	}

	// the call should fail due to inaccessible storage (DB)
	code, err := main.PerformDataExport(&configuration, cliFlags, log.Logger)
	assert.Equal(t, code, main.ExitStatusStorageError)
	assert.Error(t, err)
}

// TestConstructIgnoreTableMapEmptyInput checks the function
// constructIgnoredTablesMap for empty input.
func TestConstructIgnoreTableMapEmptyInput(t *testing.T) {
	m := main.ConstructIgnoredTablesMap("")
	assert.Len(t, m, 0, "Empty map should be returned")
}

// TestConstructIgnoreTableMapOneTable checks the function
// constructIgnoredTablesMap for input with one table only.
func TestConstructIgnoreTableMapOneTable(t *testing.T) {
	m := main.ConstructIgnoredTablesMap("table_name")
	assert.Len(t, m, 1, "Map with one item should be returned")
	assert.Contains(t, m, "table_name", "Table name should be added into a map")
}

// TestConstructIgnoreTableMapTwoTables checks the function
// constructIgnoredTablesMap for input with two tables
func TestConstructIgnoreTableMapTwoTables(t *testing.T) {
	m := main.ConstructIgnoredTablesMap("table1,table2")
	assert.Len(t, m, 2, "Map with two items should be returned")
	assert.Contains(t, m, "table1")
	assert.Contains(t, m, "table2")
}

func TestGetS3Bucket(t *testing.T) {
	config := main.ConfigStruct{}
	assert.Equal(t, "", main.GetS3Bucket(&config))

	testBucketString := "test_bucket"
	config.S3.Bucket = testBucketString
	assert.Equal(t, testBucketString, main.GetS3Bucket(&config))

	testBucketPrefixString := "test_prefix"
	config.S3.Prefix = testBucketPrefixString
	assert.Equal(t, testBucketPrefixString+"/"+testBucketString, main.GetS3Bucket(&config))
}
