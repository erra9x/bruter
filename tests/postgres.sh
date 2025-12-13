#!/bin/bash

docker run --name postgres-password -e POSTGRES_PASSWORD=12345678 -p 5432:5432 -d postgres

sleep 5
go run . postgres -u postgres -p tests/passwords.txt -t 127.0.0.1 -D

docker rm -f postgres-password
