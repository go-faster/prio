#!/bin/bash

set -ex

minikube start --nodes 2 --driver kvm2 --cpus 8

helm upgrade --install \
  --namespace tetragon --create-namespace \
  --values tetragon.yml \
  tetragon cilium/tetragon

