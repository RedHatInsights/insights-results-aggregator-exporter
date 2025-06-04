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

package main

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-exporter
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/storage.html

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"database/sql"

	_ "github.com/lib/pq"           // PostgreSQL database driver
	_ "github.com/mattn/go-sqlite3" // SQLite database driver

	"github.com/rs/zerolog/log"

	"github.com/minio/minio-go/v7"
)

// Driver types
const (
	// DBDriverSQLite3 shows that db driver is sqlite
	DBDriverSQLite3 DBDriver = iota
	// DBDriverPostgres shows that db driver is postgres
	DBDriverPostgres
)

// Error messages for all database-relevant errors
const (
	unableToCloseDBRowsHandle   = "Unable to close the DB rows handle"
	sqlStatementExecutionError  = "SQL statement execution error"
	unableToRetrieveColumnTypes = "Unable to retrieve column types"
	readTableContentFailed      = "Read table content failed"
	readListOfRecordsFailed     = "Unable to read list of records"
	writeOneRowToCSV            = "Write one row to CSV"
	sqlStatementExecuted        = "SQL statement"
)

// SQL statements
const (
	// Select all public tables from open database
	selectListOfTablesInPostgres = `
           SELECT tablename
             FROM pg_catalog.pg_tables
            WHERE schemaname != 'information_schema'
              AND schemaname != 'pg_catalog'
              AND schemaname != 'dvo';
   `

	selectListOfTablesInSQLite = `
           SELECT name FROM sqlite_master
            WHERE type IN ('table','view')
              AND name NOT LIKE 'sqlite_%'
            ORDER BY 1;
   `

	selectDisabledRules = `
           SELECT rule_id, count(rule_id) AS rule_count
	     FROM rule_disable
	    GROUP BY rule_id
	   HAVING count(rule_id)>1
	    ORDER BY rule_count DESC;
   `
)

// CSVFileExtension is common extension used for files with comma-separated records
const CSVFileExtension = ".csv"

var (
	selectiveExportAllowedTables = []TableName{
		"report",
		"recommendation",
		"rule_hit",
		"rule_disable",
		"rule_toggle",
		"cluster_rule_user_feedback",
		"cluster_user_rule_disable_feedback",
		"advisor_ratings",
	}

	whereOrgIDFilter = " WHERE org_id IN ('%v')"
)

// Storage represents an interface to almost any database or storage system
type Storage interface {
	Close() error

	ReadListOfTables() ([]TableName, error)
	ReadTable(tableName string, limit int) error
}

// DBStorage is an implementation of Storage interface that use selected SQL like database
// like SQLite, PostgreSQL, MariaDB, RDS etc. That implementation is based on the standard
// sql package. It is possible to configure connection via Configuration structure.
// SQLQueriesLog is log for sql queries, default is nil which means nothing is logged
type DBStorage struct {
	connection   *sql.DB
	dbDriverType DBDriver
	config       *StorageConfiguration
}

// NewStorage function creates and initializes a new instance of Storage interface
func NewStorage(configuration *StorageConfiguration) (*DBStorage, error) {
	log.Info().Msg("Initializing connection to storage")

	// initialize database driver
	driverType, driverName, dataSource, err := initAndGetDriver(configuration)
	if err != nil {
		log.Error().Err(err).Msg("Unsupported driver")
		return nil, err
	}

	// print info about initialized driver
	log.Info().
		Str("driver", driverName).
		Str("datasource", dataSource).
		Msg("Making connection to data storage")

	// prepare connection to database
	connection, err := sql.Open(driverName, dataSource)
	if err != nil {
		log.Error().Err(err).Msg("Can not connect to data storage")
		return nil, err
	}

	log.Info().Msg("Connection to storage established")
	return NewFromConnection(connection, driverType, configuration), nil
}

// NewFromConnection function creates and initializes a new instance of Storage interface from prepared connection
func NewFromConnection(connection *sql.DB, dbDriverType DBDriver, config *StorageConfiguration) *DBStorage {
	return &DBStorage{
		connection:   connection,
		dbDriverType: dbDriverType,
		config:       config,
	}
}

// initAndGetDriver initializes driver(with logs if logSQLQueries is true),
// checks if it's supported and returns driver type, driver name, dataSource and error
func initAndGetDriver(configuration *StorageConfiguration) (driverType DBDriver, driverName, dataSource string, err error) {
	driverName = configuration.Driver

	switch driverName {
	case "sqlite3":
		driverType = DBDriverSQLite3
		dataSource = configuration.SQLiteDataSource
	case "postgres":
		driverType = DBDriverPostgres
		dataSource = fmt.Sprintf(
			"postgresql://%v:%v@%v:%v/%v?%v",
			configuration.PGUsername,
			configuration.PGPassword,
			configuration.PGHost,
			configuration.PGPort,
			configuration.PGDBName,
			configuration.PGParams,
		)
	default:
		err = fmt.Errorf("driver %v is not supported", driverName)
		return
	}

	return
}

// Close method closes the connection to database. Needs to be called at the
// end of application lifecycle.
func (storage DBStorage) Close() error {
	log.Info().Msg("Closing connection to data storage")

	// try to close the connection
	if storage.connection != nil {
		err := storage.connection.Close()
		if err != nil {
			log.Error().Err(err).Msg("Can not close connection to data storage")
			return err
		}
	}
	return nil
}

// ReadListOfTables method reads names of all public tables stored in opened
// database.
func (storage DBStorage) ReadListOfTables() ([]TableName, error) {
	// slice to make list of tables
	var tableList = make([]TableName, 0)

	var selectListOfTables string
	switch storage.dbDriverType {
	case DBDriverSQLite3:
		selectListOfTables = selectListOfTablesInSQLite
	case DBDriverPostgres:
		selectListOfTables = selectListOfTablesInPostgres
	default:
		return tableList, fmt.Errorf("Invalid DB driver")
	}

	rows, err := storage.connection.Query(selectListOfTables)
	if err != nil {
		return tableList, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Error().Err(err).Msg(unableToCloseDBRowsHandle)
		}
	}()

	// read all table names
	for rows.Next() {
		var tableName TableName

		err := rows.Scan(&tableName)
		if err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg(unableToCloseDBRowsHandle)
			}
			return tableList, err
		}
		tableList = append(tableList, tableName)
	}

	return tableList, nil
}

// logColumnTypes is helper function to print column names and types for
// selected table.
func logColumnTypes(tableName TableName, columnTypes []*sql.ColumnType) {
	log.Info().
		Str("table columns", string(tableName)).
		Int("columns", len(columnTypes)).
		Msg("table metadata")

	for i, columnType := range columnTypes {
		log.Info().
			Str("name", columnType.Name()).
			Str("type", columnType.DatabaseTypeName()).
			Int("column", i+1).Msg("column type")
	}
}

// logRecordCount is helper function to print number of records stored in
// given database table.
func logRecordCount(tableName TableName, count int) {
	log.Info().
		Str("table name", string(tableName)).
		Int("record count", count).
		Msg("records in table")
}

// fillInScanArgs prepares arguments for the Scan method to retrieve row from
// selected table.
//
// Based on:
// https://stackoverflow.com/questions/42774467/how-to-convert-sql-rows-to-typed-json-in-golang#60386531
func fillInScanArgs(columnTypes []*sql.ColumnType) []interface{} {
	count := len(columnTypes)

	// data structure to scan one row
	scanArgs := make([]interface{}, count)

	for i, v := range columnTypes {
		switch v.DatabaseTypeName() {
		case "VARCHAR", "TEXT", "UUID", "TIMESTAMP":
			scanArgs[i] = new(sql.NullString)
		case "BOOL":
			scanArgs[i] = new(sql.NullBool)
		case "INT4":
			scanArgs[i] = new(sql.NullInt64)
		default:
			scanArgs[i] = new(sql.NullString)
		}
	}

	return scanArgs
}

// fillInMasterData fills the structure by row data read from database from
// selected table.
//
// Based on:
// https://stackoverflow.com/questions/42774467/how-to-convert-sql-rows-to-typed-json-in-golang#60386531
func fillInMasterData(columnTypes []*sql.ColumnType, scanArgs []interface{}) map[string]interface{} {
	masterData := map[string]interface{}{}

	// fill-in the data structure by row data
	for i, v := range columnTypes {
		if z, ok := (scanArgs[i]).(*sql.NullBool); ok {
			masterData[v.Name()] = z.Bool
			continue
		}

		if z, ok := (scanArgs[i]).(*sql.NullString); ok {
			masterData[v.Name()] = z.String
			continue
		}

		if z, ok := (scanArgs[i]).(*sql.NullInt64); ok {
			masterData[v.Name()] = z.Int64
			continue
		}

		if z, ok := (scanArgs[i]).(*sql.NullFloat64); ok {
			masterData[v.Name()] = z.Float64
			continue
		}

		if z, ok := (scanArgs[i]).(*sql.NullInt32); ok {
			masterData[v.Name()] = z.Int32
			continue
		}

		masterData[v.Name()] = scanArgs[i]
	}

	return masterData
}

// select1FromTable is helper function to construct query to database - read
// one record from given table.
func select1FromTable(tableName TableName) string {
	// it is not possible to use parameter for table name or a key
	// disable "G201 (CWE-89): SQL string formatting (Confidence: HIGH, Severity: MEDIUM)"
	// #nosec G201
	return fmt.Sprintf("SELECT * FROM %s LIMIT 1", string(tableName))
}

// selectCountFromTable is helper function to construct query to database -
// read number of records in table.
func selectCountFromTable(tableName TableName) string {
	// it is not possible to use parameter for table name or a key
	// disable "G201 (CWE-89): SQL string formatting (Confidence: HIGH, Severity: MEDIUM)"
	// #nosec G201
	return fmt.Sprintf("SELECT count(*) FROM %s", string(tableName))
}

func selectAllFromTable(tableName TableName) string {
	// it is not possible to use parameter for table name or a key
	// disable "G201 (CWE-89): SQL string formatting (Confidence: HIGH, Severity: MEDIUM)"
	// #nosec G201
	return fmt.Sprintf("SELECT * FROM %s", string(tableName))
}

// ReadTable method reads the whole content of selected table.
func (storage DBStorage) ReadTable(tableName TableName, limit int) ([]M, error) {
	sqlStatement := selectAllFromTable(tableName)

	storage.applySelectiveExport(&sqlStatement, tableName)

	if limit > 0 {
		sqlStatement += fmt.Sprintf(" LIMIT %d", limit)
	}

	log.Info().Str(sqlStatementExecuted, sqlStatement).Msg("Performing")

	rows, err := storage.connection.Query(sqlStatement)
	if err != nil {
		log.Error().Err(err).Str(sqlStatementExecuted, sqlStatement).Msg(sqlStatementExecutionError)
		return nil, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Error().Err(err).Msg(unableToCloseDBRowsHandle)
		}
	}()

	// try to retrieve column types
	columnTypes, err := rows.ColumnTypes()

	if err != nil {
		log.Error().Err(err).Msg(unableToRetrieveColumnTypes)
		return nil, err
	}

	logColumnTypes(tableName, columnTypes)

	// prepare data structure to hold raw values
	var finalRows []M

	// read table row by row
	for rows.Next() {
		// prepare arguments for the Scan method to retrieve row from
		// selected table.
		scanArgs := fillInScanArgs(columnTypes)

		// do the actual scan of row read from database
		err := rows.Scan(scanArgs...)

		if err != nil {
			log.Error().Err(err).Msg("Unable to scan row")
			return nil, err
		}

		// it is now needed to check each element of values for nil
		// then to use type introspection and type assertion to be
		// able to fetch the column into a typed variable if needed
		masterData := fillInMasterData(columnTypes, scanArgs)

		// TODO: make the export part there
		// println(masterData)
		finalRows = append(finalRows, masterData)
	}
	return finalRows, nil
}

// StoreTable function stores specified table into S3/Minio
func (storage DBStorage) StoreTable(ctx context.Context,
	minioClient *minio.Client, bucketName, prefix string, tableName TableName,
	limit int) error {
	columnTypes, err := storage.RetrieveColumnTypes(tableName)
	if err != nil {
		return err
	}

	colNames := getColumnNames(columnTypes)

	buffer := new(bytes.Buffer)

	// initialize CSV writer
	writer := csv.NewWriter(buffer)

	err = writeColumnNames(writer, colNames)
	if err != nil {
		return err
	}

	err = storage.WriteTableContent(writer, tableName, colNames, limit)
	if err != nil {
		return err
	}

	writer.Flush()

	reader := io.Reader(buffer)

	// Compute exact object size instead of using default value -1
	//
	// Warning: possible problems with large tables and 32bit architecture
	// Warning: passing -1 will allocate a large amount of memory
	//
	// Previous warning taken from:
	// https://docs.min.io/docs/golang-client-api-reference#PutObject
	size := buffer.Len()

	options := minio.PutObjectOptions{ContentType: "text/csv"}
	objectName := setObjectPrefix(prefix, string(tableName)) + CSVFileExtension
	_, err = minioClient.PutObject(ctx, bucketName, objectName, reader, int64(size), options)
	if err != nil {
		return err
	}

	// reset buffer before it will be garbage collected
	buffer.Reset()
	return nil
}

// StoreTableIntoFile function stores specified table into selected file
func (storage DBStorage) StoreTableIntoFile(tableName TableName,
	limit int) error {
	columnTypes, err := storage.RetrieveColumnTypes(tableName)
	if err != nil {
		return err
	}

	colNames := getColumnNames(columnTypes)

	fileName := string(tableName) + CSVFileExtension

	// open new CSV file to be filled in
	// disable "G304 (CWE-22): Potential file inclusion via variable"
	fout, err := os.Create(fileName) // #nosec G304
	if err != nil {
		return err
	}

	// initialize CSV writer
	writer := csv.NewWriter(fout)

	err = writeColumnNames(writer, colNames)
	if err != nil {
		return err
	}

	err = storage.WriteTableContent(writer, tableName, colNames, limit)
	if err != nil {
		return err
	}

	writer.Flush()

	// check for any error during export to CSV
	err = writer.Error()
	if err != nil {
		return err
	}

	// close the file and check if close operation was ok
	err = fout.Close()
	if err != nil {
		return err
	}

	return nil
}

// ReadRecordsCount method reads number of records stored in given database
// table.
func (storage DBStorage) ReadRecordsCount(tableName TableName) (int, error) {
	sqlStatement := selectCountFromTable(tableName)

	storage.applySelectiveExport(&sqlStatement, tableName)

	// try to query DB
	row := storage.connection.QueryRow(sqlStatement)

	var count int

	err := row.Scan(&count)
	if err != nil {
		return -1, err
	}

	// everything seems to be ok
	logRecordCount(tableName, count)
	return count, nil
}

// RetrieveColumnTypes read column types from given table
func (storage DBStorage) RetrieveColumnTypes(tableName TableName) ([]*sql.ColumnType, error) {
	sqlStatement := select1FromTable(tableName)

	// try to query DB
	rows, err := storage.connection.Query(sqlStatement)
	if err != nil {
		log.Error().Err(err).Str(sqlStatementExecuted, sqlStatement).Msg(sqlStatementExecutionError)
		return nil, err
	}

	// try to retrieve column types
	columnTypes, err := rows.ColumnTypes()
	if err != nil {
		log.Error().Err(err).Msg(unableToRetrieveColumnTypes)
		return nil, err
	}

	// close query
	err = rows.Close()
	if err != nil {
		log.Error().Err(err).Msg(unableToCloseDBRowsHandle)
		return nil, err
	}

	// everything seems to be ok
	logColumnTypes(tableName, columnTypes)
	return columnTypes, nil
}

// WriteTableContent method writes content of whole table into given CSV
// writera (may be file or S3 bucke)
func (storage DBStorage) WriteTableContent(writer *csv.Writer,
	tableName TableName, colNames []string, limit int) error {
	// now we know column types, time to perform export
	finalRows, err := storage.ReadTable(tableName, limit)
	if err != nil {
		log.Error().Err(err).Msg(readTableContentFailed)
		return err
	}

	for _, finalRow := range finalRows {
		var columns []string
		for _, colName := range colNames {
			value := finalRow[colName]
			str := fmt.Sprintf("%v", value)
			columns = append(columns, str)
		}
		err = writer.Write(columns)
		if err != nil {
			log.Error().Err(err).Msg(writeOneRowToCSV)
			return err
		}
	}
	return nil
}

// StoreTableMetadataIntoFile method stores metadata about given tables into
// file.
func (storage DBStorage) StoreTableMetadataIntoFile(fileName string, tableNames []TableName) error {
	// open new CSV file to be filled in
	// disable "G304 (CWE-22): Potential file inclusion via variable"
	fout, err := os.Create(fileName) // #nosec G304
	if err != nil {
		return err
	}

	err = TableMetadataToCSV(fout, tableNames, storage)
	if err != nil {
		// logging has been performed already
		return err
	}

	// initialize CSV writer
	writer := csv.NewWriter(fout)

	// check for any error during export to CSV
	err = writer.Error()
	if err != nil {
		return err
	}

	// close the file and check if close operation was ok
	err = fout.Close()
	if err != nil {
		return err
	}

	return nil
}

// StoreTableMetadataIntoS3 method stores metadata about given tables into
// S3 or Minio.
func (storage DBStorage) StoreTableMetadataIntoS3(ctx context.Context,
	minioClient *minio.Client, bucketName string, objectName string,
	tableNames []TableName) error {
	buffer := new(bytes.Buffer)

	err := TableMetadataToCSV(buffer, tableNames, storage)
	if err != nil {
		// logging has been performed already
		return err
	}

	// write CSV data into S3 bucket or Minio bucket
	reader := io.Reader(buffer)

	options := minio.PutObjectOptions{ContentType: "text/csv"}
	_, err = minioClient.PutObject(ctx, bucketName, objectName, reader, -1, options)
	if err != nil {
		return err
	}

	// everything look ok
	return nil
}

func getColumnNames(columnTypes []*sql.ColumnType) []string {
	var colNames []string
	for _, columnType := range columnTypes {
		colNames = append(colNames, columnType.Name())
	}

	return colNames
}

func writeColumnNames(writer *csv.Writer, colNames []string) error {
	err := writer.Write(colNames)
	if err != nil {
		log.Error().Err(err).Msg("Write column names to CSV")
		return err
	}
	return nil
}

// ReadDisabledRules method reads rules disabled by more than one user
func (storage DBStorage) ReadDisabledRules() ([]DisabledRuleInfo, error) {
	// slice to make list of disabled rule
	var disabledRulesInfo = make([]DisabledRuleInfo, 0)

	rows, err := storage.connection.Query(selectDisabledRules)
	if err != nil {
		return disabledRulesInfo, err
	}

	defer func() {
		err := rows.Close()
		if err != nil {
			log.Error().Err(err).Msg(unableToCloseDBRowsHandle)
		}
	}()

	// read all records
	for rows.Next() {
		var disabledRuleInfo DisabledRuleInfo

		err := rows.Scan(&disabledRuleInfo.Rule, &disabledRuleInfo.Count)
		if err != nil {
			if closeErr := rows.Close(); closeErr != nil {
				log.Error().Err(closeErr).Msg(unableToCloseDBRowsHandle)
			}
			return disabledRulesInfo, err
		}
		disabledRulesInfo = append(disabledRulesInfo, disabledRuleInfo)
	}

	return disabledRulesInfo, nil
}

// check whether table is allowed to be exported selectively by org_id
func selectiveExportAllowed(tablename TableName) bool {
	for i := range selectiveExportAllowedTables {
		if tablename == selectiveExportAllowedTables[i] {
			return true
		}
	}
	return false
}

func (storage DBStorage) applySelectiveExport(sqlStatement *string, tablename TableName) {
	if storage.config.EnableOrgIDFiltering && selectiveExportAllowed(tablename) {
		*sqlStatement += fmt.Sprintf(whereOrgIDFilter, strings.Join(storage.config.OrganizationsToExport, "','"))
	}
}
