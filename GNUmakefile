default: testacc

# Run acceptance tests
.PHONY: testacc
testacc:
	TF_ACC=1 go test -count=1 ./... -v $(DBSNAPPER_AUTHTOKEN) $(DBSNAPPER_BASE_URL) -timeout 120m
