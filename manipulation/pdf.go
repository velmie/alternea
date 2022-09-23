package manipulation

import (
	"context"
	"html/template"
	"io"

	pdf "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"golang.org/x/sync/errgroup"
)

type PDFTransformer struct {
	remapper     Remapper
	pageTemplate *template.Template
}

func NewPDFTransformer(remapper Remapper, pageTemplate *template.Template) *PDFTransformer {
	return &PDFTransformer{remapper: remapper, pageTemplate: pageTemplate}
}

func (t *PDFTransformer) Transform(ctx context.Context, pages <-chan []byte, w io.Writer) error {
	pipeReader, pipeWriter := io.Pipe()

	pdfgen, err := pdf.NewPDFGenerator()

	if err != nil {
		_ = pipeWriter.Close()
		return errors.Wrap(err, "PDFTransformer: cannot create PDF Generator instance")
	}
	pdfgen.SetOutput(w)
	pdfgen.AddPage(pdf.NewPageReader(pipeReader))

	errs, ctx := errgroup.WithContext(ctx)
	errs.Go(func() error {
		return pdfgen.CreateContext(ctx)
	})

	var data []byte
	type templateData struct {
		Data any
	}
LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case page, more := <-pages:
			if !more {
				break LOOP
			}
			data, err = t.remapper.Remap(page)
			if err != nil {
				if err != nil {
					return errors.Wrap(err, "PDFTransformer: cannot remap given input")
				}
			}
			err = t.pageTemplate.Execute(pipeWriter, &templateData{
				Data: gjson.ParseBytes(data).Value(),
			})
			if err != nil {
				return errors.Wrap(err, "PDFTransformer: cannot execute page template")
			}
		}
	}

	pipeWriter.Close()

	return errs.Wait()
}
