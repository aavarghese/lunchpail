#!/usr/bin/env bash

set -x
set -e
set -o pipefail

SCRIPTDIR=$(cd $(dirname "$0") && pwd)

uid="$1"
name="$2"
namespace="$3"
part_of="$4"
step="$5" # if part of enclosing sequence
run_id="$6"
image="$7"
command="$8"
subPath="$9"
nWorkers="${10}"
cpu="${11}"
memory="${12}"
gpu="${13}"
env="${14}"
volumes="${15}"
volumeMounts="${16}"
envFroms="${17}"
expose="${18}"
securityContext="${19}"
containerSecurityContext="${20}"
component="${21}"
enclosing_run="${22}"
workdir_repo="${23}"
workdir_pat_user="${24}"
workdir_pat_secret="${25}"
workdir_cm_data="${26}"
workdir_cm_mount_path="${27}"

# Helm's dry-run output will go to this temporary file
DRY=$(mktemp)
echo "Dry running to $DRY" 1>&2

helm install --dry-run --debug $run_id "$SCRIPTDIR"/shell/ -n ${namespace} \
     --set kind=job \
     --set uid=$uid \
     --set name=$name \
     --set partOf=$part_of \
     --set component=$component \
     --set enclosingRun=$enclosing_run \
     --set enclosingStep=$step \
     --set image=$image \
     --set namespace=$namespace \
     --set command="$command" \
     --set subPath=$subPath \
     --set workers.count=$nWorkers \
     --set workers.cpu=$cpu \
     --set workers.memory=$memory \
     --set workers.gpu=$gpu \
     --set volumes=$volumes \
     --set volumeMounts=$volumeMounts \
     --set envFroms=$envFroms \
     --set env="$env" \
     --set expose=$expose \
     --set mcad.enabled=${MCAD_ENABLED:-false } \
     --set rbac.runAsRoot=$RUN_AS_ROOT \
     --set rbac.serviceaccount="$USER_SERVICE_ACCOUNT" \
     --set securityContext=$securityContext \
     --set containerSecurityContext=$containerSecurityContext \
     --set workdir.repo=$workdir_repo \
     --set workdir.pat.user=$workdir_pat_user \
     --set workdir.pat.secret=$workdir_pat_secret \
     --set workdir.cm.data=$workdir_cm_data \
     --set workdir.cm.mount_path=$workdir_cm_mount_path \
     --set lunchpail.image.registry=$IMAGE_REGISTRY \
     --set lunchpail.image.repo=$IMAGE_REPO \
     --set lunchpail.image.version=$IMAGE_VERSION \
    | awk '$0~"Source: " {on=1} on==2 { print $0 } on==1{on=2}' \
          > $DRY

retries=20
while ! kubectl apply -f $DRY
do
    ((--retries)) || exit 1

    echo "Retrying kubectl apply"
    sleep 1
done

rm -f $DRY
