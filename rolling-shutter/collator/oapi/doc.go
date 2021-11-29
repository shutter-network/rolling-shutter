// package oapi implements the collator's openapi interface
package oapi

//go:generate oapi-codegen --generate types,chi-server,spec -o oapi.gen.go --package oapi oapi.yaml
