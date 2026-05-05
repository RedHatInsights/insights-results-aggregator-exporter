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
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/csv.html

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/rs/zerolog/log"
)

const bufferIsNil = "Buffer is nil"

// DisabledRulesToCSV function exports list of disabled rules + number of users
// who disabled rules to CSV file.
func DisabledRulesToCSV(buffer io.Writer, disabledRulesInfo []DisabledRuleInfo) error {
	if buffer == nil {
		err := errors.New(bufferIsNil)
		return err
	}

	writer := csv.NewWriter(buffer)

	var data = [][]string{{"Rule", "Count"}}

	err := writer.WriteAll(data)
	if err != nil {
		return err
	}

	for _, disabledRuleInfo := range disabledRulesInfo {
		err := writer.Write([]string{
			disabledRuleInfo.Rule,
			strconv.Itoa(disabledRuleInfo.Count)})
		if err != nil {
			return err
		}
	}

	writer.Flush()

	// check for any error during export to CSV
	err = writer.Error()
	if err != nil {
		return err
	}

	return nil
}

// TableMetadataToCSV function exports list of table names into CSV file.
func TableMetadataToCSV(buffer io.Writer, tableNames []TableName, storage DBStorage) error {
	if buffer == nil {
		err := errors.New(bufferIsNil)
		return err
	}

	writer := csv.NewWriter(buffer)

	err := writer.Write([]string{tableNameMsg, "Records"})
	if err != nil {
		log.Error().Err(err).Msg(writeOneRowToCSV)
		return err
	}

	for _, tableName := range tableNames {
		cnt, err := storage.ReadRecordsCount(tableName)
		if err != nil {
			log.Error().Err(err).Msg(readListOfRecordsFailed)
			return err
		}

		columns := []string{string(tableName), strconv.Itoa(cnt)}

		err = writer.Write(columns)
		if err != nil {
			log.Error().Err(err).Msg(writeOneRowToCSV)
			return err
		}
	}

	writer.Flush()

	// check for any error during export to CSV
	err = writer.Error()
	if err != nil {
		return err
	}

	return nil
}

// LoadOrgIDsFromCSV creates a new CSV reader and returns a list of
// organization IDs
func LoadOrgIDsFromCSV(r io.Reader) ([]string, error) {
	orgIDs := make([]string, 0)

	reader := csv.NewReader(r)

	lines, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error reading CSV file: %v", err)
	}

	for index, line := range lines {
		if index == 0 {
			continue // skip header
		}

		orgID, err := strconv.ParseUint(line[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf(
				"organization ID on line %v in CSV is not numerical. Found value: %v",
				index+1, line[0],
			)
		}

		orgIDs = append(orgIDs, fmt.Sprintf("%d", orgID))
	}

	return orgIDs, nil
}
