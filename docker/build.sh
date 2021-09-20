#! /usr/bin/env sh
echo "Static building CGRateS..."

GIT_LAST_LOG=$(git log -1 | tr -d "'")

GIT_TAG_LOG=$(git tag -l --points-at HEAD)

if [ ! -z "$GIT_TAG_LOG" ]
then
    GIT_LAST_LOG=""
fi

GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ansihook -a -ldflags '-extldflags "-f no-PIC -static"' -tags 'osusergo netgo static_build' github.com/cgrates/ansihook
cr=$?