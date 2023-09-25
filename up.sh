#!/bin/bash

set -ex

minikube start --iso-url=https://github.com/go-faster/minikube/releases/download/v1.31.2-5.15/minikube-amd64.iso --nodes 2 --driver kvm2 --cpus 8

helm upgrade --install \
  --namespace tetragon --create-namespace \
  --values tetragon.yml \
  tetragon cilium/tetragon

make priod example