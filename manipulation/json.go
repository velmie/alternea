package manipulation

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/qntfy/kazaam"
	"github.com/tidwall/gjson"

	"github.com/velmie/alternea/app"
	"github.com/velmie/alternea/dframe"
)

type JSONTablifierConfig struct {
	// Columns determines which columns should be added to the result
	// columns will be selected in the order in which they are specified
	Columns []string
}

// JSONTablifier create table from input json bytes using the Remapper
// which must end up with json object where all properties are arrays
// that represent columns
// e.g. {  "Name":["Alan","Alex","Boris"], "Age": [42, 49, 15]  }
type JSONTablifier struct {
	columnsRemapper Remapper
	logger          app.Logger
	config          *JSONTablifierConfig
}

func NewJSONTablifier(
	columnsRemapper Remapper,
	logger app.Logger,
	config *JSONTablifierConfig,
) *JSONTablifier {
	return &JSONTablifier{columnsRemapper, logger, config}
}

func (t *JSONTablifier) Table(in []byte) (*dframe.Table, error) {
	if t.logger.Level() >= app.DebugLevel {
		t.logger.Debugf("JSONTablifier: input data:\n%s\n", in)
	}
	out, err := t.columnsRemapper.Remap(in)
	if err != nil {
		return nil, errors.Wrap(err, "JSONTablifier: cannot remap given input")
	}
	if t.logger.Level() >= app.DebugLevel {
		dbg, _ := json.MarshalIndent(json.RawMessage(out), "", "  ")
		t.logger.Debugf("JSONTablifier: remapped data:\n%s\n", dbg)
	}

	result := gjson.ParseBytes(out)
	if !result.IsObject() {
		typeName := result.Type.String()
		if result.IsArray() {
			typeName = "array"
		}
		return nil, errors.Wrapf(
			ErrUnsupportedDataType,
			"JSONTablifier: remapping must result in an object, got '%s'",
			typeName,
		)
	}

	table, _ := dframe.NewTable()
	config := t.config
	addColumn := func(key string, value gjson.Result) error {
		if !value.IsArray() {
			typeName := value.Type.String()
			if value.IsObject() {
				typeName = "object"
			}
			return errors.Wrapf(
				ErrUnsupportedDataType,
				"JSONTablifier: remapped object properties must be arrays, but '%s' is '%s'",
				key,
				typeName,
			)
		}
		column := dframe.NewColumn(key, dframe.Nullable)
		column.Add(value.Value().([]any)...)
		if err = table.Append(column); err != nil {
			return errors.Wrap(err, "JSONTablifier: cannot append column")
		}
		return nil
	}

	if config != nil && len(config.Columns) > 0 {
		for _, key := range config.Columns {
			if err = addColumn(key, result.Get(key)); err != nil {
				break
			}
		}
	} else {
		result.ForEach(func(key, value gjson.Result) bool {
			if err = addColumn(key.String(), value); err != nil {
				return false
			}
			return true
		})
	}

	if err != nil {
		return nil, err
	}

	return table, nil
}

type KazaamRemapper struct {
	k *kazaam.Kazaam
}

func NewKazaamRemapper(k *kazaam.Kazaam) *KazaamRemapper {
	return &KazaamRemapper{k: k}
}

func (r *KazaamRemapper) Remap(in []byte) ([]byte, error) {
	return r.k.Transform(in)
}

type NoOpRemapper struct {
}

func NewNoOpRemapper() *NoOpRemapper {
	return &NoOpRemapper{}
}

func (n NoOpRemapper) Remap(in []byte) ([]byte, error) {
	return in, nil
}
