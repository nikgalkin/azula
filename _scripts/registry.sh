#!/bin/bash

git_root=$(git rev-parse --show-toplevel)
registry_files="$git_root/.dev/registry"
basic_auth_file="$git_root/.dev/auth/htpasswd"
registry_name="test_registry"
mkdir -p $registry_files
cmd="docker"

help(){
local cmd="./_scripts/registry"
echo "  ERROR! Command $1 not supported
  Usage:
    $cmd u    - Up registry for dev purposes
    $cmd ps   - Get status of container
    $cmd d    - Stop and destroy container
    $cmd l    - Follow the logs
    $cmd stop - Just stop container
"
}

gen_basic_auth() {
if [[ ! -f $basic_auth_file ]]; then
  echo "-- Creating basic auth file ($basic_auth_file)"
  mkdir -p $(dirname $basic_auth_file)
  local cmd="htpasswd" args="-Bbn testuser testpassword"
  if command -v $cmd &> /dev/null; then
    $cmd $args > $basic_auth_file
  else
    docker run --entrypoint $cmd httpd:2 $args > $basic_auth_file
  fi
fi
}

up_registry() {
gen_basic_auth
docker run -d \
  -p 5000:5000 \
  --restart=always \
  --name $registry_name \
  -v $(dirname $basic_auth_file):/auth \
  -e "REGISTRY_AUTH=htpasswd" \
  -e "REGISTRY_AUTH_HTPASSWD_REALM=$registry_name" \
  -e "REGISTRY_AUTH_HTPASSWD_PATH=/auth/htpasswd" \
  -e "REGISTRY_STORAGE_DELETE_ENABLED=true" \
  registry:2
}

down(){
  echo "Stopping and removing $registry_name ..."
  $cmd stop $registry_name
  $cmd rm $registry_name
}

case $1 in
  up|u)
    up_registry ;;
  ps)
    $cmd ps ;;
  stop|s)
    $cmd stop $registry_name ;;
  start)
    $cmd start $registry_name ;;
  log|logs|l)
    $cmd logs $registry_name -f -n 10 ;;
  down|d)
    down ;;
  *) help $1 ;;
esac
