#!/bin/bash
#chmod +x script.sh

docker login registry.gitlab.com
docker build -t registry.gitlab.com/prometheus_exporter/v8-1c-cluster-pde .
docker push registry.gitlab.com/prometheus_exporter/v8-1c-cluster-pde