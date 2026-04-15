include .env
export

.PHONY: dry-run diff apply

dry-run:
	@echo "=== k8s/deployment ===" && envsubst < k8s/deployment.yaml.tpl
	@echo "=== k8s/service ===" && cat k8s/service.yaml
	@echo "=== k8s/ingress ===" && envsubst < k8s/ingress.yaml.tpl

diff:
	-@envsubst < k8s/deployment.yaml.tpl | kubectl diff -f -
	-@cat k8s/service.yaml | kubectl diff -f -
	-@envsubst < k8s/ingress.yaml.tpl | kubectl diff -f -

apply:
	@envsubst < k8s/deployment.yaml.tpl | kubectl apply -f -
	@cat k8s/service.yaml | kubectl apply -f -
	@envsubst < k8s/ingress.yaml.tpl | kubectl apply -f -
