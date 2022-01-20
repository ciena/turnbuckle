/*
Copyright 2022 Ciena Corporation.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
imitations under the License.
*/

package types_test

import (
	"testing"

	"github.com/ciena/turnbuckle/internal/pkg/types"
	"github.com/stretchr/testify/assert"
)

type epTest struct {
	Name          string
	EP            types.Endpoint
	AsString      string
	ErrorExpected bool
}

func TestParseEndpoint(t *testing.T) {
	t.Parallel()

	tests := []epTest{
		{
			Name:     "name only",
			AsString: "name",
			EP: types.Endpoint{
				Name: "name",
			},
		},
		{
			Name:     "kind and name",
			AsString: "kind/name",
			EP: types.Endpoint{
				Kind: "kind",
				Name: "name",
			},
		},
		{
			Name:     "namespace and name",
			AsString: "namespace:name",
			EP: types.Endpoint{
				Namespace: "namespace",
				Name:      "name",
			},
		},
		{
			Name:     "name and IP",
			AsString: "name[1.2.3.4]",
			EP: types.Endpoint{
				Name: "name",
				IP:   "1.2.3.4",
			},
		},
		{
			Name:     "cluster and name",
			AsString: "cluster/:name",
			EP: types.Endpoint{
				Cluster: "cluster",
				Name:    "name",
			},
		},
		{
			Name:     "cluster, namespace, and name",
			AsString: "cluster/namespace:name",
			EP: types.Endpoint{
				Cluster:   "cluster",
				Namespace: "namespace",
				Name:      "name",
			},
		},
		{
			Name:     "cluster, name, and IP",
			AsString: "cluster/:name[1.2.3.4]",
			EP: types.Endpoint{
				Cluster: "cluster",
				Name:    "name",
				IP:      "1.2.3.4",
			},
		},
		{
			Name:     "namespace, name, and IP",
			AsString: "namespace:name[1.2.3.4]",
			EP: types.Endpoint{
				Namespace: "namespace",
				Name:      "name",
				IP:        "1.2.3.4",
			},
		},
		{
			Name:          "empty string",
			AsString:      "",
			EP:            types.Endpoint{},
			ErrorExpected: true,
		},
		{
			Name:          "IP only",
			AsString:      "[1.2.3.4]",
			EP:            types.Endpoint{},
			ErrorExpected: true,
		},
		{
			Name:          "name and invalid IP",
			AsString:      "name[1.a.3.4]",
			EP:            types.Endpoint{},
			ErrorExpected: true,
		},
		{
			Name:          "invalid character",
			AsString:      "invalid&char",
			EP:            types.Endpoint{},
			ErrorExpected: true,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			r, err := types.ParseEndpoint(tc.AsString)
			if tc.ErrorExpected {
				assert.NotNil(t, err, "expected error no seen")

				return
			}

			assert.Nil(t, err, "unexpected error returned")
			assert.Equal(t, tc.EP, *r, "invalid enpoint parse")
		})
	}
}

func TestEndpointToString(t *testing.T) {
	t.Parallel()

	tests := []epTest{
		{
			Name:     "name only",
			AsString: "name",
			EP: types.Endpoint{
				Name: "name",
			},
		},
		{
			Name:     "kind and name",
			AsString: "kind/name",
			EP: types.Endpoint{
				Kind: "kind",
				Name: "name",
			},
		},
		{
			Name:     "namespace and name",
			AsString: "namespace:name",
			EP: types.Endpoint{
				Namespace: "namespace",
				Name:      "name",
			},
		},
		{
			Name:     "name and IP",
			AsString: "name[1.2.3.4]",
			EP: types.Endpoint{
				Name: "name",
				IP:   "1.2.3.4",
			},
		},
		{
			Name:     "cluster and name",
			AsString: "cluster/default:name",
			EP: types.Endpoint{
				Cluster: "cluster",
				Name:    "name",
			},
		},
		{
			Name:     "cluster, namespace, and name",
			AsString: "cluster/namespace:name",
			EP: types.Endpoint{
				Cluster:   "cluster",
				Namespace: "namespace",
				Name:      "name",
			},
		},
		{
			Name:     "cluster, name, and IP",
			AsString: "cluster/default:name[1.2.3.4]",
			EP: types.Endpoint{
				Cluster: "cluster",
				Name:    "name",
				IP:      "1.2.3.4",
			},
		},
		{
			Name:     "namespace, name, and IP",
			AsString: "namespace:name[1.2.3.4]",
			EP: types.Endpoint{
				Namespace: "namespace",
				Name:      "name",
				IP:        "1.2.3.4",
			},
		},
		{
			Name:     "empty string",
			AsString: "",
			EP:       types.Endpoint{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			s := tc.EP.String()
			assert.Equal(t, tc.AsString, s)
		})
	}
}
