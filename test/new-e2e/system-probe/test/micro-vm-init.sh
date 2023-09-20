#!/bin/bash

set -eo xtrace

GOVERSION=$1
RETRY_COUNT=$2
ARCH=$3
RUNNER_CMD="$(shift 3; echo "$*")"

KITCHEN_DOCKERS=/kitchen-docker

# Add provisioning steps here !
## Set go version correctly
eval $(gimme "$GOVERSION")
## Start docker
systemctl start docker


START=`date +%s`
## Load docker images
[ -d $KITCHEN_DOCKERS ] && find $KITCHEN_DOCKERS -maxdepth 1 -type f > /kitchen-dockers.txt
while read p; do
    docker load -i "$p" &
done < /kitchen-dockers.txt
wait
END=`date +%s`
echo "Docker load time:"
echo `expr $END - $START`

# VM provisioning end !

# Start tests
IP=$(ip route get 8.8.8.8 | grep -Po '(?<=(src ))(\S+)')
rm -rf /ci-visibility

CODE=0
/test-runner -retry $RETRY_COUNT $RUNNER_CMD || CODE=$?

pushd /ci-visibility
tar czvf testjson.tar.gz testjson
tar czvf junit.tar.gz junit

exit $CODE
