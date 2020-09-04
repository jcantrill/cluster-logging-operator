#!/bin/bash -x

set -eou pipefail

repo_dir="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )/../.."
KUBECONFIG=${KUBECONFIG:-$HOME/.kube/config}
NAMESPACE=openshift-logging
generator="$repo_dir/internal/cmd/forwarder-generator/forwarder-generator"

# cleanup(){
#   local return_code="$?"

#   set +e
#   if [ "${DO_CLEANUP:-true}" == "true" ] ; then
#   fi
  
#   set -e
#   exit ${return_code}
# }
# trap cleanup exit

if [ "${DO_SETUP:-true}" == "true" ] ; then
 
  echo "Deploying ClusterLogForwarder ..."
echo 'apiVersion: "logging.openshift.io/v1"
kind: "ClusterLogForwarder"
metadata:
  name: "instance"
spec:
  outputs:
  - name: stdout
    type: stdout
  pipelines:
  - name: test-pipeline
    inputrefs: ["application"]
    outputrefs: ["stdout"]
' | $generator --include-default-store=false --file -

fi

