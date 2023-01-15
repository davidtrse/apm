# Granfana, Loki and Tempo

## Install docker driver for loki in each VPS|EC2|OS
```
$ docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
```
## Build hello-app docker image
Step 1 `make build`

## Start docker container
Step 2 `make run`

## Browse
http://localhost:8080/hello/minh

## Check Granfana Loki logs
http://localhost:3000

## Check Granfana Tempo logs
http://localhost:3000