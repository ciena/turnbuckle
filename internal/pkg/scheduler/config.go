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
package scheduler

import (
	"time"
)

type configSpec struct {
	Debug               bool
	MinDelayOnFailure   time.Duration
	MaxDelayOnFailure   time.Duration
	NumRetriesOnFailure int
	FallbackOnNoOffers  bool
	RetryOnNoOffers     bool
}

var schedulerConfig configSpec = configSpec{Debug: true,
	NumRetriesOnFailure: 3,
	MinDelayOnFailure:   time.Duration(30 * time.Second),
	MaxDelayOnFailure:   time.Duration(time.Minute),
	FallbackOnNoOffers:  false,
	RetryOnNoOffers:     false,
}
