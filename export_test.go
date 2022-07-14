/*
Copyright © 2022  Red Hat, Inc.

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

// Generated documentation is available at:
// https://pkg.go.dev/github.com/RedHatInsights/insights-results-aggregator-exporter
//
// Documentation in literate-programming-style is available at:
// https://redhatinsights.github.io/insights-results-aggregator-exporter/packages/export_test.html
package main

// Export for testing
//
// This source file contains name aliases of all package-private functions
// that need to be called from unit tests. Aliases should start with uppercase
// letter because unit tests belong to different package.
//
// Please look into the following blogpost:
// https://medium.com/@robiplus/golang-trick-export-for-test-aa16cbd7b8cd
// to see why this trick is needed.
var (
	// exported functions from the exporter.go source file
	ShowVersion               = showVersion
	ShowAuthors               = showAuthors
	ShowConfiguration         = showConfiguration
	DoSelectedOperation       = doSelectedOperation
	PrintTables               = printTables
	ParseFlags                = parseFlags
	CheckS3Connection         = checkS3Connection
	PerformDataExport         = performDataExport
	ConstructIgnoredTablesMap = constructIgnoredTablesMap
	GetS3Bucket               = getS3Bucket

	// exported functions from the s3.go source file
	S3BucketExists  = s3BucketExists
	StoreTableNames = storeTableNames

	// exported functions from the file.go source file
	StoreTableNamesIntoFile    = storeTableNamesIntoFile
	StoreDisabledRulesIntoFile = storeDisabledRulesIntoFile
)
