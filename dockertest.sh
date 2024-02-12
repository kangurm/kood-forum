#!/bin/bash

docker build -t forum -f docker/dockerfile .
docker run -p 8080:8080 --name forumcontainer forum