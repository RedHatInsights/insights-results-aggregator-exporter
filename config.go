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

// This source file contains definition of data type named ConfigStruct that
// represents configuration of Insights Results Aggregator Exporter. This source
// file also contains function named LoadConfiguration that can be used to load
// configuration from provided configuration file and/or from environment
// variables. Additionally several specific functions named
// GetStorageConfiguration and GetLoggingConfiguration
// are to be used to return specific configuration
// options.

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-exporter
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/config.html

// Default name of configuration file is config.toml
// It can be changed via environment variable INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE

// An example of configuration file that can be used in devel environment:
//
// [storage]
// db_driver = "postgres"
// pg_username = "postgres"
// pg_password = "postgres"
// pg_host = "localhost"
// pg_port = 5432
// pg_db_name = "aggregator"
// pg_params = "sslmode=disable"
//
// [s3]
// type = "minio"
// endpoint_url = "127.0.0.1"
// endpoint_port = 9000
// access_key_id = "foobar"
// secret_access_key = "foobar"
// use_ssl = false
// bucket = "test"
//
// [logging]
// debug = true
// log_level = ""
//
// Environment variables that can be used to override configuration file settings:
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__DB_DRIVER
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_USERNAME
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_PASSWORD
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_HOST
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_PORT
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_DB_NAME
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_PARAMS
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__TYPE
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__ENDPOINT_URL
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__ENDPOINT_PORT
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__ACCESS_KEY_ID
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__SECRET_ACCESS_KEY
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__USE_SSL
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__BUCKET
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__PREFIX
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__LOGGING__DEBUG
// INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__LOGGING__LOG_DEVEL

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"

	"path/filepath"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// Common constants used for during logging and error reporting
const (
	filenameAttribute               = "filename"
	parsingConfigurationFileMessage = "parsing configuration file"
)

// ConfigStruct is a structure holding the whole service configuration
type ConfigStruct struct {
	Storage StorageConfiguration `mapstructure:"storage" toml:"storage"`
	S3      S3Configuration      `mapstructure:"s3"      toml:"s3"`
	Logging LoggingConfiguration `mapstructure:"logging" toml:"logging"`
	Sentry  SentryConfiguration  `mapstructure:"sentry"  toml:"sentry"`
}

// LoggingConfiguration represents configuration for logging in general
type LoggingConfiguration struct {
	// Debug enables pretty colored logging
	Debug bool `mapstructure:"debug" toml:"debug"`

	// LogLevel sets logging level to show. Possible values are:
	// "debug"
	// "info"
	// "warn", "warning"
	// "error"
	// "fatal"
	//
	// logging level won't be changed if value is not one of listed above
	LogLevel string `mapstructure:"log_level" toml:"log_level"`

	// LoggingToCloudWatchEnabled enables logging to CloudWatch
	// (configuration for CloudWatch is in CloudWatchConfiguration)
	LoggingToCloudWatchEnabled bool `mapstructure:"logging_to_cloud_watch_enabled" toml:"logging_to_cloud_watch_enabled"`
}

// StorageConfiguration represents configuration of input data storage
// (database)
type StorageConfiguration struct {
	Driver           string `mapstructure:"db_driver"         toml:"db_driver"`
	SQLiteDataSource string `mapstructure:"sqlite_datasource" toml:"sqlite_datasource"`
	PGUsername       string `mapstructure:"pg_username"       toml:"pg_username"`
	PGPassword       string `mapstructure:"pg_password"       toml:"pg_password"`
	PGHost           string `mapstructure:"pg_host"           toml:"pg_host"`
	PGPort           int    `mapstructure:"pg_port"           toml:"pg_port"`
	PGDBName         string `mapstructure:"pg_db_name"        toml:"pg_db_name"`
	PGParams         string `mapstructure:"pg_params"         toml:"pg_params"`
	LogSQLQueries    bool   `mapstructure:"log_sql_queries"   toml:"log_sql_queries"`
}

// S3Configuration represents configuration of S3/Minio data storage
type S3Configuration struct {
	Type            string `mapstructure:"type"              toml:"type"`
	EndpointURL     string `mapstructure:"endpoint_url"      toml:"endpoint_url"`
	EndpointPort    uint   `mapstructure:"endpoint_port"     toml:"endpoint_port"`
	AccessKeyID     string `mapstructure:"access_key_id"     toml:"access_key_id"`
	SecretAccessKey string `mapstructure:"secret_access_key" toml:"secret_access_key"`
	UseSSL          bool   `mapstructure:"use_ssl"           toml:"use_ssl"`
	Bucket          string `mapstructure:"bucket"            toml:"bucket"`
	Prefix          string `mapstructure:"prefix"            toml:"prefix"`
}

// SentryConfiguration represents the configuration of Sentry logger
type SentryConfiguration struct {
	SentryDSN         string `mapstructure:"dsn" toml:"dsn"`
	SentryEnvironment string `mapstructure:"environment" toml:"environment"`
}

// LoadConfiguration function loads configuration from defaultConfigFile, file
// set in configFileEnvVariableName or from environment variables
func LoadConfiguration(configFileEnvVariableName, defaultConfigFile string) (ConfigStruct, error) {
	var config ConfigStruct

	// env. variable holding name of configuration file
	configFile, specified := os.LookupEnv(configFileEnvVariableName)
	if specified {
		log.Info().Str(filenameAttribute, configFile).Msg(parsingConfigurationFileMessage)
		// we need to separate the directory name and filename without
		// extension
		directory, basename := filepath.Split(configFile)
		file := strings.TrimSuffix(basename, filepath.Ext(basename))
		// parse the configuration
		viper.SetConfigName(file)
		viper.AddConfigPath(directory)
	} else {
		log.Info().Str(filenameAttribute, defaultConfigFile).Msg(parsingConfigurationFileMessage)
		// parse the configuration
		viper.SetConfigName(defaultConfigFile)
		viper.AddConfigPath(".")
	}

	// try to read the whole configuration
	err := viper.ReadInConfig()
	if _, isNotFoundError := err.(viper.ConfigFileNotFoundError); !specified && isNotFoundError {
		// If configuration file is not present (which might be correct
		// in some environment) we need to read configuration from
		// environment variables. The problem is that Viper is not
		// smart enough to understand the structure of config by
		// itself, so we need to read fake config file
		fakeTomlConfigWriter := new(bytes.Buffer)

		err := toml.NewEncoder(fakeTomlConfigWriter).Encode(config)
		if err != nil {
			return config, err
		}

		fakeTomlConfig := fakeTomlConfigWriter.String()

		viper.SetConfigType("toml")

		err = viper.ReadConfig(strings.NewReader(fakeTomlConfig))

		// check for error during parsing
		if err != nil {
			return config, err
		}
	} else if err != nil {
		// error is processed on caller side
		return config, fmt.Errorf("fatal error config file: %s", err)
	}

	// override config from env if there's variable in env

	const envPrefix = "INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_"

	viper.AutomaticEnv()
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "__"))

	// try to unmarshall configuration and check for (any) error
	err = viper.Unmarshal(&config)
	if err != nil {
		return config, err
	}

	// updated configuration by introducing Clowder-related things
	if err := updateConfigFromClowder(&config); err != nil {
		fmt.Println("Error loading clowder configuration")
		return config, err
	}

	return config, err
}

// GetStorageConfiguration function returns storage configuration
func GetStorageConfiguration(config *ConfigStruct) StorageConfiguration {
	return config.Storage
}

// GetLoggingConfiguration function returns logging configuration
func GetLoggingConfiguration(config *ConfigStruct) LoggingConfiguration {
	return config.Logging
}

// GetSentryConfiguration returns logging configuration
func GetSentryConfiguration(config *ConfigStruct) SentryConfiguration {
	return config.Sentry
}

// GetS3Configuration function returns S3/Minio configuration
func GetS3Configuration(config *ConfigStruct) S3Configuration {
	return config.S3
}

// updateConfigFromClowder function updates the current config with the values
// defined in clowder
func updateConfigFromClowder(c *ConfigStruct) error {
	if clowder.IsClowderEnabled() {
		// can not use Zerolog at this moment!
		fmt.Println("Clowder is enabled")

		if clowder.LoadedConfig.Database != nil {
			// get DB configuration from clowder
			c.Storage.PGDBName = clowder.LoadedConfig.Database.Name
			c.Storage.PGHost = clowder.LoadedConfig.Database.Hostname
			c.Storage.PGPort = clowder.LoadedConfig.Database.Port
			c.Storage.PGUsername = clowder.LoadedConfig.Database.Username
			c.Storage.PGPassword = clowder.LoadedConfig.Database.Password
		}

	} else {
		fmt.Println("Clowder is disabled")
	}

	return nil
}
