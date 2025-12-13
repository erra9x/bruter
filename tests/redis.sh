#!/bin/bash

docker run -d --name redis-default -p 127.0.0.1:6379:6379 redis:latest
docker run -d --name redis-password -p 127.0.0.1:10000:6379 redis redis-server --requirepass "12345678"

sleep 5
go run . redis -u tests/usernames.txt -p tests/passwords.txt -t tmp/targets-redis.txt -D

docker rm -f redis-default
docker rm -f redis-password
