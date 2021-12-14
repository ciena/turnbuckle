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

package types_test

import (
	"errors"
	"testing"

	. "github.com/ciena/turnbuckle/pkg/types"
	uv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestParseReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		d   string
		in  string
		out Reference
		err error
	}{
		{
			"all parts - good test",
			"cluster:namespace:api:Kind:name",
			Reference{
				Cluster: "cluster", Namespace: "namespace", APIVersion: "api",
				Kind: "Kind", Name: "name",
			},
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
			Reference{
				Cluster: "cluster", Namespace: "namespace", APIVersion: "api/version",
				Kind: "Kind", Name: "name",
			},
			nil,
		},
		{
			"dots and slash in API version",
			"cluster:namespace:my.co/v1:Kind:name",
			Reference{
				Cluster: "cluster", Namespace: "namespace", APIVersion: "my.co/v1",
				Kind: "Kind", Name: "name",
			},
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
			Reference{
				Namespace: "ns", APIVersion: "my.co/v1",
				Kind: "Kind", Name: "name",
			},
			nil,
		},
		{
			"no cluster or namespace",
			"my.co/v1:Kind:name",
			Reference{
				APIVersion: "my.co/v1",
				Kind:       "Kind", Name: "name",
			},
			nil,
		},
		{
			"empty cluster value",
			":namespace:my.co/v1:Kind:name",
			Reference{
				Cluster: "", Namespace: "namespace", APIVersion: "my.co/v1",
				Kind: "Kind", Name: "name",
			},
			nil,
		},
		{
			"empty cluster and namespace values",
			"::my.co/v1:Kind:name",
			Reference{
				Cluster: "", Namespace: "", APIVersion: "my.co/v1",
				Kind: "Kind", Name: "name",
			},
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
		if !errors.Is(err, test.err) {
			t.Errorf("%s(%s):\ngot  %v,\nwant %v", test.d, test.in, err, test.err)

			continue
		}

		if err != nil || test.err != nil {
			continue
		}

		if *val != test.out {
			t.Errorf("%s(%s):\ngot  %v\nwant %v", test.d, test.in, val, test.out)

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
	t.Parallel()

	tests := []struct {
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

func TestAsBindingName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		d     string
		offer string
		in    []string
		out   string
	}{
		{
			"empty reference list",
			"offer",
			[]string{},
			"offer-65bb57b6b5",
		},
		{
			"simple",
			"offer",
			[]string{":ns:v1:Pod:one", ":ns:v1:Pod:two"},
			"offer-6567d96b5f",
		},
		{
			"simple, reveresed",
			"offer",
			[]string{":ns:v1:Pod:two", ":ns:v1:Pod:one"},
			"offer-7c648b467",
		},
	}
	for _, test := range tests {
		rl := ReferenceList{}

		for _, v := range test.in {
			r, err := ParseReference(v)
			if err != nil {
				t.Errorf("converting reference: %s", err.Error())
			}

			rl = append(rl, r)
		}

		val := rl.AsBindingName(test.offer)
		if val != test.out {
			t.Errorf("%s:\ngot  %s\nwant %s\n", test.d, val, test.out)
		}
	}
}

func TestReferenceListContains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		d      string
		list   []string
		target string
		out    bool
	}{
		{
			"empty reference list",
			[]string{},
			":ns:v1:Pod:one",
			false,
		},
		{
			"non-empty list, exists in list",
			[]string{":ns:v1:Pod:one", ":ns:v1:Pod:two"},
			":ns:v1:Pod:one",
			true,
		},
		{
			"non-empty list, does not exist in list",
			[]string{":ns:v1:Pod:one", ":ns:v1:Pod:two"},
			":ns:v1:Pod:three",
			false,
		},
	}
	for _, test := range tests {
		rl := ReferenceList{}

		for _, v := range test.list {
			r, err := ParseReference(v)
			if err != nil {
				t.Errorf("converting reference: %s", err.Error())
			}

			rl = append(rl, r)
		}

		ref, err := ParseReference(test.target)
		if err != nil {
			t.Errorf("converting target: %s", err.Error())
		}

		val := rl.Contains(ref)
		if val != test.out {
			t.Errorf("%s:\ngot  %t\nwant %t\n", test.d, val, test.out)
		}
	}
}

func TestPermutations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		d       string
		listmap map[string][]string
		out     [][]string
	}{
		{
			"empty map of lists",
			map[string][]string{},
			[][]string{},
		},
		{
			"map with one empty lists",
			map[string][]string{
				"one": {":ns:v1:Pod:one"},
				"two": {},
			},
			[][]string{},
		},
		{
			"two sets, one item each, in order",
			map[string][]string{
				"one": {":ns:v1:Pod:one"},
				"two": {":ns:v1:Pod:two"},
			},
			[][]string{
				{":ns:v1:Pod:one", ":ns:v1:Pod:two"},
			},
		},
		{
			"two sets, one item each, reverse order",
			map[string][]string{
				"two": {":ns:v1:Pod:two"},
				"one": {":ns:v1:Pod:one"},
			},
			[][]string{
				{":ns:v1:Pod:one", ":ns:v1:Pod:two"},
			},
		},
		{
			"count to 16",
			map[string][]string{
				"a": {
					":ns:v1:Pod:a-0",
					":ns:v1:Pod:a-1",
				},
				"b": {
					":ns:v1:Pod:b-0",
					":ns:v1:Pod:b-1",
				},
				"c": {
					":ns:v1:Pod:c-0",
					":ns:v1:Pod:c-1",
				},
				"d": {
					":ns:v1:Pod:d-0",
					":ns:v1:Pod:d-1",
				},
			},
			[][]string{
				{":ns:v1:Pod:a-0", ":ns:v1:Pod:b-0", ":ns:v1:Pod:c-0", ":ns:v1:Pod:d-0"},
				{":ns:v1:Pod:a-0", ":ns:v1:Pod:b-0", ":ns:v1:Pod:c-0", ":ns:v1:Pod:d-1"},
				{":ns:v1:Pod:a-0", ":ns:v1:Pod:b-0", ":ns:v1:Pod:c-1", ":ns:v1:Pod:d-0"},
				{":ns:v1:Pod:a-0", ":ns:v1:Pod:b-0", ":ns:v1:Pod:c-1", ":ns:v1:Pod:d-1"},
				{":ns:v1:Pod:a-0", ":ns:v1:Pod:b-1", ":ns:v1:Pod:c-0", ":ns:v1:Pod:d-0"},
				{":ns:v1:Pod:a-0", ":ns:v1:Pod:b-1", ":ns:v1:Pod:c-0", ":ns:v1:Pod:d-1"},
				{":ns:v1:Pod:a-0", ":ns:v1:Pod:b-1", ":ns:v1:Pod:c-1", ":ns:v1:Pod:d-0"},
				{":ns:v1:Pod:a-0", ":ns:v1:Pod:b-1", ":ns:v1:Pod:c-1", ":ns:v1:Pod:d-1"},
				{":ns:v1:Pod:a-1", ":ns:v1:Pod:b-0", ":ns:v1:Pod:c-0", ":ns:v1:Pod:d-0"},
				{":ns:v1:Pod:a-1", ":ns:v1:Pod:b-0", ":ns:v1:Pod:c-0", ":ns:v1:Pod:d-1"},
				{":ns:v1:Pod:a-1", ":ns:v1:Pod:b-0", ":ns:v1:Pod:c-1", ":ns:v1:Pod:d-0"},
				{":ns:v1:Pod:a-1", ":ns:v1:Pod:b-0", ":ns:v1:Pod:c-1", ":ns:v1:Pod:d-1"},
				{":ns:v1:Pod:a-1", ":ns:v1:Pod:b-1", ":ns:v1:Pod:c-0", ":ns:v1:Pod:d-0"},
				{":ns:v1:Pod:a-1", ":ns:v1:Pod:b-1", ":ns:v1:Pod:c-0", ":ns:v1:Pod:d-1"},
				{":ns:v1:Pod:a-1", ":ns:v1:Pod:b-1", ":ns:v1:Pod:c-1", ":ns:v1:Pod:d-0"},
				{":ns:v1:Pod:a-1", ":ns:v1:Pod:b-1", ":ns:v1:Pod:c-1", ":ns:v1:Pod:d-1"},
			},
		},
	}
	for _, test := range tests {
		rm := ReferenceListMap{}

		for k, v := range test.listmap {
			rl := ReferenceList{}

			for _, s := range v {
				r, err := ParseReference(s)
				if err != nil {
					t.Errorf("converting reference: %s", err.Error())
				}

				rl = append(rl, r)
			}

			rm[k] = rl
		}

		_, ps := rm.Permutations()
		if len(test.out) != len(ps) {
			t.Errorf("'%s': unequal permutation count: %d v. %d\n",
				test.d, len(test.out), len(ps))
		}
	}
}
