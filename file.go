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

package main

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-exporter
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/file.html

import (
	"encoding/csv"
	"os"

	"github.com/rs/zerolog/log"
)

// Messages
const (
	writeTableNameToCSV        = "Write table name to CSV"
	writeDisabledRuleInfoToCSV = "Write disabled rule info to CSV"
)

// storeTableNamesIntoFile function stores names of all tables into the
// specified file
func storeTableNamesIntoFile(fileName string, tableNames []TableName) error {
	// open new CSV file to be filled in

	// disable "G304 (CWE-22): Potential file inclusion via variable"
	fout, err := os.Create(fileName) // #nosec G304
	if err != nil {
		return err
	}

	// initialize CSV writer
	writer := csv.NewWriter(fout)
	var data = [][]string{{tableNameMsg}}

	// header
	err = writer.WriteAll(data)
	if err != nil {
		return err
	}

	// table names
	for _, tableName := range tableNames {
		err := writer.Write([]string{string(tableName)})
		if err != nil {
			log.Error().Err(err).Msg(writeTableNameToCSV)
		}
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

// storeDisabledRulesIntoFile function stores info about disabled rules into
// specified file
func storeDisabledRulesIntoFile(fileName string, disabledRulesInfo []DisabledRuleInfo) error {
	// open new CSV file to be filled in

	// disable "G304 (CWE-22): Potential file inclusion via variable"
	fout, err := os.Create(fileName) // #nosec G304
	if err != nil {
		return err
	}

	// conversion to CSV
	err = DisabledRulesToCSV(fout, disabledRulesInfo)
	if err != nil {
		log.Error().Err(err).Msg(writeDisabledRuleInfoToCSV)
		return err
	}

	// close the file and check if close operation was ok
	err = fout.Close()
	if err != nil {
		return err
	}

	return nil
}
