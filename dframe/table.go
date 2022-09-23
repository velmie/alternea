package dframe

import (
	"fmt"
)

type Table struct {
	columns   []*Column
	nameIndex map[string]int
	numRows   int
}

func NewTable(columns ...*Column) (*Table, error) {
	table := &Table{
		nameIndex: make(map[string]int),
	}
	for _, c := range columns {
		if err := table.Append(c); err != nil {
			return nil, err
		}
	}
	return table, nil
}

func (t *Table) Append(c *Column) error {
	if _, ok := t.nameIndex[c.name]; ok {
		return fmt.Errorf(
			"column names must be unique within a single table: column %s is already exist",
			c.name,
		)
	}
	if len(t.columns) == 0 {
		t.numRows = len(c.values)
	}
	if len(c.values) < t.numRows {
		if err := t.expand([]*Column{c}, t.numRows); err != nil {
			return err
		}
	} else if t.numRows < len(c.values) {
		if err := t.expand(t.columns, len(c.values)); err != nil {
			return err
		}
		t.numRows = len(c.values)
	}
	t.columns = append(t.columns, c)
	t.nameIndex[c.name] = len(t.columns) - 1
	return nil
}

func (t *Table) NumCols() int {
	return len(t.columns)
}

func (t *Table) NumRows() int {
	return t.numRows
}

func (t *Table) StringSlices() [][]string {
	if t.numRows == 0 {
		return nil
	}
	numColumns := len(t.columns)
	dest := make([][]string, t.numRows)
	for i := 0; i < t.numRows; i++ {
		row := make([]string, numColumns)
		for j, col := range t.columns {
			row[j] = col.StringVal(i)
		}
		dest[i] = row
	}
	return dest
}

func (t *Table) Header() []string {
	names := make([]string, len(t.columns))
	for i, c := range t.columns {
		names[i] = c.name
	}
	return names
}

func (t *Table) Select(names ...string) error {
	columns := make([]*Column, 0, len(names))
	for _, name := range names {
		i, ok := t.nameIndex[name]
		if !ok {
			return fmt.Errorf("unknown column %s", name)
		}
		columns = append(columns, t.columns[i])
	}
	t.columns = columns
	t.compact()

	return nil
}

func (t *Table) Limit(limit uint) *Table {
	if int(limit) > t.numRows {
		return t
	}
	for _, col := range t.columns {
		col.values = col.values[:limit]
	}
	t.numRows = int(limit)
	return t
}

func (t *Table) compact() {
	newNumRows := t.numRows
	for i := t.numRows - 1; i > 0; i-- {
		setNewLen := true
		for _, col := range t.columns {
			if col.values[i] != nil {
				setNewLen = false
				break
			}
		}
		if setNewLen {
			newNumRows = i
		}
	}
	if t.numRows != newNumRows {
		t.Limit(uint(newNumRows))
	}
}

func (t *Table) expand(columns []*Column, newLen int) error {
	for _, c := range columns {
		if len(c.values) < newLen {
			if !c.nullable {
				return fmt.Errorf("cannot expand column %s becase the column is not nullable", c.name)
			}
			if cap(c.values) > newLen {
				c.values = c.values[:newLen]
			} else {
				values := make([]any, newLen)
				copy(values, c.values)
				c.values = values
			}
		}
	}
	return nil
}
