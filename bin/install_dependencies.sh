#!/usr/bin/env bash

set -e
set -o pipefail

export PATH=/home/ubuntu/.gimme/versions/go1.7.linux.amd64/bin:${PATH}
CWD=`pwd`

echo "Installing gimme"
sudo curl -sL -o /usr/local/bin/gimme https://raw.githubusercontent.com/travis-ci/gimme/master/gimme
sudo chmod +x /usr/local/bin/gimme

echo "Installing go 1.7"
eval "$(GIMME_GO_VERSION=1.7 gimme)"

echo "Installing go dependencies"
go get github.com/tools/godep github.com/axw/gocov/gocov github.com/matm/gocov-html gopkg.in/check.v1

declare -a STATIC_COMPONENTS=( "dashboard" "website" )
for STATIC_COMPONENT in "${STATIC_COMPONENTS[@]}"
do
    echo "Installing ${STATIC_COMPONENT} dependencies"
    mkdir -p /home/ubuntu/.${STATIC_COMPONENT}_node_modules
    cd ${CWD}/${STATIC_COMPONENT}
    mv /home/ubuntu/.${STATIC_COMPONENT}_node_modules node_modules
    npm install
done

for STATIC_COMPONENT in "${STATIC_COMPONENTS[@]}"
do
    cd ${CWD}/${STATIC_COMPONENT}
    cp -R node_modules /home/ubuntu/.${STATIC_COMPONENT}_node_modules
done
