#!/usr/bin/env bash

export PASSWORD="password"

docker run -d --name rabbitmq-default --rm -p 127.0.0.1:5672:5672 -e RABBITMQ_DEFAULT_USER=admin -e RABBITMQ_DEFAULT_PASS=${PASSWORD} rabbitmq:alpine
sleep 5

go run . amqp -u tests/usernames.txt -p tests/passwords.txt -t 127.0.0.1 -D

docker rm -f rabbitmq-default
