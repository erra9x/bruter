#!/bin/bash

docker run -d -p 127.0.0.1:9000:9000 -e CLICKHOUSE_SKIP_USER_SETUP=1 --name default-clickhouse-server clickhouse/clickhouse-server
docker run -d -p 127.0.0.1:10000:9000 -e CLICKHOUSE_PASSWORD=12345678 --name password-clickhouse-server clickhouse/clickhouse-server

sleep 5
go run . clickhouse -u tests/usernames.txt -p tests/passwords.txt -t tmp/targets-clickhouse.txt -D

docker rm -f default-clickhouse-server
docker rm -f password-clickhouse-server
