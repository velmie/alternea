package dframe

type ColumnOption func(c *Column)

func Nullable(c *Column) {
	c.nullable = true
}

func WithFormatter(formatter StringFormatter) ColumnOption {
	return func(c *Column) {
		c.formatter = formatter
	}
}

func WithNilPlaceholder(placeholder string) ColumnOption {
	return func(c *Column) {
		c.nilPlaceholder = placeholder
	}
}
