#!/usr/bin/env bash

kubectl delete --ignore-not-found=true -f manifests/ -f manifests/setup