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

package v1

import (
	"github.com/xsts-sh/xstatefulset/pkg/feature"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/utils/ptr"
)

func SetDefaults_XStatefulSet(obj *XStatefulSet) {
	if len(obj.Spec.PodManagementPolicy) == 0 {
		obj.Spec.PodManagementPolicy = appsv1.OrderedReadyPodManagement
	}

	if obj.Spec.UpdateStrategy.Type == "" {
		obj.Spec.UpdateStrategy.Type = appsv1.RollingUpdateStatefulSetStrategyType

		if obj.Spec.UpdateStrategy.RollingUpdate == nil {
			// UpdateStrategy.RollingUpdate will take default values below.
			obj.Spec.UpdateStrategy.RollingUpdate = &appsv1.RollingUpdateStatefulSetStrategy{}
		}
	}

	if obj.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType &&
		obj.Spec.UpdateStrategy.RollingUpdate != nil {

		if obj.Spec.UpdateStrategy.RollingUpdate.Partition == nil {
			obj.Spec.UpdateStrategy.RollingUpdate.Partition = ptr.To[int32](0)
		}
		if utilfeature.DefaultFeatureGate.Enabled(feature.MaxUnavailableStatefulSet) {
			if obj.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable == nil {
				obj.Spec.UpdateStrategy.RollingUpdate.MaxUnavailable = ptr.To(intstr.FromInt32(1))
			}
		}
	}

	if obj.Spec.PersistentVolumeClaimRetentionPolicy == nil {
		obj.Spec.PersistentVolumeClaimRetentionPolicy = &appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy{}
	}
	if len(obj.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted) == 0 {
		obj.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted = appsv1.RetainPersistentVolumeClaimRetentionPolicyType
	}
	if len(obj.Spec.PersistentVolumeClaimRetentionPolicy.WhenScaled) == 0 {
		obj.Spec.PersistentVolumeClaimRetentionPolicy.WhenScaled = appsv1.RetainPersistentVolumeClaimRetentionPolicyType
	}

	if obj.Spec.Replicas == nil {
		obj.Spec.Replicas = new(int32)
		*obj.Spec.Replicas = 1
	}
	if obj.Spec.RevisionHistoryLimit == nil {
		obj.Spec.RevisionHistoryLimit = new(int32)
		*obj.Spec.RevisionHistoryLimit = 10
	}
}

// addDefaultingFuncs is called by register.go to register the defaulting functions
// The actual RegisterDefaults function will be generated in zz_generated.defaults.go
func addDefaultingFuncs(scheme *runtime.Scheme) error {
	return RegisterDefaults(scheme)
}
