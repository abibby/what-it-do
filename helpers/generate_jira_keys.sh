#!/usr/bin/env bash

mkdir jira_credentials

openssl genrsa -out jira_credentials/privatekey.pem 1024
openssl req -newkey rsa:1024 -x509 -key jira_credentials/privatekey.pem -out jira_credentials/publickey.cer -days 365
openssl pkcs8 -topk8 -nocrypt -in jira_credentials/privatekey.pem -out jira_credentials/privatekey.pcks8
openssl x509 -pubkey -noout -in jira_credentials/publickey.cer  > jira_credentials/publickey.pem
