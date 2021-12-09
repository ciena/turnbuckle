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
limitations under the License.
*/

package types

import (
	"errors"
	"fmt"
	"hash/fnv"
	"regexp"
	"sort"
	"strings"

	uv1 "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/rand"
)

// Reference defines a cluster specific reference to a resource.
type Reference struct {
	Cluster    string `json:"cluster,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	APIVersion string `json:"apiversion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
}

type References []*Reference
type ReferencesMap map[string]References
type ReferencesList []References

func (r References) AsBindingName(offerName string) string {
	hash := fnv.New32a()
	for _, ref := range r {
		hash.Write([]byte(ref.String()))
	}
	return offerName + "-" + rand.SafeEncodeString(fmt.Sprint(hash.Sum32()))
}

func (r References) Contains(ref *Reference) bool {
	for _, have := range r {
		if *have == *ref {
			return true
		}
	}
	return false
}

func (m ReferencesMap) Permutations() ReferencesList {

	var inc func(list ReferencesList, refIdxs []int, refIdx int) bool
	inc = func(list ReferencesList, refIdxs []int, refIdx int) bool {
		if refIdx < 0 {
			return true
		}
		refIdxs[refIdx]++
		if refIdxs[refIdx] >= len(list[refIdx]) {
			refIdxs[refIdx] = 0
			return inc(list, refIdxs, refIdx-1)
		}

		return false
	}

	// if any of the map entries are empty then we have no permutations
	for _, v := range m {
		if len(v) == 0 {
			return ReferencesList{}
		}
	}

	// Sort the keys in the map so we have consistent binding naming
	// conventions
	var keys []string
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(a, b int) bool {
		return keys[a] < keys[b]
	})

	// In key order add each references object to an internal referenes list
	var list ReferencesList
	for _, key := range keys {
		list = append(list, m[key])
	}

	// Generate permutation slice.
	var permutations ReferencesList
	var refIdxs = make([]int, len(m))
	for {
		var permutation References
		for i := 0; i < len(list); i++ {
			permutation = append(permutation, list[i][refIdxs[i]])
		}
		permutations = append(permutations, permutation)

		if inc(list, refIdxs, len(refIdxs)-1) {
			break
		}
	}

	return permutations
}

// String marshal a Reference value into a string.
func (t Reference) String() string {
	var buf strings.Builder

	buf.WriteString(t.Cluster)
	buf.WriteString(":")
	buf.WriteString(t.Namespace)
	buf.WriteString(":")
	buf.WriteString(t.APIVersion)
	buf.WriteString(":")
	buf.WriteString(t.Kind)
	buf.WriteString(":")
	buf.WriteString(t.Name)

	return buf.String()
}

// the regex expression used to parse a string representation of a Reference.
var referenceRE = regexp.MustCompile(`^((([a-zA-Z_][a-zA-Z0-9-_]*)?:)?([a-zA-Z_][a-zA-Z0-9-_]*)?:)?([a-zA-Z_][a-zA-Z0-9-_\/\.]*):([A-Z][a-zA-Z0-9]*):([a-zA-Z_][a-zA-Z0-9-_]*)$`)

// ErrParseReference returned when a given string cannot be parsed as a Reference.
var ErrParseReference = errors.New("parse-reference")

// ErrConvertReference return when an unstructured object does not have
// complete information to be converted to a Reference instance.
var ErrConvertReference = errors.New("convert-unstructured-to-reference")

// ParseReference attempts to parse the given string as a Reference and returns
// the value or an error if it cannot be parsed as a Reference.
func ParseReference(in string) (*Reference, error) {
	var t Reference

	parts := referenceRE.FindStringSubmatch(in)
	if len(parts) == 0 {
		return nil, ErrParseReference
	}

	t.Cluster = parts[3]
	t.Namespace = parts[4]
	t.APIVersion = parts[5]
	t.Kind = parts[6]
	t.Name = parts[7]

	return &t, nil
}

// NewReferenceFromUnstructured creates and returns a new Reference instance
// from the given unstructured resource information.
func NewReferenceFromUnstructured(in uv1.Unstructured) *Reference {
	out := Reference{}
	if val, ok := in.Object["apiVersion"]; ok {
		out.APIVersion = val.(string)
	}
	if val, ok := in.Object["kind"]; ok {
		out.Kind = val.(string)
	}
	if md, ok := in.Object["metadata"].(map[string]interface{}); ok {
		if val, ok := md["namespace"].(string); ok {
			out.Namespace = val
		}
		if val, ok := md["name"].(string); ok {
			out.Name = val
		}
	}
	return &out
}
