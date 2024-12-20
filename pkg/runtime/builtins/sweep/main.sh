#!/bin/sh

set -e
#set -o pipefail

echo "Started TaskDispatcher min=$__LUNCHPAIL_SWEEP_MIN max=$__LUNCHPAIL_SWEEP_MAX step=$__LUNCHPAIL_SWEEP_STEP"

trap "$LUNCHPAIL_EXE queue done" EXIT

# test injected values from -f values.yaml
# taskprefix2 can be used to test that e.g. numerical values are processed correctly
if [ -n "$taskprefix" ]
then taskprefix=${taskprefix}${taskprefix2}
else taskprefix=task
fi

if [ -n "$DEBUG" ]
then
    echo "got value taskprefix=$taskprefix"
    printenv
fi

# how many tasks we've injected so far; it is useful to keep the
# filename of tasks consistent, so that tests can look for a
# deterministic set of tasks
idx=0

for parameter_value in $(seq $__LUNCHPAIL_SWEEP_MIN $__LUNCHPAIL_SWEEP_STEP $__LUNCHPAIL_SWEEP_MAX)
do
    if [ "$parameter_value" != "$__LUNCHPAIL_SWEEP_MIN" ]
    then sleep ${__LUNCHPAIL_INTERVAL-5}
    fi

    task=/tmp/${taskprefix}.${idx}.txt
    idx=$((idx + 1))

    echo -n ${parameter_value} > $task

    if [ -n "$__LUNCHPAIL_WAIT" ]
    then waitflag="--wait --ignore-worker-errors"
    fi

    if [ -n "$__LUNCHPAIL_VERBOSE" ]
    then verboseflag="--verbose"
    fi

    if [ -n "$__LUNCHPAIL_DEBUG" ]
    then debugflag="--debug"
    fi

    # If we were asked to wait, then `queue add file` will exit with the
    # exit code of the underlying worker. Here, we intentionally
    # ignore any errors from the task.
    $LUNCHPAIL_EXE queue add file $task $waitflag $verboseflag $debugflag --next
    rm -f "$task"
done

$LUNCHPAIL_EXE queue done
