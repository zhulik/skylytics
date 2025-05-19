package api

import (
	_ "embed"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/samber/lo"
)

//go:embed openapi.json
var openAPISpec []byte

var (
	spec = lo.Must(openapi3.NewLoader().LoadFromData(openAPISpec))
)
