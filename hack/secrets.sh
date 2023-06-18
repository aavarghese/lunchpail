#!/usr/bin/env bash

SETTINGS_SCRIPTDIR="$( dirname -- "$BASH_SOURCE"; )"
. "$SETTINGS_SCRIPTDIR"/my.secrets.sh

HELM_SECRETS="--set codeflare-ibm-internal.github_ibm_com.secret.user=$AI_FOUNDATION_GITHUB_USER --set codeflare-ibm-internal.github_ibm_com.secret.pat=$AI_FOUNDATION_GITHUB_PAT --set global.s3AccessKey=codeflarey --set global.s3SecretKey=codeflarey --set global.buckets.test=internal-test-bucket"
