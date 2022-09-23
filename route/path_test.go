package route

import (
	"fmt"
	"reflect"
	"testing"
)

type retrieveNamedSegmentTest struct {
	in       string
	outPath  string
	segments map[string]int
}

var retrieveNamedSegmentTests = []retrieveNamedSegmentTest{
	{
		in:      "/resource/:id/*",
		outPath: "/resource/*/*",
		segments: map[string]int{
			"id": 2,
		},
	},
	{
		in:      "resource/v1/:p1/*/::p2/",
		outPath: "resource/v1/*/*/*/",
		segments: map[string]int{
			"p1":  2,
			":p2": 4,
		},
	},
}

func TestRetrieveNamedSegments(t *testing.T) {
	for i, tt := range retrieveNamedSegmentTests {
		outPath, segments, err := retrieveNamedSegments(tt.in)
		meta := fmt.Sprintf("test #%d: retrieveNamedSegments(%q),", i, tt.in)
		if err != nil {
			t.Errorf("%s unexpected error: %s", meta, err)
		}

		if outPath != tt.outPath {
			t.Errorf("%s expected to return path %q, got %q", meta, tt.outPath, outPath)
		}
		if !reflect.DeepEqual(segments, tt.segments) {
			t.Errorf("%s expected to return segments %+v, got %+v", meta, tt.segments, segments)
		}
	}
}

type retrieveNamedParametersTest struct {
	in       string
	segments map[string]int
	out      NamedPathParameters
}

var retrieveNamedParametersTests = []retrieveNamedParametersTest{
	{
		in:       "/books/123",
		segments: map[string]int{"id": 2},
		out:      NamedPathParameters{"id": "123"},
	},
	{
		in:       "param1/param2/param3",
		segments: map[string]int{"p1": 0, "p2": 1, "p3": 2},
		out:      NamedPathParameters{"p1": "param1", "p2": "param2", "p3": "param3"},
	},
}

func TestRetrieveNamedParameters(t *testing.T) {
	for i, tt := range retrieveNamedParametersTests {
		parameters := retrieveNamedParameters(tt.segments, tt.in)
		meta := fmt.Sprintf("test #%d: retrieveNamedParameters(%+v, %q),", i, tt.segments, tt.in)
		if !reflect.DeepEqual(parameters, tt.out) {
			t.Errorf("%s expected to return parameters %+v, got %+v", meta, tt.out, parameters)
		}
	}
}

type wildcardMatcherTest struct {
	template           string
	against            string
	expectedMatch      bool
	expectedParameters NamedPathParameters
}

var wildcardMatcherTests = []wildcardMatcherTest{
	{
		template:           "/api/*/resource/:id/**",
		against:            "/api/v1/resource/resource-id/param/param/",
		expectedMatch:      true,
		expectedParameters: NamedPathParameters{"id": "resource-id"},
	},
	{
		template:           "**",
		against:            "anything matches this pattern",
		expectedMatch:      true,
		expectedParameters: NamedPathParameters{},
	},
	{
		template:           "/api/v1|v2/:param",
		against:            "/api/v1/123",
		expectedMatch:      true,
		expectedParameters: NamedPathParameters{"param": "123"},
	},
	{
		template:           "api/v1|v2/:param",
		against:            "api/v2/123",
		expectedMatch:      true,
		expectedParameters: NamedPathParameters{"param": "123"},
	},
	{
		template:           "/api/v1|v2/:param",
		against:            "/api/v3/123",
		expectedMatch:      false,
		expectedParameters: NamedPathParameters{},
	},
}

func TestWildCardMatcher(t *testing.T) {
	for i, tt := range wildcardMatcherTests {
		matcher, err := NewWildCardMatcher(tt.template)
		meta := fmt.Sprintf("test #%d: matcher.Match(%q),", i, tt.against)
		if err != nil {
			t.Errorf("%s unexpected error: %s", meta, err)
		}

		match := matcher.Match(tt.against)
		if match != tt.expectedMatch {
			t.Errorf("%s expected match to be %t, got %t", meta, tt.expectedMatch, match)
		}

		meta = fmt.Sprintf("test #%d: matcher.Parameters(%q),", i, tt.against)
		parameters := matcher.RetrieveParameters(tt.against)
		if !reflect.DeepEqual(parameters, tt.expectedParameters) {
			t.Errorf("%s expected to return parameters %+v, got %+v", meta, tt.expectedParameters, parameters)
		}
	}
}
