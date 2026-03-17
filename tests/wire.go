//go:build wireinject
// +build wireinject

package tests

import (
	"github.com/google/wire"
	"github.com/lynx-go/lynx"
)

func wireTestingSuite(app lynx.Lynx) (*TestingSuite, func(), error) {
	panic(wire.Build(ProviderSet))
}
