test-cov:
	gopherbadger -md="README.md" -png=false

lint:
	golangci-lint --exclude=copylocks run --timeout 30s