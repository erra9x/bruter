#!/bin/bash

docker run -d -p 127.0.0.1:9001:21 -e FTP_USER=anonymous -e FTP_PASS=anonymous --name anonymous-ftp-server fauria/vsftpd
docker run -d -p 127.0.0.1:10001:21 -e FTP_USER=ftpuser -e FTP_PASS=12345678 --name password-ftp-server fauria/vsftpd

sleep 5
go run . ftp -u tests/usernames.txt -p tests/passwords.txt -t tmp/targets-ftp.txt -D

docker rm -f anonymous-ftp-server
docker rm -f password-ftp-server
