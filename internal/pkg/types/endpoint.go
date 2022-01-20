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
	"regexp"
	"strings"

	"github.com/ciena/turnbuckle/pkg/apis/ruleprovider"
)

// Endpoint a concreate end point reference.
type Endpoint struct {
	Cluster   string
	Namespace string
	Kind      string
	Name      string
	IP        string
}

// String returns the string representation of the Endpoint.
// nolint:gocritic
func (ep Endpoint) String() string {
	var buf strings.Builder

	if ep.Cluster != "" {
		buf.WriteString(ep.Cluster)
		buf.WriteString("/")
	}

	if ep.Namespace != "" {
		buf.WriteString(ep.Namespace)
		buf.WriteString(":")
	} else if ep.Cluster != "" {
		buf.WriteString("default:")
	}

	if ep.Kind != "" {
		buf.WriteString(ep.Kind)
		buf.WriteString("/")
	}

	buf.WriteString(ep.Name)

	if ep.IP != "" {
		buf.WriteString("[")
		buf.WriteString(ep.IP)
		buf.WriteString("]")
	}

	return buf.String()
}

var (
	// nolint:lll
	endpointRE = regexp.MustCompile(`^((([a-zA-Z0-9_-]*)/)?([a-zA-Z0-9-]*):)?(([a-zA-Z0-9_-]*)/)?([a-zA-Z0-9_-]+)(\[([0-9.]*)\])?$`)

	// ErrInvalidEndpoint is returned when an attempt to parse an invalid
	// string representation of an EndPoint is made.
	ErrInvalidEndpoint = errors.New("invaalid-endpoint")
)

// ParseEndpoint parses a string representation of an endpoint
// to a Endpoint.
func ParseEndpoint(in string) (*Endpoint, error) {
	var ep Endpoint

	parts := endpointRE.FindStringSubmatch(in)

	// fmt.Printf("%+#v\n", parts)
	if len(parts) == 0 {
		return nil, ErrInvalidEndpoint
	}

	if len(parts) > 0 {
		ep.Cluster = parts[3]
		ep.Namespace = parts[4]
		ep.Kind = parts[6]
		ep.Name = parts[7]
		ep.IP = parts[9]
	}

	return &ep, nil
}

// FromRuleProvider returns the Endpoint for the givne rule provider
// target.
// nolint:gocritic
func FromRuleProvider(ruleProviderEP ruleprovider.Target) Endpoint {
	return Endpoint{
		Cluster:   ruleProviderEP.Cluster,
		Kind:      ruleProviderEP.Kind,
		Namespace: ruleProviderEP.Namespace,
		Name:      ruleProviderEP.Name,
	}
}

// ToRuleProvider returns a rule provider target that represents the
// given ruleprovider Target.
// nolint:gocritic
func ToRuleProvider(endpoint Endpoint) ruleprovider.Target {
	return ruleprovider.Target{
		Cluster:   endpoint.Cluster,
		Kind:      endpoint.Kind,
		Namespace: endpoint.Namespace,
		Name:      endpoint.Name,
	}
}
