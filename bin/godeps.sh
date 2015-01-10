#!/bin/bash
go get -v -u github.com/tools/godep
rsync -avzv --exclude "$WERCKER_SOURCE_DIR" "$GOPATH/" "$WERCKER_CACHE_DIR/go-pkg-cache/"
cp -R $GOPATH/src/github.com/Tapglue $GOPATH/_src
godep restore
mv $GOPATH/_src $GOPATH/src/github.com/tapglue
