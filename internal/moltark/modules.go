package moltark

import (
	"fmt"

	"go.starlark.net/starlark"
)

type localModule interface {
	Namespace() starlark.Value
	BuildComponents(model DesiredModel) ([]ComponentSpec, error)
}

type localModuleFactory func(*desiredModelBuilder) localModule

func (b *desiredModelBuilder) registerLocalModules() {
	b.moduleFactories = map[string]localModuleFactory{
		ModuleSourcePython: newPythonModuleRuntime,
		ModuleSourceUV:     newUVModuleRuntime,
		ModuleSourceCore:   newCoreModuleRuntime,
	}
	b.moduleBuildOrder = []string{
		ModuleSourcePython,
		ModuleSourceUV,
		ModuleSourceCore,
	}
}

func (b *desiredModelBuilder) loadLocalModule(source string) (localModule, error) {
	if module, ok := b.modules[source]; ok {
		return module, nil
	}

	factory, ok := b.moduleFactories[source]
	if !ok {
		return nil, fmt.Errorf("unknown module source %q", source)
	}

	module := factory(b)
	b.modules[source] = module
	return module, nil
}
