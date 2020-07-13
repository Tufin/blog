#!/bin/bash

set -euf -o pipefail

function build_and_upload_to_dockerhub() {

  local docker_user=$1
  local docker_pass=$2
  local company=$3
  local image=$4
  local tag=$5

  docker build -t "$image" -f Dockerfile."$image" .
  echo "$docker_pass" | docker login -u "$docker_user" --password-stdin
  docker tag "$image" "$company"/"$image"
  docker tag "$image" "$company"/"$image":"$tag"
  docker push "$company"/"$image"
}

build_and_upload_to_dockerhub "$1" "$2" "$3" "$4" "$5"