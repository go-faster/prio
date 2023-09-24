#!/bin/bash

set -ex

minikube start --iso-url=file:///src/minikube/out/minikube-amd64.iso --nodes 2 --driver kvm2 --cpus 8

helm upgrade --install \
  --namespace tetragon --create-namespace \
  --values tetragon.yml \
  tetragon cilium/tetragon

