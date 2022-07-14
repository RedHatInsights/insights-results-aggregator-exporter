# insights-results-aggregator-exporter

[![GoDoc](https://godoc.org/github.com/RedHatInsights/insights-results-aggregator-exporter?status.svg)](https://godoc.org/github.com/RedHatInsights/insights-results-aggregator-exporter)
[![GitHub Pages](https://img.shields.io/badge/%20-GitHub%20Pages-informational)](https://redhatinsights.github.io/insights-results-aggregator-exporter/)
[![Go Report Card](https://goreportcard.com/badge/github.com/RedHatInsights/insights-results-aggregator-exporter)](https://goreportcard.com/report/github.com/RedHatInsights/insights-results-aggregator-exporter)
[![Build Status](https://travis-ci.com/RedHatInsights/insights-results-aggregator-exporter.svg?branch=master)](https://travis-ci.com/RedHatInsights/insights-results-aggregator-exporter)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/RedHatInsights/insights-results-aggregator-exporter)
[![License](https://img.shields.io/badge/license-Apache-blue)](https://github.com/RedHatInsights/insights-results-aggregator-exporter/blob/master/LICENSE)

Exporter for Insights Results data stored by Insights Results Aggregator

<!-- vim-markdown-toc GFM -->

* [Description](#description)
* [Documentation](#documentation)
* [Contribution](#contribution)
* [Usage](#usage)
	* [Building](#building)
* [CI/CD](#cicd)
* [Makefile targets](#makefile-targets)
	* [Configuration](#configuration)
* [BDD tests](#bdd-tests)
* [Example output](#example-output)
	* [List of files/objects](#list-of-filesobjects)
	* [Content of `_tables.csv`](#content-of-_tablescsv)
	* [Content of `_metadata.csv`](#content-of-_metadatacsv)
	* [Content of `advisor_ratings.csv`](#content-of-advisor_ratingscsv)
	* [Content of `cluster_rule_toggle.csv`](#content-of-cluster_rule_togglecsv)
	* [Content of `cluster_rule_user_feedback.csv`](#content-of-cluster_rule_user_feedbackcsv)
	* [Content of `cluster_user_rule_disable_feedback.csv`](#content-of-cluster_user_rule_disable_feedbackcsv)
	* [Content of `migration_info.csv`](#content-of-migration_infocsv)
	* [Content of `consumer_error.csv`](#content-of-consumer_errorcsv)
	* [Content of `recommendation.csv`](#content-of-recommendationcsv)
	* [Content of `report.csv`](#content-of-reportcsv)
	* [Content of `report_info.csv`](#content-of-report_infocsv)
	* [Content of `rule_disable.csv`](#content-of-rule_disablecsv)
	* [Content of `rule_hit.csv`](#content-of-rule_hitcsv)
* [Package manifest](#package-manifest)

<!-- vim-markdown-toc -->

## Description

Insights Results Aggregator is simple service that is able to read data from
selected database (PostgreSQL, RDS etc.) and store the data as set of CSV files
and (optionally) into S3 bucket. It can also store metadata about the database
- list of tables, number of records stored in tables etc. That service can be
used to make a database snapshot, even for databases that are not directly
reachable by user.

## Documentation

Documentation is hosted on Github Pages <https://redhatinsights.github.io/insights-results-aggregator-exporter/>.
Sources are located in [docs](https://github.com/RedHatInsights/insights-results-aggregator-exporter/tree/master/docs).


## Contribution

Please look into document [CONTRIBUTING.md](CONTRIBUTING.md) that contains all information about how to
contribute to this project.


## Usage

```
Usage of ./irae:
  -authors
        show authors
  -check-s3-connection
        check S3 connection and exit
  -disabled-by-more-users
         export rules disabled by more than one user
  -export-log
        export log
  -ignore-tables string
        comma-separated list of tables that will be ignored
  -limit int
        limit number of exported records (default -1)
  -metadata
        export metadata
  -output string
        output to: CSV, S3
  -show-configuration
        show configuration
  -summary
        print summary table after export
  -version
        show version
```

### Building

Go version 1.16 or newer is required to build this tool.

Executable binary file with this tool can be produced by using the following
command:

```
make build
```

You can build a container image as well; make sure your current working
directory is the project root, then run

```
docker build -t insights-results-aggregator-exporter .
```

## CI/CD

Every time a commit is pushed to any branch, there are some automatic checks ran:

* Travis CI: runs some linting and unit testing. The definition can be found at [.travis.yml](.travis.yml).
* Golang CI Lint: runs a quite complete linting of the go scripts in the repo. The definition can be found at [golangci-lint.yml](.github/workflows/golangci-lint.yml). It runs on GitHub Actions.

Also, when a commit is added to `master` (basically when a PR is merged) a ci.int pipeline builds the container image for this repo, which can be located at https://quay.io/repository/cloudservices/insights-results-aggregator-exporter. The configuration of this pipeline can be found at [build_deploy.sh](build_deploy.sh). It also runs some PR checks on each commit, which are configured at [pr_check.sh](pr_check.sh).

## Makefile targets

List of all Makefile targets currently supported:

```
Usage: make <OPTIONS> ... <TARGETS>

Available targets are:

clean                Run go clean
build                Keep this rule for compatibility
fmt                  Run go fmt -w for all sources
lint                 Run golint
vet                  Run go vet. Report likely mistakes in source code
cyclo                Run gocyclo
ineffassign          Run ineffassign checker
shellcheck           Run shellcheck
errcheck             Run errcheck
goconst              Run goconst checker
gosec                Run gosec checker
abcgo                Run ABC metrics checker
style                Run all the formatting related commands (fmt, vet, lint, cyclo) + check shell scripts
run                  Build the project and executes the binary
test                 Run the unit tests
before_commit        Checks done before commit
help                 Show this help screen
```

### Configuration

Default name of configuration file is `config.toml`.
It can be changed via environment variable `INSIGHTS_RESULTS_EXPORTER_CONFIG_FILE`.

An example of configuration file that can be used in devel environment:

```
[storage]
db_driver = "postgres"
pg_username = "postgres"
pg_password = "postgres"
pg_host = "localhost"
pg_port = 5432
pg_db_name = "aggregator"
pg_params = "sslmode=disable"

[s3]
type = "minio"
endpoint_url = "127.0.0.1"
endpoint_port = 9000
access_key_id = "foobar"
secret_access_key = "foobar"
use_ssl = false
bucket = "test"
prefix = "prefix"

[logging]
debug = true
log_level = ""

[sentry]
dsn = ""
environment = "dev"
```

Environment variables that can be used to override configuration file settings:

```
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__DB_DRIVER
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_USERNAME
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_PASSWORD
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_HOST
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_PORT
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_DB_NAME
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__STORAGE__PG_PARAMS
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__TYPE
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__ENDPOINT_URL
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__ENDPOINT_PORT
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__ACCESS_KEY_ID
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__SECRET_ACCESS_KEY
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__USE_SSL
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__BUCKET
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__S3__PREFIX
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__LOGGING__DEBUG
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__LOGGING__LOG_DEVEL
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__SENTRY__DSN
INSIGHTS_RESULTS_AGGREGATOR_EXPORTER__SENTRY__ENVIRONMENT
```

## BDD tests

Behaviour tests for this service are included in [Insights Behavioral
Spec](https://github.com/RedHatInsights/insights-behavioral-spec) repository.
In order to run these tests, the following steps need to be made:

1. clone the [Insights Behavioral Spec](https://github.com/RedHatInsights/insights-behavioral-spec) repository
1. go into the cloned subdirectory `insights-behavioral-spec`
1. run the `exporter_tests.sh` from this subdirectory

List of all test scenarios prepared for this service is available at
<https://github.com/RedHatInsights/insights-behavioral-spec#insights-results-aggregator-exporter-service>



## Example output

When this tool is started against standard Insights Aggregator database, the output should be like this:

### List of files/objects

```
_tables.csv
_metadata.csv
advisor_ratings.csv
cluster_rule_toggle.csv
cluster_rule_user_feedback.csv
cluster_user_rule_disable_feedback.csv
consumer_error.csv
migration_info.csv
recommendation.csv
report.csv
report_info.csv
rule_disable.csv
rule_hit.csv
```

### Content of `_tables.csv`

```
Table name
consumer_error
advisor_ratings
report
migration_info
cluster_user_rule_disable_feedback
rule_disable
recommendation
cluster_rule_user_feedback
rule_hit
cluster_rule_toggle
report_info
```

### Content of `_metadata.csv`

```
Table name,Records
consumer_error,10
advisor_ratings,10
report,5
migration_info,1
cluster_user_rule_disable_feedback,155
rule_disable,15
recommendation,35045
cluster_rule_user_feedback,9
rule_hit,45
cluster_rule_toggle,155
report_info,5
```

### Content of `advisor_ratings.csv`

```
user_id,org_id,rule_fqdn,error_key,rated_at,last_updated_at,rating,rule_id
10,5,ocp.rules.telemetry.version_info,ocp.rules.telemetry,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,1,version_info
11,6,ocp.rules.telemetry.version_info,ocp.rules.telemetry,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,1,version_info
10,8,ocp.rules.other_err.rule,ocp.rules.other_err,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,1,rule
20,6,ocp.rules.telemetry.version_info,ocp.rules.telemetry,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,1,version_info
20,7,ocp.rules.other_err.rule,ocp.rules.other_err,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,1,rule
20,9,ocp.rules.telemetry.version_info,ocp.rules.telemetry,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,1,version_info
20,8,ocp.rules.other_err.rule,ocp.rules.other_err,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,1,rule
```
### Content of `cluster_rule_toggle.csv`

```
cluster_id,rule_id,user_id,disabled,disabled_at,enabled_at,updated_at,error_key
00000001-624a-49a5-bab8-4fdc5e51a266,ccx_rules_ocp.external.bug_rules.bug_1821905.report,4,1,2021-09-20T00:00:00Z,,2021-09-20T00:00:00Z,BUGZILLA_BUG_1821905
00000001-624a-49a5-bab8-4fdc5e51a266,ccx_rules_ocp.external.rules.nodes_requirements_check.report,3,1,2021-09-20T00:00:00Z,,2021-09-20T00:00:00Z,NODES_MINIMUM_REQUIREMENTS_NOT_MET
00000001-624a-49a5-bab8-4fdc5e51a266,ccx_rules_ocp.external.bug_rules.bug_1766907.report,3,1,2021-09-20T00:00:00Z,,2021-09-20T00:00:00Z,BUGZILLA_BUG_1766907
00000001-624a-49a5-bab8-4fdc5e51a266,ccx_rules_ocp.external.rules.nodes_kubelet_version_check.report,2,1,2021-09-20T00:00:00Z,,2021-09-20T00:00:00Z,NODE_KUBELET_VERSION
ee7d2bf4-8933-4a3a-8634-3328fe806e08,ccx_rules_ocp.external.bug_rules.bug_1821905.report,3,1,2021-09-20T00:00:00Z,,2021-09-20T00:00:00Z,BUGZILLA_BUG_1821905
ee7d2bf4-8933-4a3a-8634-3328fe806e08,ccx_rules_ocp.external.rules.nodes_requirements_check.report,2,1,2021-09-20T00:00:00Z,,2021-09-20T00:00:00Z,NODES_MINIMUM_REQUIREMENTS_NOT_MET
ee7d2bf4-8933-4a3a-8634-3328fe806e08,ccx_rules_ocp.external.bug_rules.bug_1766907.report,3,1,2021-09-20T00:00:00Z,,2021-09-20T00:00:00Z,BUGZILLA_BUG_1766907
ee7d2bf4-8933-4a3a-8634-3328fe806e08,ccx_rules_ocp.external.rules.nodes_kubelet_version_check.report,3,1,2021-09-20T00:00:00Z,,2021-09-20T00:00:00Z,NODE_KUBELET_VERSION
ee7d2bf4-8933-4a3a-8634-3328fe806e08,ccx_rules_ocp.external.rules.samples_op_failed_image_import_check.report,1,1,2021-09-20T00:00:00Z,,2021-09-20T00:00:00Z,SAMPLES_FAILED_IMAGE_IMPORT_ERR
```
### Content of `cluster_rule_user_feedback.csv`

```
cluster_id,rule_id,user_id,message,user_vote,added_at,updated_at,error_key
6d5892d3-1f74-4ccf-91af-548dfc9767aa,ocp.rules.telemetry,1,msg,1,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,VERSION_INFO
7d5892d3-1f74-4ccf-91af-548dfc9767aa,ocp.rules.telemetry,2,msg,1,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,VERSION_INFO
9d5892d3-1f74-4ccf-91af-548dfc9767ac,ocp.rules.telemetry,3,msg,1,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,VERSION_INFO
8d5892d3-1f74-4ccf-91af-548dfc9767aa,other,10,msg,1,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,RULE
5d5892d3-1f74-4ccf-91af-548dfc9767aa,other,11,msg,1,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,RULE
6d5892d3-1f74-4ccf-91af-548dfc9767aa,ocp.rules.telemetry,0,msg,1,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,VERSION_INFO
7d5892d3-1f74-4ccf-91af-548dfc9767aa,other,0,msg,1,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,RULE
9d5892d3-1f74-4ccf-91af-548dfc9767ac,ocp.rules.telemetry,0,msg,1,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,VERSION_INFO
8d5892d3-1f74-4ccf-91af-548dfc9767aa,other,0,msg,1,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,RULE
```

### Content of `cluster_user_rule_disable_feedback.csv`

```
cluster_id,user_id,rule_id,message,added_at,updated_at,error_key
00000001-624a-49a5-bab8-4fdc5e51a266,4,ccx_rules_ocp.external.bug_rules.bug_1821905.report,Rule ccx_rules_ocp.external.bug_rules.bug_1821905.report|BUGZILLA_BUG_1821905 for cluster 00000001-624a-49a5-bab8-4fdc5e51a266 disabled by 4,2021-09-20T00:00:00Z,2021-09-20T00:00:00Z,BUGZILLA_BUG_1821905
00000001-624a-49a5-bab8-4fdc5e51a266,3,ccx_rules_ocp.external.rules.nodes_requirements_check.report,Rule ccx_rules_ocp.external.rules.nodes_requirements_check.report|NODES_MINIMUM_REQUIREMENTS_NOT_MET for cluster 00000001-624a-49a5-bab8-4fdc5e51a266 disabled by 3,2021-09-20T00:00:00Z,2021-09-20T00:00:00Z,NODES_MINIMUM_REQUIREMENTS_NOT_MET
00000001-624a-49a5-bab8-4fdc5e51a266,3,ccx_rules_ocp.external.bug_rules.bug_1766907.report,Rule ccx_rules_ocp.external.bug_rules.bug_1766907.report|BUGZILLA_BUG_1766907 for cluster 00000001-624a-49a5-bab8-4fdc5e51a266 disabled by 3,2021-09-20T00:00:00Z,2021-09-20T00:00:00Z,BUGZILLA_BUG_1766907
ee7d2bf4-8933-4a3a-8634-3328fe806e08,3,ccx_rules_ocp.external.rules.nodes_kubelet_version_check.report,Rule ccx_rules_ocp.external.rules.nodes_kubelet_version_check.report|NODE_KUBELET_VERSION for cluster ee7d2bf4-8933-4a3a-8634-3328fe806e08 disabled by 3,2021-09-20T00:00:00Z,2021-09-20T00:00:00Z,NODE_KUBELET_VERSION
ee7d2bf4-8933-4a3a-8634-3328fe806e08,1,ccx_rules_ocp.external.rules.samples_op_failed_image_import_check.report,Rule ccx_rules_ocp.external.rules.samples_op_failed_image_import_check.report|SAMPLES_FAILED_IMAGE_IMPORT_ERR for cluster ee7d2bf4-8933-4a3a-8634-3328fe806e08 disabled by 1,2021-09-20T00:00:00Z,2021-09-20T00:00:00Z,SAMPLES_FAILED_IMAGE_IMPORT_ERR
```

### Content of `migration_info.csv`

```
version
23
```

### Content of `consumer_error.csv`

```
topic,partition,topic_offset,key,produced_at,consumed_at,message,error
test-topic,1,100,10,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,message 1,error 1
test-topic,1,101,11,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,message 2,error 2
test-topic,1,102,12,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,message 3,error 3
test-topic,1,103,13,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,message 4,error 4
test-topic,1,104,10,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,message 5,error 5
test-topic,1,105,11,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,message 6,error 6
test-topic,1,106,20,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,message 7,error 7
test-topic,1,107,20,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,message 8,error 8
test-topic,1,108,20,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,message 9,error 9
test-topic,1,109,20,2020-01-01T00:00:00Z,2020-01-01T00:00:00Z,message !,error !
```

### Content of `recommendation.csv`

```
org_id,cluster_id,rule_fqdn,error_key,rule_id,created_at
61209472,49d95631-6933-466c-80a8-1e97556c2289,ccx_rules_ocp.external.rules.nodes_requirements_check,NODES_MINIMUM_REQUIREMENTS_NOT_MET,ccx_rules_ocp.external.rules.nodes_requirements_check|NODES_MINIMUM_REQUIREMENTS_NOT_MET,2022-04-20T06:42:28Z
61209472,49d95631-6933-466c-80a8-1e97556c2289,ccx_rules_ocp.external.bug_rules.bug_1766907,BUGZILLA_BUG_1766907,ccx_rules_ocp.external.bug_rules.bug_1766907|BUGZILLA_BUG_1766907,2022-04-20T06:42:28Z
61209472,49d95631-6933-466c-80a8-1e97556c2289,ccx_rules_ocp.external.rules.nodes_kubelet_version_check,NODE_KUBELET_VERSION,ccx_rules_ocp.external.rules.nodes_kubelet_version_check|NODE_KUBELET_VERSION,2022-04-20T06:42:28Z
61209472,49d95631-6933-466c-80a8-1e97556c2289,ccx_rules_ocp.external.rules.samples_op_failed_image_import_check,SAMPLES_FAILED_IMAGE_IMPORT_ERR,ccx_rules_ocp.external.rules.samples_op_failed_image_import_check|SAMPLES_FAILED_IMAGE_IMPORT_ERR,2022-04-20T06:42:28Z
61209472,49d95631-6933-466c-80a8-1e97556c2289,ccx_rules_ocp.external.rules.cluster_wide_proxy_auth_check,AUTH_OPERATOR_PROXY_ERROR,ccx_rules_ocp.external.rules.cluster_wide_proxy_auth_check|AUTH_OPERATOR_PROXY_ERROR,2022-04-20T06:42:28Z
55509440,2ba44dd8-57e0-4b5a-ac60-84c0ba2739e8,ccx_rules_ocp.external.rules.nodes_requirements_check,NODES_MINIMUM_REQUIREMENTS_NOT_MET,ccx_rules_ocp.external.rules.nodes_requirements_check|NODES_MINIMUM_REQUIREMENTS_NOT_MET,2022-04-20T06:42:28Z
```

### Content of `report.csv`

```
org_id,cluster,report,reported_at,last_checked_at,kafka_offset,gathered_at
11789772,5d5892d3-1f74-4ccf-91af-548dfc9767aa,"{...too big...}",2022-04-27T08:53:00.639083Z,2020-04-02T09:00:05.268294Z,1,
```

### Content of `report_info.csv`

```
org_id,cluster_id,version_info
0,5d5892d3-1f74-4ccf-91af-548dfc9767aa,1.2.3
1,6d5892d3-1f74-4ccf-91af-548dfc9767aa,4.5.6
2,7d5892d3-1f74-4ccf-91af-548dfc9767aa,7.8.9
3,9d5892d3-1f74-4ccf-91af-548dfc9767ac,1.2.3
4,8d5892d3-1f74-4ccf-91af-548dfc9767aa,4.5.6
```

### Content of `rule_disable.csv`

```
org_id,user_id,rule_id,error_key,justification,created_at,updated_at
1,2,ccx_rules_ocp.external.rules.ccxdev_auxiliary_rule,CCXDEV_E2E_TEST_RULE,Rule ccx_rules_ocp.external.rules.ccxdev_auxiliary_rule|CCXDEV_E2E_TEST_RULE has been disabled by 2,2021-09-27T00:00:00Z,2021-09-27T00:00:00Z
1,5,ccx_rules_ocp.external.bug_rules.bug_1821905.report,BUGZILLA_BUG_1821905,Rule ccx_rules_ocp.external.bug_rules.bug_1821905.report|BUGZILLA_BUG_1821905 has been disabled by 5,2021-09-27T00:00:00Z,2021-09-27T00:00:00Z
1,5,ccx_rules_ocp.external.rules.nodes_requirements_check.report,NODES_MINIMUM_REQUIREMENTS_NOT_MET,Rule ccx_rules_ocp.external.rules.nodes_requirements_check.report|NODES_MINIMUM_REQUIREMENTS_NOT_MET has been disabled by 5,2021-09-27T00:00:00Z,2021-09-27T00:00:00Z
1,5,ccx_rules_ocp.external.bug_rules.bug_1766907.report,BUGZILLA_BUG_1766907,Rule ccx_rules_ocp.external.bug_rules.bug_1766907.report|BUGZILLA_BUG_1766907 has been disabled by 5,2021-09-27T00:00:00Z,2021-09-27T00:00:00Z
1,4,ccx_rules_ocp.external.rules.nodes_kubelet_version_check.report,NODE_KUBELET_VERSION,Rule ccx_rules_ocp.external.rules.nodes_kubelet_version_check.report|NODE_KUBELET_VERSION has been disabled by 4,2021-09-27T00:00:00Z,2021-09-27T00:00:00Z
1,1,ccx_rules_ocp.external.rules.samples_op_failed_image_import_check.report,SAMPLES_FAILED_IMAGE_IMPORT_ERR,Rule ccx_rules_ocp.external.rules.samples_op_failed_image_import_check.report|SAMPLES_FAILED_IMAGE_IMPORT_ERR has been disabled by 1,2021-09-27T00:00:00Z,2021-09-27T00:00:00Z
1,2,foo,bar,???,2021-09-27T14:11:02.400621Z,2021-09-27T14:12:05.712933Z
1234,5213476,foot,bar,xyzzy,2021-09-28T17:31:01.52242Z,
1234,5213476,foo,bar,new_justification,2021-09-29T13:47:43.943085Z,2021-09-29T17:57:32.507539Z
```

### Content of `rule_hit.csv`

```
org_id,cluster_id,rule_fqdn,error_key,template_data
11789773,6d5892d3-1f74-4ccf-91af-548dfc9767aa,ccx_rules_ocp.external.rules.nodes_requirements_check.report,NODES_MINIMUM_REQUIREMENTS_NOT_MET,"{
                    ""nodes"": [
                        {
                            ""name"": ""foo1"",
                            ""role"": ""master"",
                            ""memory"": 8.16,
                            ""memory_req"": 16
                        }
                    ],
                    ""link"": ""https://docs.openshift.com/container-platform/4.1/installing/installing_bare_metal/installing-bare-metal.html#minimum-resource-requirements_installing-bare-metal"",
                    ""type"": ""rule"",
                    ""error_key"": ""NODES_MINIMUM_REQUIREMENTS_NOT_MET""
                }"
11789773,6d5892d3-1f74-4ccf-91af-548dfc9767aa,ccx_rules_ocp.external.bug_rules.bug_1766907.report,BUGZILLA_BUG_1766907,"{
                    ""type"": ""rule"",
                    ""error_key"": ""BUGZILLA_BUG_1766907""
                }"
```

## Package manifest

Package manifest is available at [docs/manifest.txt](docs/manifest.txt).
