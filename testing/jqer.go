package testing

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"sync"
	"testing"

	"github.com/itchyny/gojq"
	"gopkg.in/yaml.v2"
)

func NewJQer(t *testing.T, j JsonProvider) *JQer {
	return &JQer{T: t, Provider: j}
}

type JQer struct {
	T        *testing.T
	Provider JsonProvider

	data []interface{}
	once sync.Once
}

func (jq *JQer) getInput() {
	input := jq.Provider.Provide(jq.T)

	var data []interface{}

	dec := json.NewDecoder(bytes.NewReader(input))
	for {
		var r interface{}
		if err := dec.Decode(&r); err == io.EOF {
			break
		} else if err != nil {
			jq.T.Fatal(err)
			return
		}
		data = append(data, r)
	}

	jq.data = data
}

func (jq *JQer) GetRaw(q string) any {
	jq.T.Helper()

	jq.once.Do(jq.getInput)

	query, err := gojq.Parse(q)
	if err != nil {
		jq.T.Fatal(err)
	}

	var res []interface{}
	iter := query.Run(jq.data)
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			jq.T.Fatal(err)
		}
		res = append(res, v)
	}

	if l := len(res); l != 1 {
		jq.T.Fatalf("expected query '%s' to return exactly one result, got: %d", q, l)
	}

	return res[0]
}

func (jq *JQer) GetString(q string) string {
	jq.T.Helper()

	raw := jq.GetRaw(q)
	o, ok := raw.(string)
	if !ok {
		jq.T.Fatalf("'%s' is not a string, but a %T", q, raw)
		return ""
	}
	return o
}

func (jq *JQer) GetNum(q string) float64 {
	jq.T.Helper()

	raw := jq.GetRaw(q)

	switch v := raw.(type) {
	case int:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		jq.T.Fatalf("'%s' is not a number, but a %T", q, raw)
		return 0
	}
}

func (jq *JQer) GetBool(q string) bool {
	jq.T.Helper()

	raw := jq.GetRaw(q)
	o, ok := raw.(bool)
	if !ok {
		jq.T.Fatalf("'%s' is not a bool, but a %T", q, raw)
		return false
	}
	return o
}

func (jq *JQer) IsString(q, e string) *JQer {
	jq.T.Helper()
	if a := jq.GetString(q); e != a {
		expectedErr(jq.T, q, e, a)
	}
	return jq
}

func (jq *JQer) IsNum(q string, e float64) *JQer {
	jq.T.Helper()
	if a := jq.GetNum(q); e != a {
		expectedErr(jq.T, q, e, a)
	}
	return jq
}

func (jq *JQer) IsBool(q string, e bool) *JQer {
	jq.T.Helper()
	if a := jq.GetBool(q); e != a {
		expectedErr(jq.T, q, e, a)
	}
	return jq
}

func (jq *JQer) IsTrue(q string) *JQer {
	jq.T.Helper()
	return jq.IsBool(q, true)
}

func (jq *JQer) IsFalse(q string) *JQer {
	jq.T.Helper()
	return jq.IsBool(q, true)
}

func (jq *JQer) MatchesString(q, re string) *JQer {
	jq.T.Helper()

	s := jq.GetString(q)
	matched, err := regexp.MatchString(re, s)
	if err != nil {
		jq.T.Errorf("cannot run RE '%s': %v", re, err)
		return jq
	}

	if !matched {
		jq.T.Errorf("expected string(%s) at '%s' to match against re(%s), but doesn't", s, q, re)
	}

	return jq
}

type unmarshaler func([]byte, any) error

func (jq *JQer) canUnmarshal(q string, t string, f unmarshaler) *JQer {
	jq.T.Helper()
	xs := jq.GetString(q)
	var x interface{}
	if err := f([]byte(xs), &x); err != nil {
		jq.T.Errorf("expected string at '%s' to be a parse-able %s string, but isn't: %v", q, t, err)
	}
	return jq
}

func (jq *JQer) IsJsonString(q string) *JQer {
	jq.T.Helper()
	return jq.canUnmarshal(q, "json", json.Unmarshal)
}

func (jq *JQer) IsYamlString(q string) *JQer {
	jq.T.Helper()
	return jq.canUnmarshal(q, "yaml", yaml.Unmarshal)
}

func expectedErr(t *testing.T, q string, e, a interface{}) {
	t.Helper()
	t.Errorf("expected %T at '%s' to be %+v, got: %+v", e, q, e, a)
}
