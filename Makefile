test:
	@./go.test.sh

coverage:
	@./go.coverage.sh

test_fast:
	go test ./...

tidy:
	go mod tidy

priod:
	kubectl apply -f priod.yml
	go build ./cmd/priod
	docker build -t ghcr.io/go-faster/prio:v0 .
	kind load docker-image ghcr.io/go-faster/prio:v0
	kubectl -n kube-system rollout restart daemonset priod

example:
	kubectl apply -f prio-example.yml
	go build ./cmd/prio-example
	docker build -f example.Dockerfile -t ghcr.io/go-faster/prio/example:v0 .
	kind load docker-image ghcr.io/go-faster/prio/example:v0
	kubectl -n default rollout restart daemonset prio-example

.PHONY: tidy coverage test test_fast priod example