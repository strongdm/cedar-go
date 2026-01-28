.PHONY: corpus-tests-json-schemas
corpus-tests-json-schemas: corpus-tests-json-schemas.tar.gz

corpus-tests-json-schemas.tar.gz: corpus-tests.tar.gz
	@echo "Generating JSON schemas from Cedar schemas (corpus tests)..."
	@go run ./cmd/corpus-json-schemas --in corpus-tests.tar.gz --out corpus-tests-json-schemas.tar.gz
	@echo "Done! Created corpus-tests-json-schemas.tar.gz"
