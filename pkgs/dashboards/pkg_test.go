package test

import (
	"testing"

	. "github.com/hoegaarden/pkg/testing"
)

func TestDashboards(t *testing.T) {
	t.Parallel()

	expectedObjs := float64(8)

	jq := NewJQer(t, Ytt{})

	subTest := func(name string, test func(jq *JQer)) {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			test(jq.WithT(t))
		})
	}

	subTest("has 8 things", func(jq *JQer) {
		jq.IsNum(`. | length`, expectedObjs)
	})

	subTest("each thing is a configmap", func(jq *JQer) {
		jq.IsTrue(`[.[].kind] | unique == ["ConfigMap"]`)
	})

	subTest(`all have the label "grafana-dashboard" set to "true"`, func(jq *JQer) {
		jq.IsNum(`[.[].metadata.labels["grafana-dashboard"] == "true"] | length`, expectedObjs)
	})

	subTest(`all have the label "grafana_dashboard" set to "true"`, func(jq *JQer) {
		jq.IsNum(`[.[].metadata.labels["grafana_dashboard"] == "true"] | length`, expectedObjs)
	})

	subTest(`all have a datum with a key "dashboard.*.json"`, func(jq *JQer) {
		jq.IsNum(`[.[].data | keys[] as $k | "\($k)"| test("^dashboard-.*\\.json$")] | length`, expectedObjs)
	})

	subTest(`all expected 8 keys are parsebale json`, func(jq *JQer) {
		jq.IsNum(`[.[].data | keys[] as $k | .[$k] | fromjson] | length`, expectedObjs)
	})
}
