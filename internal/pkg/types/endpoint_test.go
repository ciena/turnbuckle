package types

import (
	"encoding/json"
	"testing"
)

type epTestValues struct {
	EP            Endpoint
	String        string
	ErrorExpected bool
}

func toJSON(in interface{}) string {
	bytes, err := json.Marshal(in)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func TestParseEndpoint(t *testing.T) {
	vals := []epTestValues{
		{
			String: "name",
			EP: Endpoint{
				Name: "name",
			},
		},
		{
			String: "kind/name",
			EP: Endpoint{
				Kind: "kind",
				Name: "name",
			},
		},
		{
			String: "namespace:name",
			EP: Endpoint{
				Namespace: "namespace",
				Name:      "name",
			},
		},
		{
			String: "name[1.2.3.4]",
			EP: Endpoint{
				Name: "name",
				IP:   "1.2.3.4",
			},
		},
		{
			String: "cluster/:name",
			EP: Endpoint{
				Cluster: "cluster",
				Name:    "name",
			},
		},
		{
			String: "cluster/namespace:name",
			EP: Endpoint{
				Cluster:   "cluster",
				Namespace: "namespace",
				Name:      "name",
			},
		},
		{
			String: "cluster/:name[1.2.3.4]",
			EP: Endpoint{
				Cluster: "cluster",
				Name:    "name",
				IP:      "1.2.3.4",
			},
		},
		{
			String: "namespace:name[1.2.3.4]",
			EP: Endpoint{
				Namespace: "namespace",
				Name:      "name",
				IP:        "1.2.3.4",
			},
		},
		{
			String:        "",
			EP:            Endpoint{},
			ErrorExpected: true,
		},
		{
			String:        "[1.2.3.4]",
			EP:            Endpoint{},
			ErrorExpected: true,
		},
		{
			String:        "name[1.a.3.4]",
			EP:            Endpoint{},
			ErrorExpected: true,
		},
		{
			String:        "invalid&char",
			EP:            Endpoint{},
			ErrorExpected: true,
		},
	}
	for _, val := range vals {
		r, err := ParseEndpoint(val.String)
		if val.ErrorExpected && err == nil {
			t.Errorf("expected error when parsing Endpoint string '%s', but got none",
				val.String)
			continue
		}
		if !val.ErrorExpected && err != nil {
			t.Errorf("unexpected error when parsing Endpoint string '%s', but got '%s'",
				val.String, err.Error())
			continue
		}
		if val.ErrorExpected && err != nil {
			continue
		}
		if *r != val.EP {
			t.Errorf("unexpected results when parsing Endpoint string '%s', expected '%s', got '%s'",
				val.String, toJSON(val.EP), toJSON(r))
		}
	}
}

func TestEndpointString(t *testing.T) {
	vals := []epTestValues{
		{
			String: "name",
			EP: Endpoint{
				Name: "name",
			},
		},
		{
			String: "kind/name",
			EP: Endpoint{
				Kind: "kind",
				Name: "name",
			},
		},
		{
			String: "namespace:name",
			EP: Endpoint{
				Namespace: "namespace",
				Name:      "name",
			},
		},
		{
			String: "name[1.2.3.4]",
			EP: Endpoint{
				Name: "name",
				IP:   "1.2.3.4",
			},
		},
		{
			String: "cluster/default:name",
			EP: Endpoint{
				Cluster: "cluster",
				Name:    "name",
			},
		},
		{
			String: "cluster/namespace:name",
			EP: Endpoint{
				Cluster:   "cluster",
				Namespace: "namespace",
				Name:      "name",
			},
		},
		{
			String: "cluster/default:name[1.2.3.4]",
			EP: Endpoint{
				Cluster: "cluster",
				Name:    "name",
				IP:      "1.2.3.4",
			},
		},
		{
			String: "namespace:name[1.2.3.4]",
			EP: Endpoint{
				Namespace: "namespace",
				Name:      "name",
				IP:        "1.2.3.4",
			},
		},
		{
			String: "",
			EP:     Endpoint{},
		},
	}

	for _, val := range vals {
		if val.ErrorExpected {
			continue
		}
		s := val.EP.String()
		if s != val.String {
			t.Errorf("expected '%s' got '%s'", val.String, s)
		}
	}
}
