#!/usr/bin/env bash

set -e
set -o pipefail

# in case there are things we want to do differently knowing that we
# are running a test (e.g. to produce more predictible output);
# e.g. see 7/init.sh
export RUNNING_CODEFLARE_TESTS=1

while getopts "gu" opt
do
    case $opt in
        g) DEBUG=true; continue;;
        u) BRING_UP_CLUSTER=true; continue;;
    esac
done
shift $((OPTIND-1))

SCRIPTDIR=$(cd $(dirname "$0") && pwd)
. "$SCRIPTDIR"/../../hack/settings.sh

function up {
    local MAIN_SCRIPTDIR=$(cd $(dirname "$0") && pwd)
    if [[ -z "$NO_KUBEFLOW" ]]
    then
        "$MAIN_SCRIPTDIR"/../../hack/up.sh -t # -t says don't watch, just return when you are done
    else
        "$MAIN_SCRIPTDIR"/../../hack/up-no-kfp.sh -t # -t says don't watch, just return when you are done
    fi
}

function waitForIt {
    local name=$1
    local selector=app.kubernetes.io/part-of=$name
    local ns=$2
    local dones=("${@:3}") # an array formed from everything from the third argument on... 

    if [[ "$4" = ray ]]; then
        local containers="-c job-logs"
    else
        local containers="--all-containers"
    fi

    if [[ -n "$DEBUG" ]]; then set -x; fi

    # ($KUBECTL -n $ns wait pod -l $selector --for=condition=Completed --timeout=-1s && pkill $$)

    echo "$(tput setaf 2)🧪 Waiting for job to finish app=$selector ns=$ns$(tput sgr0)" 1>&2
    while true; do
        $KUBECTL -n $ns wait pod -l $selector --for=condition=Ready --timeout=5s && break || echo "$(tput setaf 5)🧪 Run not found: $selector$(tput sgr0)"

        $KUBECTL -n $ns wait pod -l $selector --for=condition=Ready=false --timeout=5s && break || echo "$(tput setaf 5)🧪 Run not found: $selector$(tput sgr0)"
        sleep 4
    done

    echo "$(tput setaf 2)🧪 Checking job output app=$selector$(tput sgr0)" 1>&2
    for done in "${dones[@]}"; do
        idx=0
        while true; do
            $KUBECTL -n $ns logs $containers -l $selector --tail=-1 | grep "$done" && break || echo "$(tput setaf 5)🧪 Still waiting for output $done... $selector$(tput sgr0)"

            if [[ -n $DEBUG ]] || (( $idx > 10 )); then
                # if we can't find $done in the logs after a few
                # iterations, start printing out raw logs to help with
                # debugging
                ($KUBECTL -n $ns logs $containers -l $selector --tail=4 || exit 0)
            fi
            idx=$((idx + 1))
            sleep 4
        done
    done

    echo "✅ PASS run-controller run test $selector"

    $KUBECTL delete run $name -n $ns
    echo "✅ PASS run-controller delete test $selector"
}

function waitForStatus {
    local name=$1
    local ns=$2
    local statuses=("${@:3}") # an array formed from everything from the third argument on... 

    if [[ -n "$DEBUG" ]]; then set -x; fi

    echo "$(tput setaf 2)🧪 Waiting for job to finish app=$selector ns=$ns$(tput sgr0)" 1>&2
    for status in "${statuses[@]}"
    do
        while true
        do
            $KUBECTL -n $ns get run.codeflare.dev $name --no-headers | grep -q "$status" && break || echo "$(tput setaf 5)🧪 Still waiting for Failed: $name$(tput sgr0)"
            ($KUBECTL -n $ns get run.codeflare.dev $name --no-headers | grep $name || exit 0)
            sleep 4
        done
    done

    echo "✅ PASS run-controller run test $name"

    $KUBECTL delete run $name -n $ns
    echo "✅ PASS run-controller delete test $name"
}

function deploy {
    "$SCRIPTDIR"/deploy-tests.sh $1 || exit 0
}

function undeploy {
    [[ -n "$2" ]] && kill $2
    ("$SCRIPTDIR"/undeploy-tests.sh $1 || exit 0)
}

function watch {
    if [[ -n "$CI" ]]; then
        $KUBECTL get appwrapper -n codeflare-test -o custom-columns=NAME:.metadata.name,CONDITIONS:.status.conditions --watch &
        $KUBECTL get pod --show-kind -n codeflare-test --watch &
    fi
    $KUBECTL get pod --show-kind -n codeflare-system --watch &
}
