#!/usr/bin/env bash

while getopts "ki" opt
do
    case $opt in
        k) KILL=true; continue;;
        i) INIT=true; continue;;
    esac
done

set -x

SCRIPTDIR=$(cd $(dirname "$0") && pwd)
TOP="$SCRIPTDIR"/../../..

. "$TOP"/hack/settings.sh

if [[ ! -n "$INIT" ]]; then 
    set +e
    $KUBECTL delete -f resources/jaas-default-user.yml --ignore-not-found & \
        $KUBECTL delete -f resources/jaas-defaults.yml --ignore-not-found
    wait
    $KUBECTL delete -f resources/jaas-lite.yml --ignore-not-found --grace-period=1
else
    "$TOP"/hack/init.sh;
    wait
fi

if [[ -n "$KILL" ]]; then exit; fi

set -e
set -o pipefail

if (podman machine inspect | grep State | grep running) && (kind get nodes -n ${CLUSTER_NAME} | grep ${CLUSTER_NAME})
then 
        if ! kubectl get nodes --context kind-${CLUSTER_NAME} | grep ${CLUSTER_NAME}
        then # podman must have restarted causing kind node to be deleted
            kind get nodes -A | xargs -n1 podman start
        fi
else
    "$TOP"/hack/init.sh;
    wait
fi

# rebuild the controller images & the dashboard includes a precompiled version of the jaas charts
../../hack/build.sh -l & ./hack/generate-installers.sh
wait

$KUBECTL apply -f resources/jaas-lite.yml
$KUBECTL apply -f resources/jaas-defaults.yml & \
    $KUBECTL apply -f resources/jaas-default-user.yml
wait