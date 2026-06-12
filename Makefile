.PHONY: generate test test-integration check-integration refresh-fixtures lint fmt coverage tidy examples

# generate re-applies the OpenAPI fixes (if a script is present) and regenerates
# the typed client via the //go:generate directives.
generate:
	@if [ -x scripts/fix-openapi.sh ]; then ./scripts/fix-openapi.sh; \
	elif [ -x scripts/convert-openapi.sh ]; then ./scripts/convert-openapi.sh; fi
	go generate ./...

test:
	go test -race -count=1 ./...

test-integration:
	go test -race -count=1 -tags=integration ./...

# check-integration verifies every exported Client method has a per-endpoint
# TestIntegration<Method> live test (see scripts/check-integration-coverage.sh).
check-integration:
	./scripts/check-integration-coverage.sh

# refresh-fixtures re-copies the live demo/examples recordings over the committed
# golden fixtures in testdata/. The deterministic fixture tests
# (TestXMLCompleteness / TestXMLGoldenRoundTrip / TestParse*) read the committed
# testdata copies, never demo/examples, so re-running the demo does not change
# test behavior until the goldens are refreshed here and the diff is reviewed.
# Run it (after `make examples`) ONLY when a human intends to update the goldens.
refresh-fixtures:
	cp demo/examples/search_patents/response.xml   testdata/patent_search.xml
	cp demo/examples/get_patent_info/response.xml  testdata/patent_info.xml
	cp demo/examples/search_trademarks/response.xml testdata/trademark_search.xml
	cp demo/examples/get_trademark_info/response.xml testdata/trademark_info.xml
	cp demo/examples/search_designs/response.xml    testdata/design_search.xml
	cp demo/examples/get_design_info/response.xml   testdata/design_info.xml
	@echo "testdata/*.xml updated from demo/examples; review the diff before committing."

lint:
	golangci-lint run

fmt:
	gofmt -w .

coverage:
	go test -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out | tail -n 1

tidy:
	go mod tidy

# examples runs the demo against the live API to refresh demo/examples (recorded
# request/response pairs). Requires the API credentials in the environment.
# The weekly response-watch workflow runs this and diffs the result.
examples:
	cd demo && go run .
