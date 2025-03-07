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

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-exporter
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/storage_test.html

import (
	"errors"
	"os"
	"testing"

	"database/sql"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	main "github.com/RedHatInsights/insights-results-aggregator-exporter"
)

const NoLimits = -1

var testConfig = main.StorageConfiguration{
	Driver:                 "postgres",
	PGUsername:             "user",
	PGPassword:             "password",
	PGHost:                 "nowhere",
	PGPort:                 1234,
	PGDBName:               "test",
	PGParams:               "",
	LogSQLQueries:          true,
	EnableOrgIDFiltering:   false,
	OrganizationIDsCSVFile: "",
	OrganizationsToExport:  nil,
}

// TestNewStorage checks whether constructor for new storage returns error for improper storage configuration
func TestNewStorageError(t *testing.T) {
	_, err := main.NewStorage(&main.StorageConfiguration{
		Driver: "non existing driver",
	})
	assert.EqualError(t, err, "driver non existing driver is not supported")
}

// TestNewStoragePostgreSQL function tests creating new storage with logs
func TestNewStoragePostgreSQL(t *testing.T) {
	_, err := main.NewStorage(&testConfig)

	// we just happen to make connection without trying to actually connect
	assert.Nil(t, err)
}

// TestNewStorageSQLite3 function tests creating new storage with logs
func TestNewStorageSQLite3(t *testing.T) {
	_, err := main.NewStorage(&main.StorageConfiguration{
		Driver:        "sqlite3",
		LogSQLQueries: true,
	})

	// we just happen to make connection without trying to actually connect
	assert.Nil(t, err)
}

// TestClose function tests database close operation.
func TestClose(t *testing.T) {
	storage, err := main.NewStorage(&main.StorageConfiguration{
		Driver:        "sqlite3",
		LogSQLQueries: true,
	})

	// we just happen to make connection without trying to actually connect
	assert.Nil(t, err)

	// try to close the storage
	err = storage.Close()

	// it should not fail
	assert.Nil(t, err)
}

// mustCreateMockConnection function tries to create a new mock connection and
// checks if the operation was finished without problems.
func mustCreateMockConnection(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	// try to initialize new mock connection
	connection, mock, err := sqlmock.New()

	// check the status
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	return connection, mock
}

// checkConnectionClose function perform mocked DB closing operation and checks
// if the connection is properly closed from unit tests.
func checkConnectionClose(t *testing.T, connection *sql.DB) {
	// connection to mocked DB needs to be closed properly
	err := connection.Close()

	// check the error status
	if err != nil {
		t.Fatalf("error during closing connection: %v", err)
	}
}

// checkAllExpectations function checks if all database-related operations have
// been really met.
func checkAllExpectations(t *testing.T, mock sqlmock.Sqlmock) {
	// check if all expectations were met
	err := mock.ExpectationsWereMet()

	// check the error status
	if err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

// Expected queries
const (
	readRecordCountQuery          = "SELECT count\\(\\*\\) FROM TESTED_TABLE"
	readDisabledRulesQuery        = "SELECT rule_id, count\\(rule_id\\) AS rule_count FROM rule_disable GROUP BY rule_id HAVING count\\(rule_id\\)\\>1 ORDER BY rule_count DESC;"
	readListOfTablesQueryPostgres = `
           SELECT tablename
             FROM pg_catalog.pg_tables
            WHERE schemaname != 'information_schema'
              AND schemaname != 'pg_catalog'
              AND schemaname != 'dvo';
`
	readListOfTablesQuerySQLite = `
           SELECT name FROM sqlite_master
            WHERE type IN \('table','view'\)
              AND name NOT LIKE 'sqlite_%'
            ORDER BY 1;

`
	readTableQuery       = "SELECT \\* FROM table_name"
	readColumnTypesQuery = "SELECT \\* FROM table_name LIMIT 1"
)

// check the function ReadRecordCount
func TestReadRecordCount(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rowsCount := sqlmock.NewRows([]string{"count"})
	expected := 100
	rowsCount.AddRow(expected)

	// expected query performed by tested function
	mock.ExpectQuery(readRecordCountQuery).WillReturnRows(rowsCount)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	count, err := storage.ReadRecordsCount("TESTED_TABLE")
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if count != expected {
		t.Errorf("wrong number records returned: %d", count)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

func TestReadRecordCountScanError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rowsCount := sqlmock.NewRows([]string{"count"})
	rowsCount.AddRow("this is not integer")

	// expected query performed by tested function
	mock.ExpectQuery(readRecordCountQuery).WillReturnRows(rowsCount)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	_, err := storage.ReadRecordsCount("TESTED_TABLE")
	if err == nil {
		t.Errorf("error is expected")
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

func TestReadRecordCountOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	mock.ExpectQuery(readRecordCountQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	count, err := storage.ReadRecordsCount("TESTED_TABLE")
	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	if count != -1 {
		t.Errorf("wrong number records returned: %d", count)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadRecordCount with selective export enabled, but table isn't in the predefined list
func TestReadRecordCountSelectiveExportDisallowedTable(t *testing.T) {
	config := &testConfig
	config.EnableOrgIDFiltering = true
	config.OrganizationsToExport = []string{"1", "42"}

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rowsCount := sqlmock.NewRows([]string{"count"})
	expected := 100
	rowsCount.AddRow(expected)

	// expected query performed by tested function
	mock.ExpectQuery(readRecordCountQuery).WillReturnRows(rowsCount)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, config)

	// call the tested method
	count, err := storage.ReadRecordsCount("TESTED_TABLE")
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if count != expected {
		t.Errorf("wrong number records returned: %d", count)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadRecordCount with seletive export enabled for a single org_id to export
func TestReadRecordCountSelectiveExportAllowedTableSingleOrgID(t *testing.T) {
	config := &testConfig
	config.EnableOrgIDFiltering = true
	config.OrganizationsToExport = []string{"1"}

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rowsCount := sqlmock.NewRows([]string{"count"})
	expected := 100
	rowsCount.AddRow(expected)

	expectedQuery := "SELECT count\\(\\*\\) FROM report WHERE org_id IN \\('1'\\)"
	// expected query performed by tested function
	mock.ExpectQuery(expectedQuery).WillReturnRows(rowsCount)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, config)

	// call the tested method
	count, err := storage.ReadRecordsCount("report")
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if count != expected {
		t.Errorf("wrong number records returned: %d", count)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadRecordCount with seletive export enabled
func TestReadRecordCountSelectiveExportAllowedTable(t *testing.T) {
	config := &testConfig
	config.EnableOrgIDFiltering = true
	config.OrganizationsToExport = []string{"1", "42"}

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rowsCount := sqlmock.NewRows([]string{"count"})
	expected := 100
	rowsCount.AddRow(expected)

	expectedQuery := "SELECT count\\(\\*\\) FROM report WHERE org_id IN \\('1','42'\\)"
	// expected query performed by tested function
	mock.ExpectQuery(expectedQuery).WillReturnRows(rowsCount)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, config)

	// call the tested method
	count, err := storage.ReadRecordsCount("report")
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if count != expected {
		t.Errorf("wrong number records returned: %d", count)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadListOfTables
func TestReadListOfTables(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"tablename"})
	rows.AddRow("foo")
	rows.AddRow("bar")
	rows.AddRow("baz")

	// expected query performed by tested function
	mock.ExpectQuery(readListOfTablesQueryPostgres).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	tableNames, err := storage.ReadListOfTables()
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if len(tableNames) != 3 {
		t.Errorf("wrong number records returned: %d", len(tableNames))
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadListOfTables
func TestReadListOfTablesSQLiteDriver(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"tablename"})
	rows.AddRow("foo")
	rows.AddRow("bar")
	rows.AddRow("baz")

	// expected query performed by tested function
	mock.ExpectQuery(readListOfTablesQuerySQLite).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverSQLite3, &testConfig)

	// call the tested method
	tableNames, err := storage.ReadListOfTables()
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if len(tableNames) != 3 {
		t.Errorf("wrong number records returned: %d", len(tableNames))
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadListOfTables
func TestReadListOfTablesInvalidDriver(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, 2+main.DBDriverSQLite3, &testConfig)

	// call the tested method
	_, err := storage.ReadListOfTables()
	if err == nil {
		t.Errorf("error was expected")
	}

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadListOfTables
func TestReadListOfTablesOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	mock.ExpectQuery(readListOfTablesQueryPostgres).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	_, err := storage.ReadListOfTables()
	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadListOfTables
func TestReadListOfTablesScanError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"tablename"})
	rows.AddRow(1)
	rows.AddRow(2)
	rows.AddRow(3)

	// expected query performed by tested function
	mock.ExpectQuery(readListOfTablesQueryPostgres).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	_, err := storage.ReadListOfTables()
	if err == nil {
		t.Errorf("error is expected")
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadTable
func TestReadTable(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	column1 := sqlmock.NewColumn("id").OfType("INT4", int64(0))
	column2 := sqlmock.NewColumn("value").OfType("FLOAT64", float64(0.0))
	column3 := sqlmock.NewColumn("text").OfType("VARCHAR", "")
	column4 := sqlmock.NewColumn("valid").OfType("BOOL", false)

	// columns of different types
	rows := mock.NewRowsWithColumnDefinition(column1, column2, column3, column4)

	rows.AddRow(1, 1.2, "foo", true)
	rows.AddRow(2, 1.5, "bar", false)
	rows.AddRow(3, 2.0, "baz", true)

	// expected query performed by tested function
	mock.ExpectQuery(readTableQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	values, err := storage.ReadTable("table_name", NoLimits)
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if len(values) != 3 {
		t.Errorf("wrong number records returned: %d", len(values))
	}

	assert.Equal(t, values[0]["id"], int64(1))
	assert.Equal(t, values[1]["id"], int64(2))
	assert.Equal(t, values[2]["id"], int64(3))
	assert.Equal(t, values[0]["text"], "foo")
	assert.Equal(t, values[1]["text"], "bar")
	assert.Equal(t, values[2]["text"], "baz")

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadTable with limits set
func TestReadTableWithLimits(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	column1 := sqlmock.NewColumn("id").OfType("INT4", int64(0))
	column2 := sqlmock.NewColumn("value").OfType("FLOAT64", float64(0.0))
	column3 := sqlmock.NewColumn("text").OfType("VARCHAR", "")
	column4 := sqlmock.NewColumn("valid").OfType("BOOL", false)

	// columns of different types
	rows := mock.NewRowsWithColumnDefinition(column1, column2, column3, column4)

	rows.AddRow(1, 1.2, "foo", true)
	rows.AddRow(2, 1.5, "bar", false)

	// expected query performed by tested function
	mock.ExpectQuery(readTableQuery + " LIMIT 2").WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	values, err := storage.ReadTable("table_name", 2)
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if len(values) != 2 {
		t.Errorf("wrong number records returned: %d", len(values))
	}

	assert.Equal(t, values[0]["id"], int64(1))
	assert.Equal(t, values[1]["id"], int64(2))
	assert.Equal(t, values[0]["text"], "foo")
	assert.Equal(t, values[1]["text"], "bar")

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadTable with selective export enabled
func TestReadTableWithSelectiveExportDisallowedTable(t *testing.T) {
	config := &testConfig
	config.EnableOrgIDFiltering = true
	config.OrganizationsToExport = []string{"1", "42"}

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	column1 := sqlmock.NewColumn("id").OfType("INT4", int64(0))
	column2 := sqlmock.NewColumn("value").OfType("FLOAT64", float64(0.0))
	column3 := sqlmock.NewColumn("text").OfType("VARCHAR", "")
	column4 := sqlmock.NewColumn("valid").OfType("BOOL", false)

	// columns of different types
	rows := mock.NewRowsWithColumnDefinition(column1, column2, column3, column4)

	rows.AddRow(1, 1.2, "foo", true)
	rows.AddRow(2, 1.5, "bar", false)

	// expected query performed by tested function
	mock.ExpectQuery(readTableQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, config)

	// call the tested method, table name must be in predefined list
	values, err := storage.ReadTable("table_name", NoLimits)
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if len(values) != 2 {
		t.Errorf("wrong number records returned: %d", len(values))
	}

	assert.Equal(t, values[0]["id"], int64(1))
	assert.Equal(t, values[1]["id"], int64(2))
	assert.Equal(t, values[0]["text"], "foo")
	assert.Equal(t, values[1]["text"], "bar")

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadTable with selective export enabled for a single org_id to export
func TestReadTableWithSelectiveExportAllowedTableSingleOrgID(t *testing.T) {
	config := &testConfig
	config.EnableOrgIDFiltering = true
	config.OrganizationsToExport = []string{"1"}

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	column1 := sqlmock.NewColumn("id").OfType("INT4", int64(0))
	column2 := sqlmock.NewColumn("value").OfType("FLOAT64", float64(0.0))
	column3 := sqlmock.NewColumn("text").OfType("VARCHAR", "")
	column4 := sqlmock.NewColumn("valid").OfType("BOOL", false)

	// columns of different types
	rows := mock.NewRowsWithColumnDefinition(column1, column2, column3, column4)

	rows.AddRow(1, 1.2, "foo", true)
	rows.AddRow(2, 1.5, "bar", false)

	// expected query performed by tested function
	mock.ExpectQuery("SELECT \\* FROM report WHERE org_id IN \\('1'\\)").WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, config)

	// call the tested method, table name must be in predefined list
	values, err := storage.ReadTable("report", NoLimits)
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if len(values) != 2 {
		t.Errorf("wrong number records returned: %d", len(values))
	}

	assert.Equal(t, values[0]["id"], int64(1))
	assert.Equal(t, values[1]["id"], int64(2))
	assert.Equal(t, values[0]["text"], "foo")
	assert.Equal(t, values[1]["text"], "bar")

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadTable with selective export enabled
func TestReadTableWithSelectiveExportAllowedTable(t *testing.T) {
	config := &testConfig
	config.EnableOrgIDFiltering = true
	config.OrganizationsToExport = []string{"1", "42"}

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	column1 := sqlmock.NewColumn("id").OfType("INT4", int64(0))
	column2 := sqlmock.NewColumn("value").OfType("FLOAT64", float64(0.0))
	column3 := sqlmock.NewColumn("text").OfType("VARCHAR", "")
	column4 := sqlmock.NewColumn("valid").OfType("BOOL", false)

	// columns of different types
	rows := mock.NewRowsWithColumnDefinition(column1, column2, column3, column4)

	rows.AddRow(1, 1.2, "foo", true)
	rows.AddRow(2, 1.5, "bar", false)

	// expected query performed by tested function
	mock.ExpectQuery("SELECT \\* FROM report WHERE org_id IN \\('1','42'\\)").WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, config)

	// call the tested method, table name must be in predefined list
	values, err := storage.ReadTable("report", NoLimits)
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if len(values) != 2 {
		t.Errorf("wrong number records returned: %d", len(values))
	}

	assert.Equal(t, values[0]["id"], int64(1))
	assert.Equal(t, values[1]["id"], int64(2))
	assert.Equal(t, values[0]["text"], "foo")
	assert.Equal(t, values[1]["text"], "bar")

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadTable with selective export enabled and LIMIT enabled
func TestReadTableWithSelectiveExportAllowedTableWithLimits(t *testing.T) {
	config := &testConfig
	config.EnableOrgIDFiltering = true
	config.OrganizationsToExport = []string{"1", "42"}

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	column1 := sqlmock.NewColumn("id").OfType("INT4", int64(0))
	column2 := sqlmock.NewColumn("value").OfType("FLOAT64", float64(0.0))
	column3 := sqlmock.NewColumn("text").OfType("VARCHAR", "")
	column4 := sqlmock.NewColumn("valid").OfType("BOOL", false)

	// columns of different types
	rows := mock.NewRowsWithColumnDefinition(column1, column2, column3, column4)

	rows.AddRow(1, 1.2, "foo", true)
	rows.AddRow(2, 1.5, "bar", false)

	// expected query performed by tested function
	mock.ExpectQuery("SELECT \\* FROM report WHERE org_id IN \\('1','42'\\) LIMIT 2").WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, config)

	// call the tested method, table name must be in predefined list, limit == 2
	values, err := storage.ReadTable("report", 2)
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if len(values) != 2 {
		t.Errorf("wrong number records returned: %d", len(values))
	}

	assert.Equal(t, values[0]["id"], int64(1))
	assert.Equal(t, values[1]["id"], int64(2))
	assert.Equal(t, values[0]["text"], "foo")
	assert.Equal(t, values[1]["text"], "bar")

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadTable in case of error
func TestReadTableOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	mock.ExpectQuery(readTableQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	_, err := storage.ReadTable("table_name", NoLimits)
	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function RetrieveColumnTypes
func TestRetrieveColumnTypes(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	column1 := sqlmock.NewColumn("id").OfType("INT4", int64(0))
	column2 := sqlmock.NewColumn("value").OfType("FLOAT64", float64(0.0))
	column3 := sqlmock.NewColumn("text").OfType("VARCHAR", "")

	// columns of different types
	rows := mock.NewRowsWithColumnDefinition(column1, column2, column3)

	rows.AddRow(1, 1.2, "foo")
	rows.AddRow(2, 1.5, "bar")
	rows.AddRow(3, 2.0, "baz")

	// expected query performed by tested function
	mock.ExpectQuery(readColumnTypesQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	types, err := storage.RetrieveColumnTypes("table_name")
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if len(types) != 3 {
		t.Errorf("wrong number of types returned: %d", len(types))
	}

	assert.Equal(t, types[0].Name(), "id")
	assert.Equal(t, types[1].Name(), "value")
	assert.Equal(t, types[2].Name(), "text")

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function RetrieveColumnTypes
func TestRetrieveColumnTypesOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	mock.ExpectQuery(readColumnTypesQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	_, err := storage.RetrieveColumnTypes("table_name")

	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function StoreTableIntoFile
func TestStoreTableIntoFile(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	column1 := sqlmock.NewColumn("id").OfType("INT4", int64(0))
	column2 := sqlmock.NewColumn("value").OfType("FLOAT64", float64(0.0))
	column3 := sqlmock.NewColumn("text").OfType("VARCHAR", "")

	// columns of different types
	rows := mock.NewRowsWithColumnDefinition(column1, column2, column3)

	rows.AddRow(1, 1.2, "foo")
	rows.AddRow(2, 1.5, "bar")
	rows.AddRow(3, 2.0, "baz")

	// expected query performed by tested function
	mock.ExpectQuery(readColumnTypesQuery).WillReturnRows(rows)

	// expected query performed by tested function
	expectedQuery2 := "SELECT \\* FROM table_name"

	mock.ExpectQuery(expectedQuery2).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	err := storage.StoreTableIntoFile("table_name", NoLimits)
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)

	// check generated file
	content, err := os.ReadFile("table_name.csv")
	if err != nil {
		t.Errorf("error during reading file %s", err)
	}

	expected := `id,value,text
1,1.2,foo
2,1.5,bar
3,2,baz
`
	assert.Equal(t, expected, string(content))
}

// check the function StoreTableIntoFile
func TestStoreTableIntoFileWithLimit(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	column1 := sqlmock.NewColumn("id").OfType("INT4", int64(0))
	column2 := sqlmock.NewColumn("value").OfType("FLOAT64", float64(0.0))
	column3 := sqlmock.NewColumn("text").OfType("VARCHAR", "")

	// columns of different types
	rows := mock.NewRowsWithColumnDefinition(column1, column2, column3)

	rows.AddRow(1, 1.2, "foo")
	rows.AddRow(2, 1.5, "bar")

	// expected query performed by tested function
	mock.ExpectQuery(readColumnTypesQuery).WillReturnRows(rows)

	// expected query performed by tested function
	expectedQuery2 := "SELECT \\* FROM table_name LIMIT 2"

	mock.ExpectQuery(expectedQuery2).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	err := storage.StoreTableIntoFile("table_name", 2)
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)

	// check generated file
	content, err := os.ReadFile("table_name.csv")
	if err != nil {
		t.Errorf("error during reading file %s", err)
	}

	expected := `id,value,text
1,1.2,foo
2,1.5,bar
`
	assert.Equal(t, expected, string(content))
}

// check the function ReadDisabledRules
func TestReadDisabledRules(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"rule", "count"})
	rows.AddRow("rule1", 1)
	rows.AddRow("rule2", 2)
	rows.AddRow("rule3", 3)

	// expected query performed by tested function
	mock.ExpectQuery(readDisabledRulesQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	results, err := storage.ReadDisabledRules()
	if err != nil {
		t.Errorf("error was not expected %s", err)
	}

	if len(results) != 3 {
		t.Errorf("wrong number records returned: %d", len(results))
	}

	// check the list of returned records
	assert.Equal(t, "rule1", results[0].Rule)
	assert.Equal(t, "rule2", results[1].Rule)
	assert.Equal(t, "rule3", results[2].Rule)

	assert.Equal(t, 1, results[0].Count)
	assert.Equal(t, 2, results[1].Count)
	assert.Equal(t, 3, results[2].Count)

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadDisabledRules
func TestReadDisabledRulesOnError(t *testing.T) {
	// error to be thrown
	mockedError := errors.New("mocked error")

	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// expected query performed by tested function
	mock.ExpectQuery(readDisabledRulesQuery).WillReturnError(mockedError)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	_, err := storage.ReadDisabledRules()

	if err != mockedError {
		t.Errorf("different error was returned: %v", err)
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}

// check the function ReadDisabledRules
func TestReadDisabledRulesScanError(t *testing.T) {
	// prepare new mocked connection to database
	connection, mock := mustCreateMockConnection(t)

	// prepare mocked result for SQL query
	rows := sqlmock.NewRows([]string{"rule", "count"})
	rows.AddRow("rule1", "not count")
	rows.AddRow("rule2", "not count")
	rows.AddRow("rule3", "not count")

	// expected query performed by tested function
	expectedQuery := "SELECT rule_id, count\\(rule_id\\) AS rule_count FROM rule_disable GROUP BY rule_id HAVING count\\(rule_id\\)\\>1 ORDER BY rule_count DESC;"
	mock.ExpectQuery(expectedQuery).WillReturnRows(rows)
	mock.ExpectClose()

	// prepare connection to mocked database
	storage := main.NewFromConnection(connection, main.DBDriverPostgres, &testConfig)

	// call the tested method
	_, err := storage.ReadDisabledRules()
	if err == nil {
		t.Errorf("error was expected")
	}

	// connection to mocked DB needs to be closed properly
	checkConnectionClose(t, connection)

	// check if all expectations were met
	checkAllExpectations(t, mock)
}
