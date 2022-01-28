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

package types_test

import (
	"testing"

	. "github.com/ciena/turnbuckle/pkg/types"
)

func TestCompareComplianceSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Desc        string
		Left, Right string
		Result      int
	}{
		{"equal-0", "", "", 0},
		{"equal-1", CompliancePending, CompliancePending, 0},
		{"equal-2", ComplianceCompliant, ComplianceCompliant, 0},
		{"equal-3", ComplianceLimit, ComplianceLimit, 0},
		{"equal-4", ComplianceViolation, ComplianceViolation, 0},
		{"equal-5", ComplianceError, ComplianceError, 0},
		{"invalid-left", "foo", ComplianceError, 1},
		{"invalid-right", ComplianceError, "foo", -1},
		{"invalid-both", "bar", "foo", 0},
		{"error v violation", ComplianceError, ComplianceViolation, -1},
		{"violation v error", ComplianceViolation, ComplianceError, 1},
		{"pending v limit", CompliancePending, ComplianceLimit, 2},
		{"limit v pending", ComplianceLimit, CompliancePending, -2},
	}

	for _, test := range tests {
		val := CompareComplianceSeverity(test.Left, test.Right)
		if val != test.Result {
			t.Errorf("FAILED: %s (%s v %s): EXPECT(%d), GOT(%d)\n",
				test.Desc, test.Left, test.Right, test.Result, val)
		}
	}
}

func TestComplanceString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Level ComplianceLevel
		S     string
	}{
		{ComplianceNone, ""},
		{CompliancePending, "Pending"},
		{ComplianceCompliant, "Compliant"},
		{ComplianceLimit, "Limit"},
		{ComplianceViolation, "Violation"},
		{ComplianceError, "Error"},
	}

	for _, test := range tests {
		val := test.Level.String()
		if val != test.S {
			t.Errorf("FAILED: %s: EXPECT(%s), GOT(%s)\n",
				test.Level, test.S, val)
		}
	}
}
