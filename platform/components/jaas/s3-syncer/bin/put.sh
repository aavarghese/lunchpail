#!/usr/bin/env sh

config=$1
remote=$2
local=$3
outbox=$4
justonce=$5

mkdir -p $local/$outbox

size=0

function report_size {
    echo "codeflare.dev queue $(basename $local) $outbox $size"
}

# initial report_size
report_size

if [[ -n "$DEBUG" ]]; then
    PROGRESS="--progress"
fi

echo "[workerpool s3-syncer-put $(basename $local)] Starting rclone put remote=$remote local=$local/$outbox"
while true; do
    if [[ -d $local/$outbox ]]; then
        rclone --config $config sync $PROGRESS --create-empty-src-dirs $local/$outbox $remote/$outbox
 
        new_size=$(ls $local/$outbox | wc -l)
        if [[ $size != $new_size ]]; then
            size=$new_size
            report_size
        fi
    fi

    if [[ -n "$justonce" ]]
    then break
    fi

    sleep 5
done
