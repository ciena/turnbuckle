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

package constraint

const (
	lkName      = "name"
	lkNamespace = "namespace"
	apiError    = "api-error"
)

//nolint:varnamelen
func areStringSlicesEqual(a, b []string) bool {
	// if both are nil then consider them equal
	if a == nil && b == nil {
		return true
	}

	// if one or the other is nil, then no equal
	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}
