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

// nolint:testpackage
package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAreStringSlicesEqual(t *testing.T) {
	t.Parallel()

	s1 := []string{"a", "b", "c"}
	s2 := []string{"a", "b", "c", "d"}
	s3 := []string{"a", "b", "z"}
	s4 := []string{"x", "y", "z"}
	tests := []struct {
		Name        string
		Left, Right []string
		Expected    bool
	}{
		{"left nil, right nil", nil, nil, true},
		{"left nil, right empty", nil, []string{}, false},
		{"left empty, right nil", []string{}, nil, false},
		{"left empty, right empty", []string{}, []string{}, true},
		{"left nil, right populated", nil, s1, false},
		{"left populated, right nil", s1, nil, false},
		{"identical", s1, s1, true},
		{"left subset of right", s1, s2, false},
		{"right subset or left", s2, s1, false},
		{"same length, partial", s1, s3, false},
		{"same length, partial (reversed)", s3, s1, false},
		{"same length, no match", s1, s4, false},
		{"same length, no match (reversed)", s4, s1, false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			result := areStringSlicesEqual(tc.Left, tc.Right)
			assert.Equal(t, tc.Expected, result)
		})
	}
}
