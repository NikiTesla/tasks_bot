#!/bin/bash

APP_LABEL="tasks_bot"

function log {
    level=$1
    msg=$2

    printf "$(date +"%Y-%m-%d %H:%M:%S") level=$level msg=$msg\n"
}

log info "removing previous build (containers, images)"
for container_id in $(docker ps -a | grep $APP_LABEL | awk '{print $1}'); do 
    docker stop $container_id || true
    docker rm $container_id || true
done

for docker_image in $(docker image ls | grep $APP_LABEL | awk '{print $3}'); do
    docker rmi $docker_image || true
done

log info "building and running"
docker compose up -d --build
