/*
Copyright The XSTS-SH Authors.

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

package feature

import (
	"k8s.io/apimachinery/pkg/util/runtime"
	feature2 "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
)

const (

	// owner: @liggitt
	//
	// Mitigates spurious xstatefulset rollouts due to controller revision comparison mismatches
	// which are not semantically significant (e.g. serialization differences or missing defaulted fields).
	StatefulSetSemanticRevisionComparison = "StatefulSetSemanticRevisionComparison"
	// owner: @krmayankk
	// kep: https://kep.k8s.io/961
	//
	// Enables maxUnavailable for StatefulSet
	MaxUnavailableStatefulSet featuregate.Feature = "MaxUnavailableStatefulSet"
)

func init() {
	runtime.Must(feature2.DefaultMutableFeatureGate.Add(map[featuregate.Feature]featuregate.FeatureSpec{
		MaxUnavailableStatefulSet:             {Default: true, PreRelease: featuregate.Beta},
		StatefulSetSemanticRevisionComparison: {Default: true, PreRelease: featuregate.Beta},
	}))
}
