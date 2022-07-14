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
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/config_test.html

import (
	"os"

	"testing"

	clowder "github.com/redhatinsights/app-common-go/pkg/api/v1"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	main "github.com/RedHatInsights/insights-results-aggregator-exporter"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.WarnLevel)
}

// mustLoadConfiguration function loads configuration file or the actual test
// will fail
func mustLoadConfiguration(envVar string) {
	_, err := main.LoadConfiguration(envVar, "tests/config1")
	if err != nil {
		panic(err)
	}
}

// mustSetEnv function set specified environment variable or the actual test
// will fail
func mustSetEnv(t *testing.T, key, val string) {
	err := os.Setenv(key, val)
	assert.NoError(t, err)
	if err != nil {
		t.Fatal(err)
	}
}

// TestLoadDefaultConfiguration test loads a configuration file for testing
// with check that load was correct
func TestLoadDefaultConfiguration(t *testing.T) {
	os.Clearenv()
	mustLoadConfiguration("nonExistingEnvVar")
}

// TestLoadConfigurationFromEnvVariable tests loading the config. file for
// testing from an environment variable
func TestLoadConfigurationFromEnvVariable(t *testing.T) {
	os.Clearenv()

	mustSetEnv(t, "INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE", "tests/config2")
	mustLoadConfiguration("INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE")
}

// TestLoadConfigurationNonEnvVarUnknownConfigFile tests loading an unexisting
// config file when no environment variable is provided
func TestLoadConfigurationNonEnvVarUnknownConfigFile(t *testing.T) {
	_, err := main.LoadConfiguration("", "foobar")
	assert.Nil(t, err)
}

// TestLoadConfigurationBadConfigFile tests loading an unexisting config file when no environment variable is provided
func TestLoadConfigurationBadConfigFile(t *testing.T) {
	_, err := main.LoadConfiguration("", "tests/config3")
	assert.Contains(t, err.Error(), `fatal error config file: While parsing config:`)
}

// TestLoadingConfigurationEnvVariableBadValueNoDefaultConfig tests loading a non-existent configuration file set in environment
func TestLoadingConfigurationEnvVariableBadValueNoDefaultConfig(t *testing.T) {
	os.Clearenv()

	mustSetEnv(t, "INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE", "non existing file")

	_, err := main.LoadConfiguration("INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE", "")
	assert.Contains(t, err.Error(), `fatal error config file: Config File "non existing file" Not Found in`)
}

// TestLoadingConfigurationEnvVariableBadValueNoDefaultConfig tests that if env var is provided, it must point to a valid config file
func TestLoadingConfigurationEnvVariableBadValueDefaultConfigFailure(t *testing.T) {
	os.Clearenv()

	mustSetEnv(t, "INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE", "non existing file")

	_, err := main.LoadConfiguration("INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE", "tests/config1")
	assert.Contains(t, err.Error(), `fatal error config file: Config File "non existing file" Not Found in`)
}

// TestLoadStorageConfiguration tests loading the storage configuration sub-tree
func TestLoadStorageConfiguration(t *testing.T) {
	envVar := "INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE"
	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	storageCfg := main.GetStorageConfiguration(&config)

	assert.Equal(t, "sqlite3", storageCfg.Driver)
	assert.Equal(t, "user", storageCfg.PGUsername)
	assert.Equal(t, "password", storageCfg.PGPassword)
	assert.Equal(t, "localhost", storageCfg.PGHost)
	assert.Equal(t, 5432, storageCfg.PGPort)
	assert.Equal(t, "notifications", storageCfg.PGDBName)
	assert.Equal(t, "", storageCfg.PGParams)
	assert.Equal(t, true, storageCfg.LogSQLQueries)
}

// TestLoadLoggingConfiguration tests loading the logging configuration sub-tree
func TestLoadLoggingConfiguration(t *testing.T) {
	envVar := "INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE"
	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	loggingCfg := main.GetLoggingConfiguration(&config)

	assert.Equal(t, true, loggingCfg.Debug)
	assert.Equal(t, "", loggingCfg.LogLevel)
}

// TestLoadSentryConfiguration tests loading the sentry configuration sub-tree
func TestLoadSentryConfiguration(t *testing.T) {
	envVar := "INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE"
	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	sentryCfg := main.GetSentryConfiguration(&config)

	assert.Equal(t, "test_dsn", sentryCfg.SentryDSN)
	assert.Equal(t, "test_env", sentryCfg.SentryEnvironment)
}

// TestLoadS3Configuration tests loading the S3 configuration sub-tree
func TestLoadS3Configuration(t *testing.T) {
	envVar := "INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE"
	mustSetEnv(t, envVar, "tests/config2")
	config, err := main.LoadConfiguration(envVar, "")
	assert.Nil(t, err, "Failed loading configuration file from env var!")

	S3Cfg := main.GetS3Configuration(&config)

	assert.Equal(t, "minio", S3Cfg.Type)
	assert.Equal(t, false, S3Cfg.UseSSL)
	assert.Equal(t, "test_path", S3Cfg.Prefix)
}

// TestLoadConfigurationFromEnvVariableClowderEnabled tests loading the config.
// file for testing from an environment variable. Clowder config is enabled in
// this case.
func TestLoadConfigurationFromEnvVariableClowderEnabled(t *testing.T) {
	var testDB = "test_db"
	os.Clearenv()

	clowder.LoadedConfig = &clowder.AppConfig{
		Database: &clowder.DatabaseConfig{
			Name: testDB,
		},
	}
	mustSetEnv(t, "ACG_CONFIG", "to enable clowder")

	config, err := main.LoadConfiguration("INSIGHTS_RESULTS_AGGREGATOR_EXPORTER_CONFIG_FILE", "tests/config2")
	assert.NoError(t, err, "Failed loading configuration file")

	dbCfg := main.GetStorageConfiguration(&config)
	assert.Equal(t, testDB, dbCfg.PGDBName)
}
