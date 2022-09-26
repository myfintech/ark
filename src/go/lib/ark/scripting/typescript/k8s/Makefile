generate:
	rm -rf dist || true
	npx openapi-typescript-codegen \
		--input https://raw.githubusercontent.com/kubernetes/kubernetes/master/api/openapi-spec/swagger.json \
 		--useUnionTypes \
		--exportCore false \
		--exportServices false \
		--output dist
.PHONY: generate