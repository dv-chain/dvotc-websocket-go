test-cov:
	gopherbadger -md="README.md" -png=false

lint:
	golangci-lint run --timeout 30s