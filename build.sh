#!/bin/bash

set -e

#DoNotBuildImage=0
Publish=0

args=`getopt pd $*`
set -- $args
for i; do
    case "$i" in
        -p)
            Publish=1
            shift ;;
        --) shift; break ;;
    esac
done

VERSION=0.1
CONTAINER=devops

echo Build Sources
go build

echo Building image $CONTAINER:$VERSION
docker build -t $CONTAINER:$VERSION .
# docker tag $CONTAINER:$VERSION $DOCKER_ID_USER/$CONTAINER
docker tag $CONTAINER:$VERSION abhinababasu.azurecr.io/$CONTAINER

echo Done ....

if [ $Publish == "1" ]; then
    az login
    az acr login --name abhinababasu
    docker push abhinababasu.azurecr.io/devops:latest
fi
echo
