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
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/exporter.html

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Messages
const (
	versionMessage         = "Insights Results Aggregator Exporter version 1.0"
	authorsMessage         = "Pavel Tisnovsky, Red Hat Inc."
	operationFailedMessage = "Operation failed"
	listOfTablesMsg        = "List of tables"
	tableNameMsg           = "Table name"
	tableIsIgnored         = "Table is ignored, skipping export"
)

// Exit codes
const (
	// ExitStatusOK means that the tool finished with success
	ExitStatusOK = iota

	// ExitStatusLoggingError is returned in case of any logging initialization
	// error
	ExitStatusLoggingError

	// ExitStatusStorageError is returned in case of any consumer-related
	// error
	ExitStatusStorageError

	// ExitStatusS3Error is returned in case of any error related with
	// S3/Minio connection
	ExitStatusS3Error

	// ExitStatusConfigurationError is returned in case user provide wrong
	// configuration on command line or in configuration file
	ExitStatusConfigurationError

	// ExitStatusIOError is returned in case of any I/O error (export data
	// into file failed etc.)
	ExitStatusIOError
)

const (
	configFileEnvVariableName = "INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE"
	defaultConfigFileName     = "config"
)

// output files or objects containing metadata
const (
	listOfTables  = "_tables.csv"
	metadataTable = "_metadata.csv"
	disabledRules = "_disabled_rules.csv"
	logFile       = "_logs.txt"
)

// messages
const (
	readDisabledRulesInfoFailed      = "Read disabled rules info failed"
	storeDisabledRulesIntoFileFailed = "Store disabled rules into file failed"
	readingListOfTables              = "Reading list of tables"
	exportingDisabledRules           = "Exporting disabled rules"
	closingConnectionToStorage       = "Closing connection to storage"
	exportingTables                  = "Exporting tables"
	exportingTable                   = "Exporting table"
	exportingMetadata                = "Exporting metadata"
	unknownOutputType                = "Unknown output type: %s"
)

// flags
const (
	s3Output   = "S3"
	fileOutput = "file"
)

// showVersion function displays version information.
func showVersion() {
	fmt.Println(versionMessage)
}

// showAuthors function displays information about authors.
func showAuthors() {
	fmt.Println(authorsMessage)
}

// showConfiguration function displays actual configuration.
func showConfiguration(config *ConfigStruct) {
	storageConfig := GetStorageConfiguration(config)
	log.Info().
		Str("Driver", storageConfig.Driver).
		Str("DB Name", storageConfig.PGDBName).
		Str("Username", storageConfig.PGUsername). // password is omitted on purpose
		Str("Host", storageConfig.PGHost).
		Int("DB Port", storageConfig.PGPort).
		Bool("LogSQLQueries", storageConfig.LogSQLQueries).
		Msg("Storage configuration")

	loggingConfig := GetLoggingConfiguration(config)
	log.Info().
		Str("Level", loggingConfig.LogLevel).
		Bool("Pretty colored debug logging", loggingConfig.Debug).
		Msg("Logging configuration")

	s3Configuration := GetS3Configuration(config)
	log.Info().
		Str("Type", s3Configuration.Type).
		Str("URL", s3Configuration.EndpointURL).
		Uint("S3 Port", s3Configuration.EndpointPort).
		Str("AccessKeyID", s3Configuration.AccessKeyID).
		Str("SecretAccessKey", s3Configuration.SecretAccessKey).
		Bool("Use SSL", s3Configuration.UseSSL).
		Str("Bucket name", s3Configuration.Bucket).
		Str("Bucket prefix", s3Configuration.Prefix).
		Msg("S3 configuration")
}

// constructIgnoredTablesMap helper function splits list of tables by comma and
// constructs a map from it where keys are taken from splitted string
func constructIgnoredTablesMap(input string) IgnoredTables {
	tables := strings.Split(input, ",")

	// prepare empty map with given capacity
	var m IgnoredTables = make(IgnoredTables, len(tables))

	// don't add empty string into a map
	if input == "" {
		return m
	}

	// put ignored tables into such map
	for _, table := range tables {
		m[table] = struct{}{}
	}

	return m
}

// performDataExport function exports all data into selected output
func performDataExport(configuration *ConfigStruct, cliFlags CliFlags, operationLogger zerolog.Logger) (int, error) {
	operationLogger.Info().Msg("Retrieving connection to storage")

	// prepare the storage
	storageConfiguration := GetStorageConfiguration(configuration)
	storage, err := NewStorage(&storageConfiguration)
	if err != nil {
		log.Err(err).Msg(operationFailedMessage)
		operationLogger.Err(err).Msg("Unable to retrieve connection to storage")
		return ExitStatusStorageError, err
	}

	ignoredTablesMap := constructIgnoredTablesMap(cliFlags.IgnoredTables)

	switch cliFlags.Output {
	case s3Output:
		return performDataExportToS3(configuration, storage,
			cliFlags.ExportMetadata, cliFlags.ExportDisabledRules,
			operationLogger, cliFlags.Limit, ignoredTablesMap)
	case fileOutput:
		return performDataExportToFiles(configuration, storage,
			cliFlags.ExportMetadata, cliFlags.ExportDisabledRules,
			operationLogger, cliFlags.Limit, ignoredTablesMap)
	default:
		err := fmt.Errorf(unknownOutputType, cliFlags.Output)
		operationLogger.Err(err).Msg("Wrong output type selected")
		return ExitStatusConfigurationError, err
	}
}

// performDataExportToS3 exports all tables and metadata info configured S3
// bucket
func performDataExportToS3(configuration *ConfigStruct,
	storage *DBStorage, exportMetadata bool,
	exportDisabledRules bool,
	operationLogger zerolog.Logger, limit int,
	ignoredTables IgnoredTables) (int, error) {

	operationLogger.Info().Msg("Exporting to S3")

	operationLogger.Info().Msg(readingListOfTables)

	minioClient, context, err := NewS3Connection(configuration)
	if err != nil {
		return ExitStatusS3Error, err
	}

	tableNames, err := storage.ReadListOfTables()
	if err != nil {
		log.Err(err).Msg(operationFailedMessage)
		operationLogger.Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	log.Info().Int("tables count", len(tableNames)).Msg(listOfTablesMsg)

	// log into terminal
	printTables(tableNames)

	s3config := GetS3Configuration(configuration)
	bucket, bucketPrefix := s3config.Bucket, s3config.Prefix
	log.Info().Str("bucket name", bucket).Msg("S3 bucket to write to")
	listOfTablesObject := setObjectPrefix(bucketPrefix, listOfTables)
	metadataTableObject := setObjectPrefix(bucketPrefix, metadataTable)

	if exportMetadata {
		operationLogger.Info().Msg(exportingMetadata)

		// export list of all tables into S3
		err = storeTableNames(context, minioClient,
			bucket, listOfTablesObject, tableNames)
		if err != nil {
			const msg = "Store table list to S3 failed"
			log.Err(err).Msg(msg)
			operationLogger.Err(err).Msg(msg)
			return ExitStatusStorageError, err
		}

		// export tables metadata into S3
		err = storage.StoreTableMetadataIntoS3(context, minioClient,
			bucket, metadataTableObject, tableNames)
		if err != nil {
			const msg = "Store tables metadata to S3 failed"
			log.Err(err).Msg(msg)
			return ExitStatusStorageError, err
		}
	}

	if exportDisabledRules {
		operationLogger.Info().Msg(exportingDisabledRules)

		// export rules disabled by more users into CSV file
		disabledRulesInfo, err := storage.ReadDisabledRules()
		if err != nil {
			log.Err(err).Msg(readDisabledRulesInfoFailed)
			operationLogger.Err(err).Msg(readDisabledRulesInfoFailed)
			return ExitStatusStorageError, err
		}

		// export list of disabled rules
		err = storeDisabledRulesIntoS3(context, minioClient, bucket,
			disabledRules, disabledRulesInfo)
		if err != nil {
			log.Err(err).Msg(storeDisabledRulesIntoFileFailed)
			operationLogger.Err(err).Msg(storeDisabledRulesIntoFileFailed)
			return ExitStatusIOError, err
		}
	}

	operationLogger.Info().Msg(exportingTables)

	// read content of all tables and perform export
	for _, tableName := range tableNames {
		// ignore table if specified by user
		if _, found := ignoredTables[string(tableName)]; found {
			operationLogger.Info().
				Str(tableNameMsg, string(tableName)).
				Msg(tableIsIgnored)
			continue
		}
		operationLogger.Info().
			Str(tableNameMsg, string(tableName)).
			Msg(exportingTable)
		err = storage.StoreTable(context, minioClient, bucket, bucketPrefix, tableName, limit)
		if err != nil {
			const msg = "Store table into S3 failed"
			log.Err(err).Str(tableNameMsg, string(tableName)).
				Msg(msg)
			operationLogger.Err(err).Str(tableNameMsg, string(tableName)).
				Msg(msg)
			return ExitStatusStorageError, err
		}
	}

	operationLogger.Info().Msg(closingConnectionToStorage)

	// we have finished, let's close the connection to database
	err = storage.Close()
	if err != nil {
		log.Err(err).Msg(operationFailedMessage)
		operationLogger.Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	// default exit value + no error
	return ExitStatusOK, nil
}

// performDataExportToFiles exports all tables and metadata info files
func performDataExportToFiles(configuration *ConfigStruct,
	storage *DBStorage, exportMetadata bool,
	exportDisabledRules bool,
	operationLogger zerolog.Logger, limit int,
	ignoredTables IgnoredTables) (int, error) {

	operationLogger.Info().Msg("Exporting to file")

	operationLogger.Info().Msg(readingListOfTables)

	tableNames, err := storage.ReadListOfTables()
	if err != nil {
		log.Err(err).Msg(operationFailedMessage)
		operationLogger.Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	log.Info().Int("count", len(tableNames)).Msg(listOfTablesMsg)

	// log into terminal
	printTables(tableNames)

	if exportMetadata {
		operationLogger.Info().Msg(exportingMetadata)

		// export list of all tables into CSV file
		err = storeTableNamesIntoFile(listOfTables, tableNames)
		if err != nil {
			const msg = "Store table list to file failed"
			log.Err(err).Msg(msg)
			operationLogger.Err(err).Msg(msg)
			return ExitStatusStorageError, err
		}

		// export tables metadata into CSV file
		err = storage.StoreTableMetadataIntoFile(metadataTable, tableNames)
		if err != nil {
			const msg = "Store tables metadata to file failed"
			log.Err(err).Msg(msg)
			operationLogger.Err(err).Msg(msg)
			return ExitStatusStorageError, err
		}
	}

	if exportDisabledRules {
		operationLogger.Info().Msg(exportingDisabledRules)

		// export rules disabled by more users into CSV file
		disabledRulesInfo, err := storage.ReadDisabledRules()
		if err != nil {
			log.Err(err).Msg(readDisabledRulesInfoFailed)
			operationLogger.Err(err).Msg(readDisabledRulesInfoFailed)
			return ExitStatusStorageError, err
		}

		// export list of disabled rules
		err = storeDisabledRulesIntoFile(disabledRules, disabledRulesInfo)
		if err != nil {
			log.Err(err).Msg(storeDisabledRulesIntoFileFailed)
			operationLogger.Err(err).Msg(storeDisabledRulesIntoFileFailed)
			return ExitStatusIOError, err
		}
	}

	operationLogger.Info().Msg(exportingTables)

	// read content of all tables and perform export
	for _, tableName := range tableNames {
		// ignore table if specified by user
		if _, found := ignoredTables[string(tableName)]; found {
			operationLogger.Info().
				Str(tableNameMsg, string(tableName)).
				Msg(tableIsIgnored)
			continue
		}
		operationLogger.Info().
			Str(tableNameMsg, string(tableName)).
			Msg(exportingTable)
		err = storage.StoreTableIntoFile(tableName, limit)
		if err != nil {
			const msg = "Store table into file failed"
			log.Err(err).Str(tableNameMsg, string(tableName)).
				Msg(msg)
			operationLogger.Err(err).Str(tableNameMsg, string(tableName)).
				Msg(msg)
			return ExitStatusStorageError, err
		}
	}

	operationLogger.Info().Msg(closingConnectionToStorage)

	// we have finished, let's close the connection to database
	err = storage.Close()
	if err != nil {
		log.Err(err).Msg(operationFailedMessage)
		operationLogger.Err(err).Msg(operationFailedMessage)
		return ExitStatusStorageError, err
	}

	// default exit value + no error
	return ExitStatusOK, nil
}

func printTables(tableNames []TableName) {
	for i, tableName := range tableNames {
		log.Info().Int("#", i+1).Str("table", string(tableName)).Msg("Table in database")
	}
}

// checkS3Connection checks if connection to S3 is possible
func checkS3Connection(configuration *ConfigStruct) (int, error) {
	log.Info().Msg("Checking connection to S3")
	minioClient, context, err := NewS3Connection(configuration)
	if err != nil {
		return ExitStatusS3Error, err
	}

	exists, err := s3BucketExists(context, minioClient, GetS3Configuration(configuration).Bucket)
	if err != nil {
		return ExitStatusS3Error, err
	}

	if !exists {
		log.Error().Msg("Can not find expected bucket")
	} else {
		log.Info().Msg("Bucket has been found")
	}

	log.Info().Msg("Connection to S3 seems to be ok")
	return ExitStatusOK, nil
}

func storeOpertionLogIntoS3(configuration *ConfigStruct,
	buffer bytes.Buffer) error {
	minioClient, context, err := NewS3Connection(configuration)
	if err != nil {
		return err
	}

	s3config := GetS3Configuration(configuration)
	bucketName, bucketPrefix := s3config.Bucket, s3config.Prefix
	logFileObject := setObjectPrefix(bucketPrefix, logFile)
	return storeBufferToS3(context, minioClient, bucketName, logFileObject, buffer)
}

// doSelectedOperation function perform operation selected on command line.
// When no operation is specified, the Notification writer service is started
// instead.
func doSelectedOperation(configuration *ConfigStruct, cliFlags CliFlags,
	operationLogger zerolog.Logger) (int, error) {
	switch {
	case cliFlags.ShowVersion:
		showVersion()
		return ExitStatusOK, nil
	case cliFlags.ShowAuthors:
		showAuthors()
		return ExitStatusOK, nil
	case cliFlags.ShowConfiguration:
		showConfiguration(configuration)
		return ExitStatusOK, nil
	case cliFlags.CheckS3Connection:
		return checkS3Connection(configuration)
	default:
		// default operation - data export
		return performDataExport(configuration, cliFlags, operationLogger)
	}
	// this can not happen: return ExitStatusOK, nil
}

func parseFlags() (cliFlags CliFlags) {
	// define and parse all command line options
	flag.BoolVar(&cliFlags.ShowVersion, "version", false, "show version")
	flag.BoolVar(&cliFlags.ShowAuthors, "authors", false, "show authors")
	flag.BoolVar(&cliFlags.ShowConfiguration, "show-configuration", false, "show configuration")
	flag.BoolVar(&cliFlags.PrintSummaryTable, "summary", false, "print summary table after export")
	flag.StringVar(&cliFlags.Output, "output", "S3", "output to: file, S3")
	flag.BoolVar(&cliFlags.ExportMetadata, "metadata", false, "export metadata")
	flag.BoolVar(&cliFlags.ExportDisabledRules, "disabled-by-more-users", false, "export rules disabled by more users")
	flag.BoolVar(&cliFlags.CheckS3Connection, "check-s3-connection", false, "check S3 connection and exit")
	flag.BoolVar(&cliFlags.ExportLog, "export-log", false, "export log")
	flag.IntVar(&cliFlags.Limit, "limit", -1, "limit number of exported records")
	flag.StringVar(&cliFlags.IgnoredTables, "ignore-tables", "", "comma-separated list of tables that will be ignored")

	// parse all command line flags
	flag.Parse()

	return
}

// DummyWriter satisfies Writer interface, but with noop write
type DummyWriter struct{}

// Write method satisfies noop io.Write
func (w DummyWriter) Write(p []byte) (n int, err error) {
	return 0, nil
}

// createOperationLog function constructs operation log instance
func createOperationLog(cliFlags CliFlags, buffer *bytes.Buffer) (zerolog.Logger, error) {
	dummyLogger := zerolog.New(DummyWriter{}).With().Logger()

	if cliFlags.ExportLog {
		switch cliFlags.Output {
		case s3Output:
			memoryLogger := zerolog.New(buffer).With().Logger()
			memoryLogger.Info().Msg("Memory logger initialized")
			return memoryLogger, nil
		case fileOutput:
			logFile, err := os.Create(logFile)
			if err != nil {
				return dummyLogger, err
			}
			fileLogger := zerolog.New(logFile).With().Logger()
			fileLogger.Info().Msg("File logger initialized")
			return fileLogger, nil
		default:
			return dummyLogger, fmt.Errorf(unknownOutputType, cliFlags.Output)
		}
	}

	return dummyLogger, nil

}

func setObjectPrefix(prefix, object string) string {
	if prefix != "" {
		return prefix + "/" + object
	}
	return object
}

func mainWithStatusCode() int {
	// parse all command line flags
	cliFlags := parseFlags()

	// config has exactly the same structure as *.toml file
	config, err := LoadConfiguration(configFileEnvVariableName, defaultConfigFileName)
	if err != nil {
		log.Err(err).Msg("Load configuration")
	}

	loggingCloser, err := InitLogging(&config)
	if err != nil {
		log.Err(err).Msg("Init logging")
		return ExitStatusLoggingError
	}

	defer loggingCloser()

	log.Error().Err(errors.New("test error")).Msg("this is a test error")

	var buffer bytes.Buffer
	operationLogger, err := createOperationLog(cliFlags, &buffer)
	if err != nil {
		log.Err(err).Msg("Create operation log")
		return ExitStatusIOError
	}

	// perform selected operation
	exitStatus, err := doSelectedOperation(&config, cliFlags, operationLogger)
	if err != nil {
		log.Err(err).Msg("Do selected operation")
		return exitStatus
	}

	if cliFlags.ExportLog && cliFlags.Output == s3Output {
		err := storeOpertionLogIntoS3(&config, buffer)
		if err != nil {
			log.Err(err).Msg("Storing log into S3 failed")
			return ExitStatusS3Error
		}
	}

	log.Debug().Msg("Finished")
	return ExitStatusOK
}

func main() {
	exitStatus := mainWithStatusCode()
	os.Exit(exitStatus)
}
