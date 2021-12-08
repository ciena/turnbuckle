/*
Copyright 2021 Ciena Corporation.

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

package types

import (
	"testing"

	uv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestParseReference(t *testing.T) {
	var tests = []struct {
		d   string
		in  string
		out Reference
		err error
	}{
		{
			"all parts - good test",
			"cluster:namespace:api:Kind:name",
			Reference{Cluster: "cluster", Namespace: "namespace", APIVersion: "api",
				Kind: "Kind", Name: "name"},
			nil,
		},
		{
			"kind not capitalized",
			"cluster:namespace:api:kind:name",
			Reference{},
			ErrParseReference,
		},
		{
			"slash in API version",
			"cluster:namespace:api/version:Kind:name",
			Reference{Cluster: "cluster", Namespace: "namespace", APIVersion: "api/version",
				Kind: "Kind", Name: "name"},
			nil,
		},
		{
			"dots and slash in API version",
			"cluster:namespace:my.co/v1:Kind:name",
			Reference{Cluster: "cluster", Namespace: "namespace", APIVersion: "my.co/v1",
				Kind: "Kind", Name: "name"},
			nil,
		},
		{
			"empty string",
			"",
			Reference{},
			ErrParseReference,
		},
		{
			"no cluster specified",
			"ns:my.co/v1:Kind:name",
			Reference{Namespace: "ns", APIVersion: "my.co/v1",
				Kind: "Kind", Name: "name"},
			nil,
		},
		{
			"no cluster or namespace",
			"my.co/v1:Kind:name",
			Reference{APIVersion: "my.co/v1",
				Kind: "Kind", Name: "name"},
			nil,
		},
		{
			"empty cluster value",
			":namespace:my.co/v1:Kind:name",
			Reference{Cluster: "", Namespace: "namespace", APIVersion: "my.co/v1",
				Kind: "Kind", Name: "name"},
			nil,
		},
		{
			"empty cluster and namespace values",
			"::my.co/v1:Kind:name",
			Reference{Cluster: "", Namespace: "", APIVersion: "my.co/v1",
				Kind: "Kind", Name: "name"},
			nil,
		},
		{
			"empty cluster, namespace, and API version values",
			":::Kind:name",
			Reference{},
			ErrParseReference,
		},
		{
			"no name specified",
			"cluster:ns:my.co/v1:Kind",
			Reference{},
			ErrParseReference,
		},
		{
			"empty name specified",
			"cluster:ns:my.co/v1:Kind:",
			Reference{},
			ErrParseReference,
		},
	}

	for _, test := range tests {
		val, err := ParseReference(test.in)
		if err != test.err {
			t.Errorf("%s(%s):\ngot  %v,\nwant %v", test.d, test.in, err, test.err)
			continue
		}
		if err != nil || test.err != nil {
			continue
		}
		if *val != test.out {
			t.Errorf("%s(%s):\ngot  %v,\nwant %v", test.d, test.in, val, test.out)
			continue
		}
	}
}

func toUnstructured(namespace, apiVersion, kind, name string) uv1.Unstructured {
	u := uv1.Unstructured{}

	if len(namespace+apiVersion+kind+name) == 0 {
		return u
	}

	u.Object = make(map[string]interface{})
	if apiVersion != "" {
		u.Object["apiVersion"] = apiVersion
	}
	if kind != "" {
		u.Object["kind"] = kind
	}
	if len(namespace+name) == 0 {
		return u
	}
	md := make(map[string]interface{})
	if namespace != "" {
		md["namespace"] = namespace
	}
	if name != "" {
		md["name"] = name
	}
	u.Object["metadata"] = md

	return u
}

func TestReferenceFromUnstructured(t *testing.T) {
	var tests = []struct {
		d   string
		in  uv1.Unstructured
		out string
	}{
		{
			"fully populated",
			toUnstructured("ns", "my.co/v1", "MyCRD", "my-resource"),
			":ns:my.co/v1:MyCRD:my-resource",
		},
		{
			"empty",
			toUnstructured("", "", "", ""),
			"::::",
		},
		{
			"no meta data",
			toUnstructured("", "my.co/v1", "MyCRD", ""),
			"::my.co/v1:MyCRD:",
		},
		{
			"no type data",
			toUnstructured("ns", "", "", "my-resource"),
			":ns:::my-resource",
		},
		{
			"partial data, kind and name",
			toUnstructured("", "", "MyKind", "my-resource"),
			":::MyKind:my-resource",
		},
		{
			"partial data, just name",
			toUnstructured("", "", "", "my-resource"),
			"::::my-resource",
		},
	}

	for _, test := range tests {
		ref := NewReferenceFromUnstructured(test.in)
		if ref.String() != test.out {
			t.Errorf("%s:\ngot  %s\nwant %s\n", test.d, ref, test.out)
		}
	}
}
