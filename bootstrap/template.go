package bootstrap

import (
	"fmt"
	"html/template"
	"time"

	"github.com/pkg/errors"
)

func registerCustomFunctions(t *template.Template) *template.Template {
	return t.Funcs(templateFunctions())
}

func templateFunctions() template.FuncMap {
	return template.FuncMap{
		"fdate": formatDate,
	}
}

// formatDate parses the value into `time.Time` and returns a formatted string.
// Reason of creating that function: in an HTML template all values have interface{} type. So we cannot use usual
// way of date formatting like `{{ .CreatedOn.Format "2006 Jan 02" }}`.
func formatDate(inputFormat, outputFormat string, val any) (string, error) {
	var stringTime string
	switch v := val.(type) {
	case any:
		s, ok := val.(string)
		if !ok {
			return "", fmt.Errorf("cannot convert the value %v to string", val)
		}
		stringTime = s

	case string:
		stringTime = v
	}
	if stringTime == "" {
		return "", nil
	}
	t, err := time.Parse(inputFormat, stringTime)
	if err != nil {
		return "", errors.Wrapf(err, "cannot parse %s to time.Time", stringTime)
	}

	return t.Format(outputFormat), nil
}
