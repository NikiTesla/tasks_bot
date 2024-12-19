//go:build tools
// +build tools

package tools

//go:generate go build -o ../bin/mockery github.com/vektra/mockery/v2

// Package tools contains go:generate commands for all project tools with versions stored in local go.mod file
// See https://github.com/golang/go/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module
import (
	_ "github.com/vektra/mockery/v2"
)
