package route

import (
	"fmt"
	"strings"
)

type NamedPathParameters map[string]string

// PathMatcher must check if the given path matches template path.
type PathMatcher interface {
	PathParametersRetriever
	Match(path string) bool
}

type PathParametersRetriever interface {
	RetrieveParameters(path string) NamedPathParameters
}

type WildCardPathMatcher struct {
	template      []byte
	s             string
	namedSegments map[string]int
}

func NewWildCardMatcher(template string) (PathMatcher, error) {
	template, segments, err := retrieveNamedSegments(template)
	if err != nil {
		return nil, err
	}
	tplBytes := []byte(template)
	if len(tplBytes) > 0 && tplBytes[len(tplBytes)-1] != '/' {
		tplBytes = append(tplBytes, '/')
	}
	return &WildCardPathMatcher{
		template:      tplBytes,
		s:             template,
		namedSegments: segments,
	}, nil
}

// Match checks if the given path matches template path.
// The template path may include wildcards
//
//	"*"  - to match one segment
//	"**" - to match everything
//
// for example template "/public/**" will return true for any path prefixed with "/public"
// template "/public/*/id will return true for those paths that starts from "/public", has any value
// in segment 2 and "id" in segment 3
// The template path may also include "or" logical expressions such as "/this|or-this"
// which could be combined with the wildcards e.g. "/v1/service-*|resource/**"
func (m *WildCardPathMatcher) Match(path string) bool {
	return m.match(path)
}

func (m *WildCardPathMatcher) RetrieveParameters(path string) NamedPathParameters {
	if !m.match(path) {
		return NamedPathParameters{}
	}
	return retrieveNamedParameters(m.namedSegments, path)
}

func (m *WildCardPathMatcher) match(path string) bool {
	if path == m.s {
		return true
	}

	pathBytes := []byte(path)
	if pathBytes[len(pathBytes)-1] != '/' {
		pathBytes = append(pathBytes, '/')
	}
	tpl := m.template
	matchOne := false
	delimiterPrev := false
	j := 0
	pathSegStartPos := -1

	for i := 0; i < len(pathBytes) || matchOne; i++ {
		for delimiterPrev {
			if len(pathBytes) < i+1 {
				break
			}
			delimiterPrev = pathBytes[i] == '/'
			if delimiterPrev {
				i++
				continue
			}
			pathSegStartPos = i
		}

		if pathSegStartPos == -1 {
			if pathBytes[i] == ' ' {
				continue
			}
			if tpl[j] == ' ' {
				j++
				i--
				continue
			}
		}

		for matchOne {
			if len(pathBytes) < i+1 {
				if len(tpl) == j+1 && tpl[j] == '/' {
					return true
				}
				if len(tpl) > j+1 && tpl[j] == '|' {
					for tpl[j] != '/' {
						j++
					}
				}
				return len(tpl) <= j+1
			}
			if pathBytes[i] == '/' {
				matchOne = false
				break
			}
			i++
		}

		if len(tpl) > j+1 && tpl[j] == '|' {
			for len(tpl) > j-1 && tpl[j] != '/' {
				j++
			}
			i--
			continue
		}

		if len(tpl) < j+1 {
			return len(pathBytes) == i
		}

		if tpl[j] == '*' {
			matchOne = true
			if len(tpl) >= j+2 && tpl[j+1] == '*' {
				return true
			}
			j++
			continue
		}

		if pathBytes[i] != tpl[j] {
			if delimiterPrev && pathBytes[i] == '/' {
				continue
			}

			checkOption := false
			for len(tpl) > j-1 && tpl[j] != '/' {
				j++
				if tpl[j] == '|' {
					checkOption = true
					j++
					if pathSegStartPos > -1 {
						i = pathSegStartPos - 1
						break
					}
					i--
					break
				}
			}
			if checkOption {
				continue
			}
			return false
		}
		delimiterPrev = pathBytes[i] == '/'
		j++
	}
	return len(tpl) <= j
}

func retrieveNamedParameters(segments map[string]int, path string) NamedPathParameters {
	parameters := make(NamedPathParameters)
	if len(segments) == 0 {
		return parameters
	}
	parts := strings.Split(path, "/")
	for name, index := range segments {
		if len(parts) >= index {
			parameters[name] = parts[index]
		}
	}
	return parameters
}

func retrieveNamedSegments(template string) (string, map[string]int, error) {
	segments := make(map[string]int)
	parts := strings.Split(template, "/")
	newParts := make([]string, len(parts))
	twoStars := false
	for i, part := range parts {
		if part == "**" {
			twoStars = true
		}
		if len(part) > 1 && part[0] == ':' {
			if twoStars {
				return "", nil, fmt.Errorf("named parameters are not allowed after the '**' entry  %s", template)
			}
			segments[part[1:]] = i
			newParts[i] = "*"
			continue
		}
		newParts[i] = part
	}
	return strings.Join(newParts, "/"), segments, nil
}
