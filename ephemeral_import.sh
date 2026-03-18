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

# to run this script, you need to have bonfire and oc set up correctly and be logged in to the ephemeral cluster.
# put the exported .csv files into a ./testdata path, the rest will be done automatically
# see the docs on how to login to the eph. cluster https://ccx.pages.redhat.com/ccx-docs/docs/processing/howto/ephemeral_env/

GREEN_BG=$(tput setab 2)
RED_BG=$(tput setab 1)
NC=$(tput sgr0) # No Color

APP_NAME="ccx-data-pipeline"
REF_ENV="insights-production"
TESTDATA_PATH="testdata"
DB_POD_DATA_PATH="/tmp/"
TABLES=(
   "report"
   "recommendation"
   "rule_hit"
   "rule_disable"
   "rule_toggle"
   "cluster_rule_user_feedback"
   "cluster_user_rule_disable_feedback"
   "advisor_ratings"
)

# define reserved namespace if you want to use a special one
EPH_NAMESPACE=${EPH_NAMESPACE:=""}
DB_POD=""

export IQE_ENV="ephemeral"

function deploy_ephemeral() {
   # deploy all external pipeline services as they are in REF_ENV. output to separate file
   bonfire deploy ccx-data-pipeline --source=appsre --ref-env ${REF_ENV} --namespace ${EPH_NAMESPACE} > bonfire_deploy.out 2>&1 &
}

function cleanup() {
   rm -f bonfire_deploy.out
   rm -f ${TESTDATA_PATH}/import.sql
}

function copy_testdata() {
   if ! oc rsync ${TESTDATA_PATH} ${DB_POD}:${DB_POD_DATA_PATH}; then
      echo "unable to copy files over to DB pod"
      exit 1
   fi
}

function generate_import_script() {
   # generate import script in correct order
   FILENAME="${TESTDATA_PATH}/import.sql"

   # connect to DB
   echo "\c ${APP_NAME};" > ${FILENAME}
   echo "BEGIN;" >> ${FILENAME}

   # import only existing files
   for TABLENAME in "${TABLES[@]}"; do
      if [ ! -f ${TESTDATA_PATH}/${TABLENAME}.csv ]; then
         echo "${TABLENAME}.csv not found, skipping"
         continue
      fi
      # copy command for psql
      echo "\COPY ${TABLENAME} FROM '${DB_POD_DATA_PATH}/${TESTDATA_PATH}/${TABLENAME}.csv' DELIMITER ',' CSV HEADER;"  >> ${FILENAME}

   done
   echo "COMMIT;"  >> ${FILENAME}
}

function import_testdata_to_db {
   # cannot check for return code via this method
   oc exec ${DB_POD} -- bash -c "psql < ${DB_POD_DATA_PATH}/${TESTDATA_PATH}/import.sql"
}

function wait_for_db_pod() {
   # wait for DB pod to appear in deployments
   DB_POD=""
   while [[ ${DB_POD} == "" ]]
   do
      sleep 5
      DB_POD=$(oc get pods --selector='app=ccx-insights-results, service=db' | grep "ccx-insights-results-db-*" | awk '{print $1}')
   done

   # we have the pod name, we can use oc wait command
   echo "${GREEN_BG}DB pod ${DB_POD} spawned, waiting for Ready state${NC}"
   if ! oc wait --for=condition=Ready pod/${DB_POD} --timeout=2m; then
      echo "waiting for DB pod timed out, exiting"
      exit 1
   fi
   echo "${GREEN_BG}DB pod ${DB_POD} ready${NC}"
}

function wait_for_db_writer_pod() {
   # wait for DB writer pod to be Ready, it will not be in a Ready state until all init containers finish,
   # ensuring the DB migrations will be finished before we proceed
   DB_WRITER_POD=""
   while [[ ${DB_WRITER_POD} == "" ]]
   do
      sleep 3
      DB_WRITER_POD=$(oc get pods --selector='app=ccx-insights-results, pod=ccx-insights-results-db-writer' | grep "ccx-insights-results-db-writer-*" | awk '{print $1}')
   done

   # we have the pod name, we can use oc wait command to wait for Ready state, ensuring init contianers finish
   echo "${GREEN_BG}DB writer pod ${DB_WRITER_POD} found, waiting for database migrations to finish${NC}"
   if ! oc wait --for=condition=Ready pod/${DB_WRITER_POD} --timeout=2m; then
      echo "waiting for DB migrations to finish timed out, exiting"
      exit 1
   fi
   echo "${GREEN_BG}DB pod ${DB_WRITER_POD} ready${NC}"
}

function check_login() {
   if ! oc whoami;
   then
      echo "${RED_BG}please login to the ephemeral cluster via 'oc login'${NC}"
      echo "${RED_BG}https://ccx.pages.redhat.com/ccx-docs/docs/processing/howto/ephemeral_env/${NC}"
      exit 1
   fi
}

function reserve_new_ephemeral_namespace() {
   echo "${GREEN_BG}reserving new ephemeral namespace${NC}"
   EPH_NAMESPACE=$(bonfire namespace reserve)
   echo "${GREEN_BG}extending namespace reservation by 2 hours${NC}"
   bonfire namespace extend ${EPH_NAMESPACE} -d '2h0m'
}

function get_ephemeral_namespace() {
   if [[ $EPH_NAMESPACE == "" ]];
   then
      # try to find an already existing namespace
      FIRST_AVAILABLE=$(bonfire namespace list --mine | grep 'ephemeral-*' | awk '{print $1}' | head -1)
      if [[ $FIRST_AVAILABLE != "" ]];
      then
         echo "${GREEN_BG}Existing ephemeral namespace found, do you want to use ${FIRST_AVAILABLE}?${NC}"
         select yn in "Yes" "No"; do
            case $yn in
               Yes ) EPH_NAMESPACE=${FIRST_AVAILABLE}; break;;
               No ) echo "${RED_BG}please pick a namespace and export the EPH_NAMESPACE env var${NC}"; break;;
            esac
         done
      else
         echo "${GREEN_BG}no existing namespace found${NC}"
         reserve_new_ephemeral_namespace
      fi
   fi
}

echo "checking login"
check_login
echo "${GREEN_BG}user logged in to the ephemeral cluster${NC}"

echo "${GREEN_BG}getting ephemeral namespace${NC}"
get_ephemeral_namespace

if [[ $EPH_NAMESPACE == "" ]]; then
   echo "failed to reserve a namespace, exiting"
   exit 1
fi

echo "${GREEN_BG}chosen namespace ${EPH_NAMESPACE}, deploying external pipeline services${NC}"
deploy_ephemeral

echo "${GREEN_BG}waiting for database pod to start${NC}"
wait_for_db_pod

echo "${GREEN_BG}generating import script${NC}"
generate_import_script

echo "${GREEN_BG}copying exported data to DB pod${NC}"
copy_testdata
echo "${GREEN_BG}data copied succesfully to DB pod${NC}"

echo "${GREEN_BG}waiting for database migrations to finish${NC}"
wait_for_db_writer_pod

echo "${GREEN_BG}executing import script on DB pod${NC}"
import_testdata_to_db

echo "${GREEN_BG}data should now be imported to the ephemeral DB. you should be able to retrieve the data via the smart-proxy API${NC}"
cleanup
