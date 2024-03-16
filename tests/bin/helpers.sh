#!/usr/bin/env bash

set -e
set -o pipefail

# in case there are things we want to do differently knowing that we
# are running a test (e.g. to produce more predictible output);
# e.g. see 7/init.sh
export RUNNING_CODEFLARE_TESTS=1

while getopts "c:lgui:e:opr" opt
do
    case $opt in
        l) export HELM_INSTALL_FLAGS="--set lite=true"; export UP_FLAGS="$UP_FLAGS -l"; echo "$(tput setaf 3)🧪 Running in lite mode$(tput sgr0)"; continue;;
        e) EXCLUDE=$OPTARG; continue;;
        i) INCLUDE=$OPTARG; continue;;
        g) DEBUG=true; continue;;
        c) export UP_FLAGS="$UP_FLAGS -c $OPTARG"; continue;;
        o) export UP_FLAGS="$UP_FLAGS -o"; continue;;
        p) export UP_FLAGS="$UP_FLAGS -p"; continue;;
        r) export UP_FLAGS="$UP_FLAGS -r"; continue;;
        u) BRING_UP_CLUSTER=true; continue;;
    esac
done
xOPTIND=$OPTIND
OPTIND=1

TEST_FROM_ARGV_idx=$((xOPTIND))
export TEST_FROM_ARGV="${!TEST_FROM_ARGV_idx}"

SCRIPTDIR=$(cd $(dirname "$0") && pwd)
. "$SCRIPTDIR"/../../hack/settings.sh

function up {
    local MAIN_SCRIPTDIR=$(cd $(dirname "$0") && pwd)
    if [[ -z "$NO_KUBEFLOW" ]]
    then
        "$MAIN_SCRIPTDIR"/../../hack/up.sh -t $UP_FLAGS # -t says don't watch, just return when you are done
    else
        "$MAIN_SCRIPTDIR"/../../hack/up-no-kfp.sh -t $UP_FLAGS # -t says don't watch, just return when you are done
    fi
}

function waitForIt {
    local name=$1
    local ns=$2
    local dones=("${@:3}") # an array formed from everything from the third argument on... 

    # Future readers: the != part is meant to avoid any pods that are
    # known to be short-lived without this, we may witness a
    # combination of Ready and Complete (i.e. not-Ready) pods. This is
    # important because pthe kubectl waits below expect the pods
    # either to be all-Ready or all-not-Ready.
    local selector=app.kubernetes.io/part-of=$name,app.kubernetes.io/component!=workdispatcher

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
            $KUBECTL -n $ns logs $containers -l $selector --tail=-1 | grep -E "$done" && break || echo "$(tput setaf 5)🧪 Still waiting for output $done... $selector$(tput sgr0)"

            if [[ -n $DEBUG ]] || (( $idx > 10 )); then
                # if we can't find $done in the logs after a few
                # iterations, start printing out raw logs to help with
                # debugging
                if (( $idx < 12 ))
                then TAIL=1000
                else TAIL=10
                fi
                ($KUBECTL -n $ns logs $containers -l $selector --tail=$TAIL || exit 0)
            fi
            idx=$((idx + 1))
            sleep 4
        done
    done

    echo "✅ PASS run-controller run test $selector"

    $KUBECTL delete run $name -n $ns
    echo "✅ PASS run-controller delete test $selector"
}

# Checks if the the amount of unassigned tasks remaining is 0 and the number of tasks in the outbox is 6
function waitForUnassignedAndOutbox {
    local name=$1
    local ns=$2
    local expectedUnassignedTasks=$3
    local expectedNumInOutbox=$4
    local dataset=$5
    
    echo "$(tput setaf 2)🧪 Waiting for job to finish app=$selector ns=$ns$(tput sgr0)" 1>&2

    runNum=1
    while true
    do
        echo
        echo "Run #${runNum}: here's expected unassigned tasks=${expectedUnassignedTasks}"
        # here we use jq to sum up all of the unassigned annotations
        actualUnassignedTasks=$($KUBECTL -n $ns get dataset $dataset -o json | jq '.metadata.annotations | to_entries | map(select(.key | match("^jaas.dev/unassigned"))) | map(.value | tonumber) | reduce .[] as $num (0; .+$num)' || echo "there was an issue running the kubectl command😢")

        if ! [[ $actualUnassignedTasks =~ ^[0-9]+$ ]]; then echo "error: actualUnassignedTasks not a number: '$actualUnassignedTasks'"; fi
        if ! [[ $expectedUnassignedTasks =~ ^[0-9]+$ ]]; then echo "error: expectedUnassignedTasks not a number: '$expectedUnassignedTasks'"; fi

        echo "expected unassigned tasks=${expectedUnassignedTasks} and actual num unassigned=${actualUnassignedTasks}"
        if [[ "$actualUnassignedTasks" != "$expectedUnassignedTasks" ]]
        then
            echo "unassigned tasks should be ${expectedUnassignedTasks} but we got ${actualUnassignedTasks}"
            sleep 2
        else
            break
        fi

        runNum=$((runNum+1))
    done

    runIter=1
    while true
    do
        echo
        echo "Run #${runIter}: here's the expected num in Outboxes=${expectedNumInOutbox}"
        actualNumInOutbox=$($KUBECTL get queue -A -o custom-columns=INBOX:.metadata.annotations.codeflare\\.dev/outbox --no-headers)
        
        if ! [[ $actualNumInOutbox =~ ^[0-9]+$ ]]; then echo "error: actualNumInOutbox not a number: '$actualNumInOutbox'"; fi
        if ! [[ $expectedNumInOutbox =~ ^[0-9]+$ ]]; then echo "error: expectedNumInOutbox not a number: '$expectedNumInOutbox'"; fi

        if [[ "$actualNumInOutbox" != "$expectedNumInOutbox" ]]; then echo "tasks in outboxes should be ${expectedNumInOutbox} but we got ${actualNumInOutbox}"; sleep 2; else break; fi
        runIter=$((runIter+1))
    done

    echo "✅ PASS run-controller run test $name"

    $KUBECTL delete run $name -n $ns
    echo "✅ PASS run-controller delete test $name"
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
            ($KUBECTL -n $ns get run.codeflare.dev $name --show-kind --no-headers | grep $name || exit 0)
            sleep 4
        done
    done

    echo "✅ PASS run-controller run test $name"

    $KUBECTL delete run $name -n $ns
    echo "✅ PASS run-controller delete test $name"
}

function deploy {
    "$SCRIPTDIR"/deploy-tests.sh $1
}

function undeploy {
    if [[ -n "$2" ]]
    then kill $2 || true
    fi

    ("$SCRIPTDIR"/undeploy-tests.sh $1 || exit 0)
}

function watch {
    if [[ -n "$CI" ]]; then
        $KUBECTL get appwrapper --show-kind -n $CLUSTER_NAME-test -o custom-columns=NAME:.metadata.name,CONDITIONS:.status.conditions --watch &
        $KUBECTL get pod --show-kind -n $CLUSTER_NAME-test --watch &
    fi
    $KUBECTL get pod --show-kind -n $NAMESPACE_SYSTEM --watch &
    $KUBECTL get run --all-namespaces --watch &
    $KUBECTL get workerpool --watch --all-namespaces -o custom-columns=KIND:.kind,NAME:.metadata.name,STATUS:.metadata.annotations.codeflare\\.dev/status,MESSAGE:.metadata.annotations.codeflare\\.dev/message &
}
