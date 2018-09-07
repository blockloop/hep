package parse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Jeffail/gabs"
	"github.com/blockloop/hep/internal/log"
)

var re = regexp.MustCompile(`([^:=@]+)((?:==)|(?::=)|(?:=)|(?::))(.+)`)

type argType int

const (
	headerArg argType = iota
	fieldArg
	queryArg
	jsonArg
)

// KV is a result of parseKV
type KV struct {
	k    string
	v    interface{}
	kind argType
}

// ParseCommand parses command line arguments into an HTTP request
func ParseCommand(args ...string) (r *http.Request, err error) {
	unknown := make(chan string, len(args))
	defer close(unknown)

	var (
		method, path string
		uri          *url.URL
		q            = url.Values{}
		fields       = gabs.New()
		h            = http.Header{}
	)

	uriOrMethod := args[0]
	if isMethod(uriOrMethod) {
		method = uriOrMethod
		uri, err = parseHost(args[1])
		if err != nil {
			return nil, err
		}

		args = args[2:]
	} else {
		uri, err = parseHost(uriOrMethod)
		if err != nil {
			return nil, err
		}

		method = http.MethodGet
		args = args[1:]
	}

	for _, arg := range args {
		if arg == "" {
			continue
		}

		kv, ok := parseKV(arg)
		if !ok {
			unknown <- arg
			continue
		}

		switch kv.kind {
		case fieldArg, jsonArg:
			_, err = fields.SetP(kv.v, kv.k)
			if err != nil {
				return
			}
		case headerArg:
			h.Add(kv.k, kv.v.(string))
		case queryArg:
			q.Add(kv.k, kv.v.(string))
		}

	}

	b := bytes.NewBuffer(fields.EncodeJSON())

	r, err = http.NewRequest(method, path, b)
	if err != nil {
		return
	}

	r.URL = uri
	r.URL.RawQuery = q.Encode()
	r.Header = h

	return
}

func isMethod(s string) bool {
	switch s {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut,
		http.MethodPatch, http.MethodDelete, http.MethodConnect,
		http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func parseKV(arg string) (res KV, ok bool) {
	g := re.FindStringSubmatch(arg)
	if len(g) != 4 {
		return
	}

	k := g[1]
	sp := g[2]
	v := g[3]

	res = KV{k: k, v: v}

	switch sp {
	case "==":
		res.kind = queryArg
	case ":":
		res.kind = headerArg
	case ":=":
		res.kind = jsonArg
		res.v = parseLiteral(v)
	case "=":
		res.kind = fieldArg
	}

	return res, true
}

func parseLiteral(v string) interface{} {
	var r interface{}
	if err := json.Unmarshal([]byte(v), &r); err != nil {
		log.Debug.Printf("could not parse %q: %v", v, err)
		return v
	}
	return r
}

func parseHost(s string) (*url.URL, error) {
	if s == ":" || s == "localhost" {
		s = "http://localhost/"
	}
	if strings.HasPrefix(s, ":") {
		s = fmt.Sprintf("http://localhost%s", s)
	}
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		s = fmt.Sprintf("http://%s", s)
	}

	u, err := url.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("parseHost: %v", err)
	}

	return u, nil
}
