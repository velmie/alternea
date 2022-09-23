package bootstrap

import (
	"errors"
	"fmt"
	"math/big"
	"sort"

	"github.com/zclconf/go-cty/cty"
)

// extractGoValues converts cty.Value to Go value
func extractGoValues(v cty.Value, t cty.Type) (interface{}, error) {
	if v.IsNull() {
		return nil, nil
	}
	if !v.IsKnown() {
		return nil, errors.New("value is unknown")
	}

	if t.IsPrimitiveType() {
		switch t {
		case cty.String:
			return v.AsString(), nil
		case cty.Number:
			if v.RawEquals(cty.PositiveInfinity) {
				return (&big.Float{}).SetInf(false), nil
			}
			if v.RawEquals(cty.NegativeInfinity) {
				return (&big.Float{}).SetInf(true), nil
			}
			result, _ := v.AsBigFloat().Float64()
			return result, nil
		case cty.Bool:
			return v.True(), nil
		default:
			// should never happen
			panic("unsupported primitive type")
		}
	}

	switch {
	case t.IsListType(), t.IsSetType():
		ety := t.ElementType()
		it := v.ElementIterator()
		result := make([]interface{}, 0, v.LengthInt())
		for it.Next() {
			_, ev := it.Element()
			res, err := extractGoValues(ev, ety)
			if err != nil {
				return nil, err
			}
			result = append(result, res)
		}
		return result, nil

	case t.IsTupleType():
		etys := t.TupleElementTypes()
		it := v.ElementIterator()

		i := 0
		result := make([]interface{}, v.LengthInt())
		for it.Next() {
			ety := etys[i]
			_, ev := it.Element()
			res, err := extractGoValues(ev, ety)
			if err != nil {
				return nil, err
			}
			result[i] = res
			i++
		}
		return result, nil

	case t.IsMapType():
		ety := t.ElementType()
		it := v.ElementIterator()
		result := map[interface{}]interface{}{}
		for it.Next() {
			ek, ev := it.Element()
			kres, err := extractGoValues(ek, ek.Type())
			if err != nil {
				return nil, err
			}
			vres, err := extractGoValues(ev, ety)
			if err != nil {
				return nil, err
			}
			result[kres] = vres
		}
		return result, nil

	case t.IsObjectType():
		atys := t.AttributeTypes()
		names := make([]string, 0, len(atys))
		for k := range atys {
			names = append(names, k)
		}
		sort.Strings(names)

		result := map[string]interface{}{}
		for _, name := range names {
			aty := atys[name]
			av := v.GetAttr(name)
			vres, err := extractGoValues(av, aty)
			if err != nil {
				return nil, err
			}
			result[name] = vres
		}
		return result, nil

	case t.IsCapsuleType():
		return v.EncapsulatedValue(), nil
	}
	// should never happen also
	return nil, fmt.Errorf("cannot extract value %s", t.FriendlyName())
}
