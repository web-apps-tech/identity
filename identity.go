//go:generate go run internal/cmd/genidentity/genidentity.go -f spec/v1/spec.json -o identity_gen.go
//go:generate go run internal/cmd/genhandler/genhandler.go -f spec/v1/spec.json -o handler.go

package identity
