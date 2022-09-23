package dframe

import (
	"fmt"
)

type StringFormatter func(v any) string

type Column struct {
	name           string
	nullable       bool
	formatter      StringFormatter
	nilPlaceholder string
	values         []any
}

func NewColumn(name string, options ...ColumnOption) *Column {
	c := &Column{name: name}
	for _, option := range options {
		option(c)
	}
	return c
}

func (c *Column) Add(vals ...any) {
	for _, v := range vals {
		if c.nullable && v == nil {
			c.values = append(c.values, v)
			continue
		}
		c.values = append(c.values, v)
	}
}

func (c *Column) StringVal(index int) string {
	v := c.values[index]
	if v == nil {
		return c.nilPlaceholder
	}
	formatter := c.formatter
	if formatter == nil {
		formatter = defaultFormatter
	}

	return formatter(v)
}

func (c *Column) StringSlice() []string {
	dest := make([]string, len(c.values))
	for i := 0; i < len(c.values); i++ {
		dest[i] = c.StringVal(i)
	}
	return dest
}

func defaultFormatter(v any) string {
	return fmt.Sprintf("%v", v)
}
