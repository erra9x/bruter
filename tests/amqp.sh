#!/usr/bin/env bash

export PASSWORD="password"

docker run -d --rm -p 5672:5672 -e RABBITMQ_DEFAULT_USER=admin -e RABBITMQ_DEFAULT_PASS=${PASSWORD} rabbitmq:alpine
sleep 10
