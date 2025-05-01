.PHONY: generate

generate-proto:
	protoc --go_out=generate --go-grpc_out=generate proto/medicine.proto

generate-openapi:
	oapi-codegen -package api -generate types,client openAPI/rest.yaml > generate/openapi.go

generate: generate-proto generate-openapi
