package test

import (
	"testing"

	. "github.com/hoegaarden/pkg/testing"
)

func TestPackage(t *testing.T) {
	t.Parallel()

	t.Run("with no data values, with defaults", func(t *testing.T) {
		NewJQer(t, Ytt{}).
			IsTrue(`.[0] | (.kind == "ConfigMap" and .apiVersion == "v1" and .data.some == "data")`).
			IsTrue(`.[1] | (.kind == "Secret" and .apiVersion == "v1" and (.data.some|@base64d) == "data" )`).
			IsYamlString(".[0].data.foo").
			IsString(".[0].metadata.name", "cm").
			IsString(".[1].metadata.name", "sec").
			IsString(".[0].data.foo", "secretName: sec\nconfigMapName: cm\n")
	})

	t.Run("with data values", func(t *testing.T) {
		t.Parallel()

		testCases := map[string]struct {
			dvs  []DV
			test func(*JQer)
		}{
			"setting configMapName": {
				dvs: []DV{PlainDV("configMapName", "some-cm")},
				test: func(jq *JQer) {
					jq.IsString(".[0].metadata.name", "some-cm").
						MatchesString(".[0].data.foo", "configMapName: some-cm")
				},
			},
			"setting secretName": {
				dvs: []DV{PlainDV("secretName", "some-secret")},
				test: func(jq *JQer) {
					jq.IsString(".[1].metadata.name", "some-secret").
						MatchesString(".[0].data.foo", "secretName: some-secret")
				},
			},
			"setting secretName and configMapName": {
				dvs: []DV{FileDV(
					"#@data/values",
					"---",
					"configMapName: some-cm",
					"secretName: some-secret",
				)},
				test: func(jq *JQer) {
					jq.IsString(".[1].metadata.name", "some-secret").
						MatchesString(".[0].data.foo", "secretName: some-secret").
						MatchesString(".[0].data.foo", "configMapName: some-cm")
				},
			},
		}

		for name, tc := range testCases {
			tc := tc
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				jq := NewJQer(t, Ytt{DVs: tc.dvs})
				tc.test(jq)
			})
		}
	})
}
