package test

import (
	"testing"

	. "github.com/hoegaarden/pkg/testing"
)

func TestPackage(t *testing.T) {
	commonTests := func(jq *JQer) {
		jq.
			IsNum("length", 1).
			IsTrue(`.[0] | (.kind == "ConfigMap" and .apiVersion == "v1" and .data.some == "data")`).
			IsYamlString(".[0].data.foo")
	}

	jq := NewJQer(t, Ytt{})
	jq.
		IsString(".[0].metadata.name", "something").
		IsString(".[0].data.foo", "name: something\n")
	commonTests(jq)

	t.Run("with data values", func(t *testing.T) {
		ytt := Ytt{
			DVs: []DV{FileDV(
				"#@data/values",
				"---",
				"name: some-random-name",
			)},
		}

		jq := NewJQer(t, ytt)
		jq.
			IsString(".[0].metadata.name", "some-random-name").
			IsString(".[0].data.foo", "name: some-random-name\n")
		commonTests(jq)
	})
}
