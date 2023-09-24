#!/bin/bash

minikube start --nodes 2 --driver kvm2 --container-runtime=containerd  --network-plugin=cni --cni=false --cpus 8

helm upgrade --install --version 1.14.2 \
   --namespace cilium --create-namespace \
   --set image.pullPolicy=IfNotPresent \
   --set ipam.mode=kubernetes \
   cilium cilium/cilium

helm upgrade --install \
  --namespace cilium --create-namespace \
  --values tetragon.yml \
  tetragon cilium/tetragon

cilium status --wait
