test:
	@./go.test.sh

coverage:
	@./go.coverage.sh

test_fast:
	go test ./...

tidy:
	go mod tidy

priod:
	go build ./cmd/priod
	minikube image build --all -t ghcr.io/go-faster/prio:v0 .
	kubectl apply -f priod.yml
	kubectl -n kube-system rollout restart daemonset priod

example:
	minikube image build --all -f example.Dockerfile -t ghcr.io/go-faster/prio/example:v0 .
	go build ./cmd/prio-example
	kubectl apply -f prio-example.yml
	kubectl -n default rollout restart daemonset prio-example

.PHONY: tidy coverage test test_fast priod example