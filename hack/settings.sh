#!/usr/bin/env bash

set -e
set -o pipefail

SETTINGS_SCRIPTDIR="$( dirname -- "$BASH_SOURCE"; )"

PLA=$(grep name "$SETTINGS_SCRIPTDIR"/../platform/Chart.yaml | awk '{print $2}' | head -1)
IBM=$(grep name "$SETTINGS_SCRIPTDIR"/../watsonx_ai/Chart.yaml | awk '{print $2}' | head -1)

ARCH=${ARCH-$(uname -m)}
export KFP_VERSION=2.0.0

# Note: a trailing slash is required, if this is non-empty
IMAGE_REPO=ghcr.io/project-codeflare/

# for local testing
CLUSTER_NAME=${CLUSTER_NAME-jaas}

# namespaces to use
NAMESPACE_SYSTEM=codeflare-system
HELM_INSTALL_FLAGS="$HELM_INSTALL_FLAGS --set global.jaas.namespace.name=$NAMESPACE_SYSTEM --set global.jaas.context.name=kind-$CLUSTER_NAME"

if lspci 2> /dev/null | grep -iq nvidia; then
    HAS_NVIDIA=true
else
    HAS_NVIDIA=false
fi

export KUBECTL="kubectl --context kind-${CLUSTER_NAME}"
export HELM="helm --kube-context kind-${CLUSTER_NAME}"

if [[ -z "$NO_GETOPTS" ]]
then
    while getopts "ltk:" opt
    do
        case $opt in
            l) echo "Running up in lite mode"; export LITE=1; export HELM_INSTALL_FLAGS="$HELM_INSTALL_FLAGS --set global.lite=true --set tags.examples=false --set tags.defaults=false --set tags.full=false --set tags.core=true"; continue;;
            t) RUNNING_TESTS=true; continue;;
            k) NO_KIND=true; export KUBECONFIG=${OPTARG}; continue;;
        esac
    done
    shift $((OPTIND-1))
fi

if [[ -z "$NO_KIND" ]]; then
    VERSION=dev
else
    VERSION=0.0.1 # FIXME
fi


