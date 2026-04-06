package module

import (
	"fmt"

	"github.com/ophidiarium/moltark/internal/model"
	"go.starlark.net/starlark"
)

type localModule interface {
	Namespace() starlark.Value
	BuildComponents(model model.DesiredModel) ([]model.ComponentSpec, error)
}

type localModuleFactory func(*desiredModelBuilder) localModule

func (b *desiredModelBuilder) registerLocalModules() {
	b.moduleFactories = map[string]localModuleFactory{
		model.ModuleSourcePython: newPythonModuleRuntime,
		model.ModuleSourceUV:     newUVModuleRuntime,
		model.ModuleSourceCore:   newCoreModuleRuntime,
	}
	b.moduleBuildOrder = []string{
		model.ModuleSourcePython,
		model.ModuleSourceUV,
		model.ModuleSourceCore,
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
