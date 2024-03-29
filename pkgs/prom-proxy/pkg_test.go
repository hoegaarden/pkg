package test

import (
	"testing"

	. "github.com/hoegaarden/pkg/testing"
)

func TestPromProxy(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		test       func(*JQer)
		dataValues []string
	}{
		"fails without htpasswd": {
			test: func(jq *JQer) {
				jq.HasLoadError()
			},
		},
		"defaults": {
			dataValues: []string{
				"htpasswd=something",
			},
			test: func(jq *JQer) {
				jq.
					// has cm with htpasswd set
					IsString(`.[] | select(.metadata.name == "prom-proxy-htpasswd") | .data.htpasswd | @base64d`, "something").
					// has one deployment
					IsNum(`.[] | select(.kind == "Deployment") | .spec.replicas`, 1).
					// has one service
					IsNum(`[ .[] | select(.kind == "Service") ] | length`, 1).
					// has configmap with the upstream
					MatchesString(`.[] | select(.kind == "ConfigMap") | .data["server.conf"]`, "proxy_pass http://prometheus-server:80/;").
					// has no ingress or that like
					IsEmpty(`.[] | select(.kind == "Ingress" or .kind == "HTTPProxy")`).
					// has not TLS secret
					IsEmpty(`.[] | select(.kind == "Secret" and .metadata.name == "prom-proxy-tls")`).
					// has limits and resources set
					IsTrue(`
						.[] | select(.kind == "Deployment")
							| .spec.template.spec.containers[].resources
							| (
								(.limits | type == "object")
									and
								(.requests | type == "object")
							)
					`)
			},
		},
		"provide TLS cert stuff inline": {
			dataValues: []string{
				"htpasswd=ignore",
				"ingress.type=contour",
				"ingress.fqdn=someFQDN",
				"ingress.tls.inline.cert=someCert",
				"ingress.tls.inline.key=someKey",
			},
			test: func(jq *JQer) {
				// a secret with the TLS data is created
				jq.IsString(`
					.[] | select(.kind == "Secret" and .metadata.name == "prom-proxy-tls")
						| .data | map_values(@base64d)
						| [ .["tls.crt"] , .["tls.key"] ]
						| join("|")
					`,
					"someCert|someKey",
				)
			},
		},
		"with ingress": {
			dataValues: []string{
				"htpasswd=ignore",
				"ingress.type=ingress",
				"ingress.fqdn=someFQDN",
			},
			test: func(jq *JQer) {
				jq.
					IsNum(`[ .[] | select(.kind == "Ingress") ] | length`, 1).
					IsString(`.[] | select(.kind == "Ingress") | .spec.rules[0].host`, "someFQDN").
					IsString(`.[] | select(.kind == "Ingress") | .spec.tls[0].hosts[0]`, "someFQDN").
					IsEmpty(`.[] | select(.kind == "HTTPProxy")`)
			},
		},
		"with httpproxy": {
			dataValues: []string{
				"htpasswd=ignore",
				"ingress.type=contour",
				"ingress.fqdn=someFQDN",
			},
			test: func(jq *JQer) {
				jq.
					IsNum(`[ .[] | select(.kind == "HTTPProxy") ] | length`, 1).
					IsString(`.[] | select(.kind == "HTTPProxy") | .spec.virtualhost.fqdn`, "someFQDN").
					IsEmpty(`.[] | select(.kind == "Ingress")`)
			},
		},
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			jq := NewJQer(t, Ytt{DataValues: tc.dataValues})
			tc.test(jq)
		})
	}
}
