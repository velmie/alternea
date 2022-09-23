package bootstrap

import (
	"fmt"
)

// Factory base factory interface
type Factory[T any] interface {
	Create(name string, config Config) (instance T, err error)
}

// FactoryMap allows using multiple factory by name
type FactoryMap[T any] map[string]Factory[T]

func (f FactoryMap[T]) Create(name string, config Config) (instance T, err error) {
	if factory, ok := f[name]; ok {
		return factory.Create(name, config)
	}
	var empty T
	return empty, fmt.Errorf("unknown name '%s'", name)
}

// FactoryFunc allows using functions as factories
type FactoryFunc[T any] func(name string, config Config) (T, error)

func (f FactoryFunc[T]) Create(name string, config Config) (T, error) {
	return f(name, config)
}
