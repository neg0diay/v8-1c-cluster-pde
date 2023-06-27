#!/bin/bash
#chmod +x script.sh

#docker login registry.git.a.kluatr.ru
#docker build -t registry.git.a.kluatr.ru/kafka_streaming/object1skafkastreamer:latest .
#docker push registry.git.a.kluatr.ru/kafka_streaming/object1skafkastreamer:latest

docker login registry.gitlab.com
docker build -t registry.gitlab.com/prometheus_exporter/v8-1c-cluster-pde .
docker push registry.gitlab.com/prometheus_exporter/v8-1c-cluster-pde