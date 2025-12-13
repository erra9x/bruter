#!/bin/bash

docker run -d --name ssh-password -p 22:2222 -e USER_NAME=admin -e USER_PASSWORD=12345678 -e PASSWORD_ACCESS=true lscr.io/linuxserver/openssh-server:latest

sleep 5
go run . ssh -u admin -p tests/passwords.txt -t 127.0.0.1 -D -c 1 --delay 4s

docker rm -f ssh-password
