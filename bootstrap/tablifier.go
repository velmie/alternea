package bootstrap

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/velmie/alternea/manipulation"
)

const (
	TablifierReferenceName = "tablifier"
	TablifierJSON          = "json"
)

var Tablifier = FactoryMap[manipulation.Tablifier]{
	TablifierJSON: JSONTablifierFactory(Remapper, manipulation.NewNoOpRemapper()),
}

func JSONTablifierFactory(
	remapperFactory Factory[manipulation.Remapper],
	defaultRemapper manipulation.Remapper,
) Factory[manipulation.Tablifier] {
	return FactoryFunc[manipulation.Tablifier](func(name string, config Config) (manipulation.Tablifier, error) {
		if name != TablifierJSON {
			return nil, fmt.Errorf(
				"CreateJSONTablifier: called with unexpected name '%s', want '%s'",
				name,
				TablifierJSON,
			)
		}
		config, _ = extractConfigIfSet(name, config)
		var err error
		remapperConfig, exist := extractConfigIfSet(RemapperReferenceName, config)
		remapper := defaultRemapper
		const entryName = TablifierReferenceName + "." + TablifierJSON

		if exist {
			remapper, err = remapperFactory.Create(remapperConfig.GetString("name"), remapperConfig)
			if err != nil {
				return nil, errors.Wrapf(err, "%s: cannot create remapper", entryName)
			}
		}
		tablifierConfig := &manipulation.JSONTablifierConfig{}
		if err = decode(config, tablifierConfig); err != nil {
			return nil, errors.Wrapf(err, "%s: cannot decode configuration", entryName)
		}
		return manipulation.NewJSONTablifier(remapper, GetLogger(), tablifierConfig), nil
	})
}
