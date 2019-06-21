#!/bin/bash
docker ps -a -q | xargs docker rm -f ; docker rmi fakestorage ; docker build -t fakestorage . ; docker run -d --name fakestorage -p 4443:4443 -v $(pwd)/data:/data fakestorage