package bootstrap

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/qntfy/kazaam"

	"github.com/velmie/alternea/manipulation"
)

const (
	RemapperReferenceName = "remapper"
	RemapperKazaam        = "kazaam" // https://github.com/qntfy/kazaam
	RemapperNoOp          = "noOp"
)

var Remapper = FactoryMap[manipulation.Remapper]{
	RemapperKazaam: FactoryFunc[manipulation.Remapper](CreateRemapperKazaam),
	RemapperNoOp:   FactoryFunc[manipulation.Remapper](CreateRemapperNoOp),
}

func CreateRemapperKazaam(name string, config Config) (manipulation.Remapper, error) {
	if name != RemapperKazaam {
		return nil, fmt.Errorf(
			"KazaamRemapperFactory: called with unexpected name '%s', want '%s'",
			name,
			RemapperKazaam,
		)
	}
	config, _ = extractConfigIfSet(name, config)
	specString := ""
	if payload, ok := config["spec"]; ok {
		switch spec := payload.(type) {
		case string:
			specString = spec
		case map[string]any:
			specBytes, _ := json.Marshal(spec)
			specString = string(specBytes)
		}
	}

	const entryName = RemapperReferenceName + "." + RemapperKazaam

	kazaamInst, err := kazaam.NewKazaam(specString)
	if err != nil {
		return nil, errors.Wrapf(err, "%s: cannot create kazaam instance", entryName)
	}

	return manipulation.NewKazaamRemapper(kazaamInst), nil
}

func CreateRemapperNoOp(name string, _ Config) (manipulation.Remapper, error) {
	if name != RemapperNoOp {
		return nil, fmt.Errorf(
			"CreateRemapperNoOp: called with unexpected name '%s', want '%s'",
			name,
			RemapperNoOp,
		)
	}
	return manipulation.NewNoOpRemapper(), nil
}
