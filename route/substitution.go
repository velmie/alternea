package route

import (
	"net/http"
	"strings"
)

type PathTemplate interface {
	Render(params NamedPathParameters) string
}

type ColonParamsReplaceTemplate string

func (c ColonParamsReplaceTemplate) Render(params NamedPathParameters) string {
	result := string(c)
	for name, value := range params {
		result = strings.ReplaceAll(result, ":"+name, value)
	}
	return result
}

func PathSubstitution(
	template PathTemplate,
	next http.Handler,
) Handler {
	return func(w http.ResponseWriter, r *http.Request, p NamedPathParameters) {
		r.URL.Path = template.Render(p)
		next.ServeHTTP(w, r)
	}
}
