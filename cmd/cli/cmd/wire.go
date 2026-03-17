//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.
package cmd

import (
	"github.com/google/wire"
	"github.com/lynx-go/lynx"
)

func wireCLIContext(app lynx.Lynx) (*CLIContext, func(), error) {
	panic(wire.Build(ProviderSet, NewCLIContext))
}
