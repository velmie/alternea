package bootstrap

import (
	"encoding/json"
	"os"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

var evalContext = &hcl.EvalContext{
	Functions: map[string]function.Function{
		"duration":      duration,
		"env":           env,
		"unmarshalJSON": unmarshalJSON,
		"fromFile":      fromFile,
		"format":        stdlib.FormatFunc,
		"printf":        stdlib.FormatFunc,
	},
}

var (
	duration = function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "val",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.Number),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			d, err := time.ParseDuration(args[0].AsString())
			if err != nil {
				return cty.NumberIntVal(0), err
			}
			return cty.NumberIntVal(int64(d)), nil
		},
	})
	env = function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "val",
				Type: cty.String,
			},
		},
		VarParam: &function.Parameter{
			Name: "fallback",
			Type: cty.String,
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			v, ok := os.LookupEnv(args[0].AsString())
			if !ok && len(args) > 1 {
				return args[1], nil
			}
			return cty.StringVal(v), nil
		},
	})
	unmarshalJSON = function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "val",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			arg := args[0].AsString()
			if arg == "" {
				return cty.NilVal, nil
			}
			v := new(ctyjson.SimpleJSONValue)
			if err := json.Unmarshal([]byte(arg), v); err != nil {
				return cty.Value{}, err
			}
			return v.Value, nil
		},
	})
	fromFile = function.New(&function.Spec{
		Params: []function.Parameter{
			{
				Name: "path",
				Type: cty.String,
			},
		},
		Type: function.StaticReturnType(cty.String),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			arg := args[0].AsString()
			if arg == "" {
				return cty.NilVal, nil
			}
			content, err := os.ReadFile(arg)
			if err != nil {
				return cty.NilVal, err
			}
			return cty.StringVal(string(content)), nil
		},
	})
)
