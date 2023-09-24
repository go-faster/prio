#!/bin/bash

kind create cluster --config=kind-config.yaml
docker pull quay.io/cilium/cilium:v1.14.2
kind load docker-image quay.io/cilium/cilium:v1.14.2

helm install cilium cilium/cilium --version 1.14.2 \
   --namespace kube-system \
   --set image.pullPolicy=IfNotPresent \
   --set ipam.mode=kubernetes
helm upgrade --install -n kube-system --values tetragon.yml tetragon cilium/tetragon

cilium status --wait
