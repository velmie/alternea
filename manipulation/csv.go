package manipulation

import (
	"context"
	"encoding/csv"
	"io"
	"unicode/utf8"

	"github.com/pkg/errors"
)

type CSVTransformerConfig struct {
	UseHeader bool   // True to use header as a first line
	Delimiter string // Field delimiter (set to ',' by default)
	UseCRLF   bool   // True to use \r\n as the line terminator
}

type CSVTransformer struct {
	config    CSVTransformerConfig
	tablifier Tablifier
}

func NewCSVTransformer(tablifier Tablifier, cfg CSVTransformerConfig) *CSVTransformer {
	return &CSVTransformer{cfg, tablifier}
}

func (t *CSVTransformer) Transform(ctx context.Context, pages <-chan []byte, w io.Writer) error {
	csvWriter := csv.NewWriter(w)
	if t.config.Delimiter != "" {
		r, _ := utf8.DecodeRuneInString(t.config.Delimiter)
		csvWriter.Comma = r
	}
	if t.config.UseCRLF {
		csvWriter.UseCRLF = true
	}
	headerSet := false
	for page := range pages {
		table, err := t.tablifier.Table(page)
		if err != nil {
			return errors.Wrap(err, "CSVTransformer: cannot get table")
		}
		if t.config.UseHeader && !headerSet {
			headerSet = true
			if err = csvWriter.Write(table.Header()); err != nil {
				return errors.Wrap(err, "CSVTransformer: cannot write header")
			}
		}
		if err = csvWriter.WriteAll(table.StringSlices()); err != nil {
			return errors.Wrap(err, "CSVTransformer: cannot write table")
		}
	}
	return nil
}
