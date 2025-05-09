.PHONY: generate

generate-proto:
	protoc --go_out=generate --go-grpc_out=generate proto/medicine.proto

generate-openapi:
	oapi-codegen -generate types -package api openAPI/rest.yaml > api/types.go
	oapi-codegen -generate gin -package api openAPI/rest.yaml > api/server.go
	oapi-codegen -generate spec -package api openAPI/rest.yaml > api/spec.go

generate-types:
	mkdir -p api
	oapi-codegen -generate types -package api openAPI/rest.yaml > api/types.go

generate-server:
	oapi-codegen -generate gin -package api openAPI/rest.yaml > api/server.go

generate-client:
	oapi-codegen -generate client -package api openAPI/rest.yaml > api/client.go

generate: generate-proto generate-openapi