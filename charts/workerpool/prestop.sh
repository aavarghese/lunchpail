#!/usr/bin/env bash

SCRIPTDIR=$(cd $(dirname "$0") && pwd)
. "$SCRIPTDIR"/settings.sh

PATH="${HOME}/rclone-v1.66.0-linux-$(uname -m)/":$PATH

echo "DEBUG Marking worker as done..."

# deregister ourselves as a live worker
rclone --config $config delete $alive

# register ourselves as a dead worker
rclone --config $config touch $dead

echo "INFO This worker is shutting down $(echo $POD_NAME | sed -E "s#^${RUN_NAME}-##")!"
