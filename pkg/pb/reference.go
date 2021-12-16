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

package bp

import (
	"github.com/ciena/turnbuckle/apis/ruleprovider"
	"github.com/ciena/turnbuckle/pkg/types"
)

// AsPBTarget converts the k8s representation of a reference to a
// protobuf representation.
func AsPBTarget(from map[string]types.Reference) map[string]*ruleprovider.Target {
	to := make(map[string]*ruleprovider.Target)

	// nolint: gocritic
	for k, v := range from {
		to[k] = &ruleprovider.Target{
			Cluster:    v.Cluster,
			Namespace:  v.Namespace,
			ApiVersion: v.APIVersion,
			Kind:       v.Kind,
			Name:       v.Name,
		}
	}

	return to
}
