#!/bin/bash

VERSION=0.1
CONTAINER=devops

echo Build Sources
go build 

echo Building image $CONTAINER:$VERSION
docker build -t $CONTAINER:$VERSION .
docker tag $CONTAINER:$VERSION $DOCKER_ID_USER/$CONTAINER
docker tag $CONTAINER:$VERSION abhinababasu.azurecr.io/$CONTAINER

echo Done ....

echo
