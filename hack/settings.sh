#!/usr/bin/env bash

set -e
set -o pipefail

SETTINGS_SCRIPTDIR="$( dirname -- "$BASH_SOURCE"; )"

IMAGE_REGISTRY=ghcr.io # image registry part of image url
IMAGE_REPO=project-codeflare # image repo part of image url
VERSION=$("$SETTINGS_SCRIPTDIR"/version.sh) # will be used in image tag part of image url
NAMESPACE_SYSTEM=codeflare-system # namespace to use for system resources

PLA=$(grep name "$SETTINGS_SCRIPTDIR"/../platform/Chart.yaml | awk '{print $2}' | head -1)
IBM=$(grep name "$SETTINGS_SCRIPTDIR"/../watsonx_ai/Chart.yaml | awk '{print $2}' | head -1)

ARCH=${ARCH-$(uname -m)}
export KFP_VERSION=2.0.0

# Note: a trailing slash is required, if this is non-empty
IMAGE_REPO_FOR_BUILD=$IMAGE_REGISTRY/$IMAGE_REPO/

# for local testing
CLUSTER_NAME=${CLUSTER_NAME-jaas}

HELM_INSTALL_FLAGS="$HELM_INSTALL_FLAGS --set global.jaas.namespace.name=$NAMESPACE_SYSTEM --set global.jaas.context.name=kind-$CLUSTER_NAME --set global.image.registry=$IMAGE_REGISTRY --set global.image.repo=$IMAGE_REPO --set global.image.version=$VERSION"

# this will limit the platform to just api=workqueue
HELM_INSTALL_LITE_FLAGS="--set global.lite=true --set tags.default-user=false --set tags.defaults=false --set tags.full=false --set tags.core=true"

if lspci 2> /dev/null | grep -iq nvidia; then
    HAS_NVIDIA=true
else
    HAS_NVIDIA=false
fi

export KUBECTL="kubectl --context kind-${CLUSTER_NAME}"
export HELM="helm --kube-context kind-${CLUSTER_NAME}"

if [[ -z "$NO_GETOPTS" ]]
then
    while getopts "ltk:p" opt
    do
        case $opt in
            l) echo "Running up in lite mode"; export LITE=1; export HELM_INSTALL_FLAGS="$HELM_INSTALL_FLAGS $HELM_INSTALL_LITE_FLAGS"; continue;;
            t) RUNNING_TESTS=true; continue;;
            k) NO_KIND=true; export KUBECONFIG=${OPTARG}; continue;;
            p) PROD=true; continue;;
        esac
    done
    shift $((OPTIND-1))
fi
