#!/bin/bash
# Copyright 2023 Red Hat, Inc
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
APP_NAME="ccx-data-pipeline"  # name of app-sre "application" folder this component lives in
COMPONENT_NAME="insights-results-aggregator-exporter"  # name of app-sre "resourceTemplate" in deploy.yaml for this component
IMAGE="quay.io/cloudservices/insights-results-aggregator-exporter"
COMPONENTS="ccx-insights-results insights-results-aggregator-exporter ccx-notification-service ccx-notification-db-cleaner"  # space-separated list of components to laod
COMPONENTS_W_RESOURCES="insights-results-aggregator-exporter"  # component to keep
CACHE_FROM_LATEST_IMAGE="true"
# Set the correct images for pull requests.
# pr_check in pull requests still uses the old cloudservices images
EXTRA_DEPLOY_ARGS="--set-parameter insights-results-aggregator-exporter/IMAGE=quay.io/cloudservices/insights-results-aggregator-exporter"

export IQE_PLUGINS="ccx"
export IQE_MARKER_EXPRESSION=""
# Workaround: There are no cleaner specific integration tests. Check that the service loads and iqe plugin works.
export IQE_FILTER_EXPRESSION="test_plugin_accessible"
export IQE_REQUIREMENTS_PRIORITY=""
export IQE_TEST_IMPORTANCE=""
export IQE_CJI_TIMEOUT="30m"


function build_image() {
    source $CICD_ROOT/build.sh
}

function deploy_ephemeral() {
    source $CICD_ROOT/deploy_ephemeral_env.sh
}

function run_smoke_tests() {
    source $CICD_ROOT/cji_smoke_test.sh
    source $CICD_ROOT/post_test_results.sh  # publish results in Ibutsu
}


# Install bonfire repo/initialize
CICD_URL=https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd
curl -s $CICD_URL/bootstrap.sh > .cicd_bootstrap.sh && source .cicd_bootstrap.sh
echo "creating PR image"
build_image

echo "deploying to ephemeral"
deploy_ephemeral

echo "running PR smoke tests"
run_smoke_tests
