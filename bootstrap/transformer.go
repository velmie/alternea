package bootstrap

import (
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/velmie/alternea/manipulation"

	"github.com/pkg/errors"
)

const (
	TransformerReferenceName = "transformer"
	TransformerCSV           = "csv"
	TransformerPDF           = "pdf"
)

var Transformer = FactoryMap[manipulation.DataTransformer]{
	TransformerCSV: CSVTransformer(Tablifier),
	TransformerPDF: PDFTransformer(Remapper),
}

func CSVTransformer(tablifierFactory Factory[manipulation.Tablifier]) Factory[manipulation.DataTransformer] {
	return FactoryFunc[manipulation.DataTransformer](
		func(name string, config Config) (manipulation.DataTransformer, error) {
			if name != TransformerCSV {
				return nil, fmt.Errorf(
					"CSVTransformer: called with unexpected name '%s', want '%s'",
					name,
					TransformerCSV,
				)
			}

			tablifierConfig, exist := extractConfigIfSet(TablifierReferenceName, config)
			const entryName = TransformerReferenceName + "." + TransformerCSV
			if !exist {
				return nil, errRequiredConfiguration(entryName, TablifierReferenceName)
			}
			tablifier, err := tablifierFactory.Create(tablifierConfig.GetString("name"), tablifierConfig)
			if err != nil {
				return nil, errors.Wrapf(err, "%s: cannot create %s", entryName, TablifierReferenceName)
			}

			csvConfig := manipulation.CSVTransformerConfig{}
			if err = decode(config, &csvConfig); err != nil {
				return nil, errors.Wrapf(err, "%s: cannot decode configuration", entryName)
			}

			return manipulation.NewCSVTransformer(tablifier, csvConfig), nil
		})
}

func PDFTransformer(remapperFactory Factory[manipulation.Remapper]) Factory[manipulation.DataTransformer] {
	return FactoryFunc[manipulation.DataTransformer](
		func(name string, config Config) (manipulation.DataTransformer, error) {
			if name != TransformerPDF {
				return nil, fmt.Errorf(
					"PDFTransformer: called with unexpected name '%s', want '%s'",
					name,
					TransformerPDF,
				)
			}

			const entryName = "transformer." + TransformerPDF

			if err := wkhtmltopdfFindPath(); err != nil {
				return nil, fmt.Errorf("%s requires wkhtmltopdf https://wkhtmltopdf.org/", entryName)
			}

			templateAttr := config.GetString("template")
			if templateAttr == "" {
				return nil, errRequiredConfiguration(entryName, "template")
			}
			htmlTemplate, err := template.New(TransformerPDF).Parse(templateAttr)
			if err != nil {
				return nil, errors.Wrapf(err, "%s: cannot parse template", entryName)
			}

			var remapper manipulation.Remapper
			remapperConfig, exist := extractConfigIfSet(RemapperReferenceName, config)
			if !exist {
				remapper = manipulation.NewNoOpRemapper()
			} else {
				remapper, err = remapperFactory.Create(remapperConfig.GetString("name"), remapperConfig)
				if err != nil {
					return nil, errors.Wrapf(err, "%s: cannot create remapper", entryName)
				}
			}

			pdfTransformer := manipulation.NewPDFTransformer(remapper, htmlTemplate)

			return pdfTransformer, nil
		})
}

func wkhtmltopdfFindPath() error {
	const exe = "wkhtmltopdf"

	exeDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}
	path, err := exec.LookPath(filepath.Join(exeDir, exe))
	if err == nil && path != "" {
		return nil
	}
	path, err = exec.LookPath(exe)
	if err == nil && path != "" {
		return nil
	}
	dir := os.Getenv("WKHTMLTOPDF_PATH")
	if dir == "" {
		return fmt.Errorf("%s not found", exe)
	}
	path, err = exec.LookPath(filepath.Join(dir, exe))
	if err == nil && path != "" {
		return nil
	}
	return fmt.Errorf("%s not found", exe)
}
