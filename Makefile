
export NAMESPACE=monitoring

docker_build_and_push:
	./scripts/docker-build-push.sh

get_file_base64:
	cat deployments/k8s/.dockerconfigjson | base64

switch_context:
	kubectl config use-context streaming-k8s-prod-default

create_namespace_env:
	kubectl create namespace ${NAMESPACE}

install_secret:
	kubectl apply -f deployment/gitlab-com-pull-secret.yml --namespace ${NAMESPACE}

install_app_debug:
	helm install std-app21-prom-extr \
		./deployment/helm/v8-1c-cluster-extractor \
		-f ./deployment/helm/ut-spb-values.yaml --namespace ${NAMESPACE}  --dry-run --debug

install_app:
#std-app21
	helm install std-app21-prom-extr \
		./deployment/helm/v8-1c-cluster-extractor \
		-f ./deployment/helm/std-app21-values.yaml --namespace ${NAMESPACE}

#std-app04
	helm install std-app04-prom-extr \
		./deployment/helm/v8-1c-cluster-extractor \
		-f ./deployment/helm/std-app04-values.yaml --namespace ${NAMESPACE}


#cfs-app01
	helm install cfs-app01-prom-extr \
		./deployment/helm/v8-1c-cluster-extractor \
		-f ./deployment/helm/cfs-app01-values.yaml --namespace ${NAMESPACE}

#std-app06
	helm install std-app06-prom-extr \
		./deployment/helm/v8-1c-cluster-extractor \
		-f ./deployment/helm/std-app06-values.yaml --namespace ${NAMESPACE}

#std-app11
	helm install std-app11-prom-extr \
		./deployment/helm/v8-1c-cluster-extractor \
		-f ./deployment/helm/std-app11-values.yaml --namespace ${NAMESPACE}

#std-app19
	helm install std-app19-prom-extr \
		./deployment/helm/v8-1c-cluster-extractor \
		-f ./deployment/helm/std-app19-values.yaml --namespace ${NAMESPACE}

#std-app20
	helm install std-app20-prom-extr \
		./deployment/helm/v8-1c-cluster-extractor \
		-f ./deployment/helm/std-app20-values.yaml --namespace ${NAMESPACE}

#std-app18
	helm install std-app18-prom-extr \
		./deployment/helm/v8-1c-cluster-extractor \
		-f ./deployment/helm/std-app18-values.yaml --namespace ${NAMESPACE}



inst:
	helm install std-app04-prom-extr \
		./deployment/helm/v8-1c-cluster-extractor \
		-f ./deployment/helm/std-app04-values.yaml --namespace ${NAMESPACE}


un_install_app:
	helm uninstall std-app21-prom-extr --namespace ${NAMESPACE}
	helm uninstall std-app04-prom-extr --namespace ${NAMESPACE}
	helm uninstall cfs-app01-prom-extr --namespace ${NAMESPACE}
	helm uninstall std-app06-prom-extr --namespace ${NAMESPACE}
	helm uninstall std-app11-prom-extr --namespace ${NAMESPACE}
	helm uninstall std-app19-prom-extr --namespace ${NAMESPACE}
	helm uninstall std-app20-prom-extr --namespace ${NAMESPACE}
	helm uninstall std-app18-prom-extr --namespace ${NAMESPACE}


#install: switch_context create_namespace_env install_secret install_app

install: switch_context install_secret install_app