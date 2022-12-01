#!/bin/bash

########################################
# Presubmit script triggered by Prow.  #
########################################
action=$1
if [[ ! ${action} ]]; then
    action='default'
fi

WD=$(dirname $0)
WD=$(cd $WD; pwd)
ROOT=$(dirname $WD)
source prow/functions.sh

commenter

# Exit immediately for non zero status
set -e
# Check unset variables
set -u
# Print command trace
set -x

run_unit_tests


exit 0
