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

package webhook

import (
	"context"
	"testing"

	xappsv1 "github.com/xsts-sh/xstatefulset/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestXStatefulSetDefaulter_Default(t *testing.T) {
	tests := []struct {
		name           string
		xsts           *xappsv1.XStatefulSet
		expectDefaults bool
	}{
		{
			name: "minimal xstatefulset should get defaults",
			xsts: &xappsv1.XStatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-xsts",
					Namespace: "default",
				},
				Spec: xappsv1.XStatefulSetSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			},
			expectDefaults: true,
		},
		{
			name: "xstatefulset with explicit values should preserve them",
			xsts: &xappsv1.XStatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-xsts-custom",
					Namespace: "default",
				},
				Spec: xappsv1.XStatefulSetSpec{
					Replicas: func() *int32 { r := int32(3); return &r }(),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "test"},
					},
					PodManagementPolicy: appsv1.ParallelPodManagement,
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "test"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "nginx",
									Image: "nginx:latest",
								},
							},
						},
					},
				},
			},
			expectDefaults: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaulter := &XStatefulSetDefaulter{}

			// Apply defaults
			err := defaulter.Default(context.Background(), tt.xsts)
			if err != nil {
				t.Fatalf("Default() error = %v", err)
			}

			if tt.expectDefaults {
				// Verify defaults were applied
				if tt.xsts.Spec.Replicas == nil {
					t.Error("expected replicas to be set")
				}

				if tt.xsts.Spec.PodManagementPolicy == "" {
					t.Error("expected podManagementPolicy to be set")
				}

				if tt.xsts.Spec.UpdateStrategy.Type == "" {
					t.Error("expected updateStrategy.type to be set")
				}

				if tt.xsts.Spec.RevisionHistoryLimit == nil {
					t.Error("expected revisionHistoryLimit to be set")
				}

				if tt.xsts.Spec.PersistentVolumeClaimRetentionPolicy == nil {
					t.Error("expected persistentVolumeClaimRetentionPolicy to be set")
				}
			}
		})
	}
}

func TestSetDefaults_XStatefulSet(t *testing.T) {
	xsts := &xappsv1.XStatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: xappsv1.XStatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "test"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "test"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "nginx", Image: "nginx:latest"},
					},
				},
			},
		},
	}

	// Apply defaults
	xappsv1.SetDefaults_XStatefulSet(xsts)

	// Verify defaults
	if xsts.Spec.Replicas == nil || *xsts.Spec.Replicas != 1 {
		t.Errorf("expected replicas=1, got %v", xsts.Spec.Replicas)
	}

	if xsts.Spec.PodManagementPolicy != appsv1.OrderedReadyPodManagement {
		t.Errorf("expected podManagementPolicy=OrderedReady, got %v", xsts.Spec.PodManagementPolicy)
	}

	if xsts.Spec.UpdateStrategy.Type != appsv1.RollingUpdateStatefulSetStrategyType {
		t.Errorf("expected updateStrategy.type=RollingUpdate, got %v", xsts.Spec.UpdateStrategy.Type)
	}

	if xsts.Spec.RevisionHistoryLimit == nil || *xsts.Spec.RevisionHistoryLimit != 10 {
		t.Errorf("expected revisionHistoryLimit=10, got %v", xsts.Spec.RevisionHistoryLimit)
	}

	if xsts.Spec.PersistentVolumeClaimRetentionPolicy == nil {
		t.Error("expected persistentVolumeClaimRetentionPolicy to be set")
	} else {
		if xsts.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted != appsv1.RetainPersistentVolumeClaimRetentionPolicyType {
			t.Errorf("expected whenDeleted=Retain, got %v", xsts.Spec.PersistentVolumeClaimRetentionPolicy.WhenDeleted)
		}
		if xsts.Spec.PersistentVolumeClaimRetentionPolicy.WhenScaled != appsv1.RetainPersistentVolumeClaimRetentionPolicyType {
			t.Errorf("expected whenScaled=Retain, got %v", xsts.Spec.PersistentVolumeClaimRetentionPolicy.WhenScaled)
		}
	}
}
