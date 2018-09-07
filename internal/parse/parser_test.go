package parse

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCommandSetsTheCorrectMethod(t *testing.T) {
	r, _ := ParseCommand(http.MethodPut, "/")
	require.EqualValues(t, http.MethodPut, r.Method)
}

func TestParseCommandDefaultsMethodToGet(t *testing.T) {
	r, _ := ParseCommand("httpbin.org")
	require.EqualValues(t, http.MethodGet, r.Method)
}

func TestParseCommandSetsHost(t *testing.T) {
	raw := "httpbin.org"
	r, err := ParseCommand("GET", raw)
	require.NoError(t, err)
	require.EqualValues(t, "http://httpbin.org", r.URL.String())
}

func TestParseCommandSetsHeaders(t *testing.T) {
	r, _ := ParseCommand("GET", "https://httpbin.org/post", "Auth:token", "Accept:application/json")
	require.EqualValues(t, http.Header{
		"Auth":   []string{"token"},
		"Accept": []string{"application/json"},
	}, r.Header)
}

func TestParseCommandSetsMultiple(t *testing.T) {
	r, _ := ParseCommand("GET", "https://httpbin.org/post",
		"Auth:token", "Accept:application/json",
		"q==search", "page==1",
	)
	require.EqualValues(t, http.Header{
		"Auth":   []string{"token"},
		"Accept": []string{"application/json"},
	}, r.Header)

	require.EqualValues(t, url.Values{
		"q":    []string{"search"},
		"page": []string{"1"},
	}, r.URL.Query())
}

func TestParseCommandSetsJSONBody(t *testing.T) {
	r, _ := ParseCommand("GET", "https://httpbin.org/post",
		"name=brett",
		"age:=100",
	)
	body, err := ioutil.ReadAll(r.Body)
	require.NoError(t, err, "failed to read json body")

	expected := `{
		"name": "brett",
		"age": 100
	}`

	require.JSONEq(t, expected, string(body))
}

func TestParseCommandSetsNestedJSONBody(t *testing.T) {
	r, _ := ParseCommand("GET", "https://httpbin.org/post",
		"person.name=brett",
		"person.age:=100",
		"type=person",
	)
	body, err := ioutil.ReadAll(r.Body)
	require.NoError(t, err, "failed to read json body")

	expected := `{
		"type": "person",
		"person": {
			"name": "brett",
			"age": 100
		}
	}`

	require.JSONEq(t, expected, string(body))
}

func TestParseCommandParsesQueryStrings(t *testing.T) {
	type result struct {
		k string
		v []string
	}

	tests := map[string]result{
		"a==b": {
			k: "a",
			v: []string{"b"},
		},
		"hell?o==world": {
			k: "hell?o",
			v: []string{"world"},
		},
	}

	for in, out := range tests {
		t.Run(in, func(t *testing.T) {
			r, _ := ParseCommand("/", in)
			v := r.URL.Query()[out.k]
			assert.EqualValues(t, out.v, v)
		})
	}

	t.Run("multiple values", func(t *testing.T) {
		r, _ := ParseCommand("/", "hello==1", "hello==2")
		v := r.URL.Query()["hello"]
		assert.EqualValues(t, []string{"1", "2"}, v)
	})
}

func TestParseKV(t *testing.T) {
	type testCase struct {
		raw string
		ok  bool
		kv  KV
	}

	tests := map[string]testCase{
		"header": {
			raw: "a:b",
			kv: KV{
				k:    "a",
				v:    "b",
				kind: headerArg,
			},
		},
		"query": {
			raw: "a==b",
			kv: KV{
				k:    "a",
				v:    "b",
				kind: queryArg,
			},
		},
		"field": {
			raw: "a=b",
			kv: KV{
				k:    "a",
				v:    "b",
				kind: fieldArg,
			},
		},
		"field_number": {
			raw: "a=1",
			kv: KV{
				k:    "a",
				v:    "1",
				kind: fieldArg,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			kv, ok := parseKV(tc.raw)
			if assert.True(t, ok) {
				assert.EqualValues(t, tc.kv, kv)
			}
		})
	}
}

func TestParseHost(t *testing.T) {
	type tc struct {
		name, test string
		result     *url.URL
	}

	tests := []tc{
		{
			name:   "colon",
			test:   ":",
			result: mustParseURL(t, "http://localhost/"),
		},
		{
			name:   "localhost",
			test:   "localhost",
			result: mustParseURL(t, "http://localhost/"),
		},
		{
			name:   "port_only",
			test:   ":8081",
			result: mustParseURL(t, "http://localhost:8081"),
		},
		{
			name:   "port_and_path",
			test:   ":8080/hello",
			result: mustParseURL(t, "http://localhost:8080/hello"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := parseHost(test.test)
			if assert.NoError(t, err) {
				assert.Equal(t, test.result, res)
			}
		})
	}
}

func TestParseKVJSONArray(t *testing.T) {
	type testCase struct {
		name, raw string
		res       interface{}
	}
	type obj map[string]interface{}

	// Numbers are parsed as float64
	tests := []testCase{
		{
			name: "bool",
			raw:  `items:=true`,
			res:  true},
		{
			name: "number",
			raw:  `items:=1`,
			res:  1.0},
		{
			name: "array",
			raw:  `items:=[1,2,3]`,
			res:  []interface{}{1.0, 2.0, 3.0}},
		{
			name: "object",
			raw:  `items:={ "name": "brett" }`,
			res:  obj{"name": "brett"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, _ := parseKV(test.raw)
			assert.EqualValues(t, test.res, res.v)
		})
	}
}

func mustParseURL(t *testing.T, s string) *url.URL {
	u, err := url.Parse(s)
	require.NoError(t, err)
	return u
}
