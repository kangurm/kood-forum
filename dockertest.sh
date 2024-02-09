#!/bin/bash

docker build -t forum .
docker run -p 8080:8080 forum