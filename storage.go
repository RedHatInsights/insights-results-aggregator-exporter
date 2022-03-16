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

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"

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

// Error messages
const (
	unableToCloseDBRowsHandle   = "Unable to close the DB rows handle"
	sqlStatementExecutionError  = "SQL statement execution error"
	unableToRetrieveColumnTypes = "Unable to retrieve column types"
	readTableContentFailed      = "Read table content failed"
)

// SQL statements
const (
	// Select all public tables from open database
	selectListOfTables = `
           SELECT tablename
             FROM pg_catalog.pg_tables
            WHERE schemaname != 'information_schema'
              AND schemaname != 'pg_catalog';
   `
)

// Storage represents an interface to almost any database or storage system
type Storage interface {
	Close() error

	ReadListOfTables() ([]TableName, error)
	ReadTable(tableName string) error
}

// DBStorage is an implementation of Storage interface that use selected SQL like database
// like SQLite, PostgreSQL, MariaDB, RDS etc. That implementation is based on the standard
// sql package. It is possible to configure connection via Configuration structure.
// SQLQueriesLog is log for sql queries, default is nil which means nothing is logged
type DBStorage struct {
	connection   *sql.DB
	dbDriverType DBDriver
}

// NewStorage function creates and initializes a new instance of Storage interface
func NewStorage(configuration StorageConfiguration) (*DBStorage, error) {
	log.Info().Msg("Initializing connection to storage")

	driverType, driverName, dataSource, err := initAndGetDriver(configuration)
	if err != nil {
		log.Error().Err(err).Msg("Unsupported driver")
		return nil, err
	}

	log.Info().
		Str("driver", driverName).
		Str("datasource", dataSource).
		Msg("Making connection to data storage")

	// prepare connection
	connection, err := sql.Open(driverName, dataSource)
	if err != nil {
		log.Error().Err(err).Msg("Can not connect to data storage")
		return nil, err
	}

	log.Info().Msg("Connection to storage established")
	return NewFromConnection(connection, driverType), nil
}

// NewFromConnection function creates and initializes a new instance of Storage interface from prepared connection
func NewFromConnection(connection *sql.DB, dbDriverType DBDriver) *DBStorage {
	return &DBStorage{
		connection:   connection,
		dbDriverType: dbDriverType,
	}
}

// initAndGetDriver initializes driver(with logs if logSQLQueries is true),
// checks if it's supported and returns driver type, driver name, dataSource and error
func initAndGetDriver(configuration StorageConfiguration) (driverType DBDriver, driverName, dataSource string, err error) {
	driverName = configuration.Driver

	switch driverName {
	case "sqlite3":
		driverType = DBDriverSQLite3
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
	log.Info().Str("table columns", string(tableName)).Int("columns", len(columnTypes)).Msg("table metadata")

	for i, columnType := range columnTypes {
		log.Info().
			Str("name", columnType.Name()).
			Str("type", columnType.DatabaseTypeName()).
			Int("column", i+1).Msg("column type")
	}
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

func select1FromTable(tableName TableName) string {
	// it is not possible to use parameter for table name or a key
	// disable "G201 (CWE-89): SQL string formatting (Confidence: HIGH, Severity: MEDIUM)"
	// #nosec G201
	return fmt.Sprintf("SELECT * FROM %s LIMIT 1", string(tableName))
}

func selectAllFromTable(tableName TableName) string {
	// it is not possible to use parameter for table name or a key
	// disable "G201 (CWE-89): SQL string formatting (Confidence: HIGH, Severity: MEDIUM)"
	// #nosec G201
	return fmt.Sprintf("SELECT * FROM %s", string(tableName))
}

// ReadTable method reads the whole content of selected table.
func (storage DBStorage) ReadTable(tableName TableName) ([]M, error) {
	sqlStatement := selectAllFromTable(tableName)
	log.Info().Str("SQL statement", sqlStatement).Msg("Performing")

	rows, err := storage.connection.Query(sqlStatement)
	if err != nil {
		log.Error().Err(err).Msg(sqlStatementExecutionError)
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
// TODO: Really needs refactoring!!!
// TODO: refactor retrieving column types info function
func (storage DBStorage) StoreTable(ctx context.Context,
	minioClient *minio.Client, bucketName string, tableName TableName) error {
	sqlStatement := select1FromTable(tableName)

	rows, err := storage.connection.Query(sqlStatement)
	if err != nil {
		log.Error().Err(err).Msg(sqlStatementExecutionError)
		return err
	}

	// try to retrieve column types
	columnTypes, err := rows.ColumnTypes()

	if err != nil {
		log.Error().Err(err).Msg(unableToRetrieveColumnTypes)
		return err
	}

	err = rows.Close()
	if err != nil {
		log.Error().Err(err).Msg(unableToCloseDBRowsHandle)
	}

	logColumnTypes(tableName, columnTypes)
	buffer := new(bytes.Buffer)

	writer := csv.NewWriter(buffer)
	var colNames []string

	for _, columnType := range columnTypes {
		colNames = append(colNames, columnType.Name())
	}

	err = writer.Write(colNames)
	if err != nil {
		log.Error().Err(err).Msg("Write column names to CSV")
		return err
	}

	// now we know column types, time to perform export
	finalRows, err := storage.ReadTable(tableName)
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
			log.Error().Err(err).Msg("Write one row to CSV")
			return err
		}
	}

	writer.Flush()

	reader := io.Reader(buffer)

	options := minio.PutObjectOptions{ContentType: "text/csv"}
	_, err = minioClient.PutObject(ctx, bucketName, string(tableName), reader, -1, options)
	if err != nil {
		return err
	}
	return nil
}
