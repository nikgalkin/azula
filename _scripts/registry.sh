#!/bin/bash

git_root=$(git rev-parse --show-toplevel)
registry_files="$git_root/.dev/registry"
basic_auth_file="$git_root/.dev/auth/htpasswd"
registry_name="test_registry"
registry_port=5000
cmd="docker"

mkdir -p $registry_files

help(){
local cmd="./_scripts/registry"
echo "  ERROR! Command '$1' not supported
  Usage:
    $cmd u    - Up registry for dev purposes
    $cmd f    - Fill the local registry with some images (\$2 = count, \$3 = name)
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
  -p $registry_port:5000 \
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

fill(){
  local src_img=hello-world:latest
  if [[ $fill_max < 1 ]]; then
    fill_max=1
  fi
  if [[ -z $img_name ]]; then
    img_name=test
  fi
  $cmd pull $src_img
  for ((i=1; i<=$fill_max; i++)); do
    dst_img=127.0.0.1:$registry_port/$img_name-$i:v0.1
    $cmd tag $src_img $dst_img
    $cmd push $dst_img
    $cmd rmi $dst_img
  done
}

case $1 in
  up|u)
    up_registry ;;
  fill|f)
    fill_max=$2
    img_name=$3
    fill ;;
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
