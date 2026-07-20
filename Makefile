test:
	@go test ./...

demo:
	@go run ./cmd/policy-playground run --scenario testdata/control/scenario.yaml --policies testdata/control/policies.yaml --out testdata/control/alerts.jsonl