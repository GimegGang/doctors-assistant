// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.16.3 DO NOT EDIT.
package openAPI

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/9RWTW/jNhD9K8S0R9VyPloEurkoUBho0qDdW2AYXHFsMyuRCjlyYhj67wtSX5YlGQ6c",
	"HHKJEoZ88zjvzQz3EOs00woVWYj2YOMNptz/OhPi/3iDIk/wP7SZVhbdcmZ0hoYk+k1SuJ8rbVJOEIFU",
	"9MctBEC7DMs/cY0GiqJZ0t+fMSYoAvgbqca34wFsvWVZhpKEqT0rZrPCjeG7YQ73KGQs1UBcwQk7YdzC",
	"byRTbENZMlKtHc6ZaQhA8dTD9hDqe3bu2NvVvVEAxH9ItV6K3HCSWh2TuLmGAFKpZJqnEF0N5sggpxQV",
	"XQKSWzTLkRycOlkEYPAllwYFRE9lcvqXGiTZBl0M6PqAb/TNw/SVHZXAi9v/R983bkmqlfY+QRsbmZWJ",
	"g9njnK20YSlXfC3VmqWVwVjjY8aVYGR47NgxhW/Eygtbd1FJzgKNL1ldIWz2OIcAtmhsGelqMp1MHWud",
	"oeKZhAhuJtPJFQSQcdr4m4YOfVmjR3tYI7mPy4bP4lxA5OqwTZf15w1PkdBYiJ72IF24lxzNDmr/Nrk/",
	"1I9MjkHVQN7vhIWDKpuAp3o9nbpPrBWh8qx5liUy9rzDZ1v6tI3W1MyvBlcQwS9h29jCqquFB7bot4ci",
	"OBLzH2mJ6VVXoyKA25Jad/O9tNYJqg2TassTKViVJNbkszx82z/8oGt8ttK5Em7j70NR5orQKJ4wi2aL",
	"hqExuqwjm6cpN7tSz5Jy470a+1XSRipGG2TO6yxDI7XwFef88eTLhtVGWDjY8LAvjdmnNul53qkRP9Y/",
	"wVcx6il/NuNowI1NJxBIXCbvcuKBKmMWbOCVpg8yoT1m3BqtGfywKALItB0w1qO2h85y+qGlP7XYfVK+",
	"uw4pPlHnoafVKcm5ECiYzeMYrV3lSbIbVX9eSV6l6yIRZ0IwzhS+9sfYiJaHHcOe0zK+7Lw5Je/gy/bE",
	"gGkKZf7X5QPmopLlSdJh4x8z3AcaU9xheNBSvaMb6pgnTOAWE525B1xFwD3fTAIRbIiyKAwTt2+jLUV3",
	"07spOHWqYMeI/9ZOKsfZwAOrdU3L042HURzSbD02M1uwzmwsFsXPAAAA//+qZoOIOg0AAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %w", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %w", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	res := make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	resolvePath := PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		pathToFile := url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
