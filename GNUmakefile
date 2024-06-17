default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test ./... -v $(DBSNAPPER_AUTHTOKEN) $(DBSNAPPER_BASE_URL) -timeout 120m
