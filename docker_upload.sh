#!/usr/bin/env bash

TAG=$(git tag | sort -V | tail -1)
VERSION="${TAG:1}"

echo "version: ${VERSION}"

# tag version
docker tag z0rr0/tgtpgybot:latest z0rr0/tgtpgybot:${VERSION}

# push
docker push z0rr0/tgtpgybot:${VERSION}
docker push z0rr0/tgtpgybot:latest
