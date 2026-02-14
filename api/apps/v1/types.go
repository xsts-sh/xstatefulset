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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ControllerRevisionHashLabelKey = "controller-revision-hash"
	StatefulSetRevisionLabel       = ControllerRevisionHashLabelKey
	StatefulSetPodNameLabel        = "xstatefulset.x-k8s.io/pod-name"
	PodIndexLabel                  = "apps.x-k8s.io/pod-index"
)

// +genclient
// +genclient:method=GetScale,verb=get,subresource=scale,result=k8s.io/api/autoscaling/v1.Scale
// +genclient:method=UpdateScale,verb=update,subresource=scale,input=k8s.io/api/autoscaling/v1.Scale,result=k8s.io/api/autoscaling/v1.Scale
// +genclient:method=ApplyScale,verb=apply,subresource=scale,input=k8s.io/api/autoscaling/v1.Scale,result=k8s.io/api/autoscaling/v1.Scale
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:prerelease-lifecycle-gen:introduced=1.9
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=xsts
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="DESIRED",type="integer",description="Desired number of pods",JSONPath=".spec.replicas"
// +kubebuilder:printcolumn:name="CURRENT",type="integer",description="Current number of pods",JSONPath=".status.replicas"
// +kubebuilder:printcolumn:name="READY",type="integer",description="Number of ready pods",JSONPath=".status.readyReplicas"
// +kubebuilder:printcolumn:name="AGE",type="date",description="Creation timestamp",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

// XStatefulSet represents a set of pods with consistent identities.
// Identities are defined as:
//   - Network: A single stable DNS and hostname.
//   - Storage: As many VolumeClaims as requested.
//
// The StatefulSet guarantees that a given network identity will always
// map to the same storage identity.
type XStatefulSet struct {
	metav1.TypeMeta `json:",inline"`
	// Standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Spec defines the desired identities of pods in this set.
	// +optional
	Spec XStatefulSetSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`

	// Status is the current status of Pods in this StatefulSet. This data
	// may be out of date by some window of time.
	// +optional
	Status XStatefulSetStatus `json:"status,omitempty" protobuf:"bytes,3,opt,name=status"`
}

// A XStatefulSetSpec is the specification of a StatefulSet.
type XStatefulSetSpec struct {
	// replicas is the desired number of replicas of the given Template.
	// These are replicas in the sense that they are instantiations of the
	// same Template, but individual replicas also have a consistent identity.
	// If unspecified, defaults to 1.
	// TODO: Consider a rename of this field.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`

	// selector is a label query over pods that should match the replica count.
	// It must match the pod template's labels.
	// More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors
	Selector *metav1.LabelSelector `json:"selector" protobuf:"bytes,2,opt,name=selector"`

	// template is the object that describes the pod that will be created if
	// insufficient replicas are detected. Each pod stamped out by the StatefulSet
	// will fulfill this Template, but have a unique identity from the rest
	// of the StatefulSet. Each pod will be named with the format
	// <statefulsetname>-<podindex>. For example, a pod in a StatefulSet named
	// "web" with index number "3" would be named "web-3".
	// The only allowed template.spec.restartPolicy value is "Always".
	Template corev1.PodTemplateSpec `json:"template" protobuf:"bytes,3,opt,name=template"`

	// volumeClaimTemplates is a list of claims that pods are allowed to reference.
	// The StatefulSet controller is responsible for mapping network identities to
	// claims in a way that maintains the identity of a pod. Every claim in
	// this list must have at least one matching (by name) volumeMount in one
	// container in the template. A claim in this list takes precedence over
	// any volumes in the template, with the same name.
	// TODO: Define the behavior if a claim already exists with the same name.
	// +optional
	// +listType=atomic
	VolumeClaimTemplates []corev1.PersistentVolumeClaim `json:"volumeClaimTemplates,omitempty" protobuf:"bytes,4,rep,name=volumeClaimTemplates"`

	// serviceName is the name of the service that governs this StatefulSet.
	// This service must exist before the StatefulSet, and is responsible for
	// the network identity of the set. Pods get DNS/hostnames that follow the
	// pattern: pod-specific-string.serviceName.default.svc.cluster.local
	// where "pod-specific-string" is managed by the StatefulSet controller.
	// +optional
	ServiceName string `json:"serviceName" protobuf:"bytes,5,opt,name=serviceName"`

	// podManagementPolicy controls how pods are created during initial scale up,
	// when replacing pods on nodes, or when scaling down. The default policy is
	// `OrderedReady`, where pods are created in increasing order (pod-0, then
	// pod-1, etc) and the controller will wait until each pod is ready before
	// continuing. When scaling down, the pods are removed in the opposite order.
	// The alternative policy is `Parallel` which will create pods in parallel
	// to match the desired scale without waiting, and on scale down will delete
	// all pods at once.
	// +optional
	PodManagementPolicy appsv1.PodManagementPolicyType `json:"podManagementPolicy,omitempty" protobuf:"bytes,6,opt,name=podManagementPolicy,casttype=PodManagementPolicyType"`

	// updateStrategy indicates the StatefulSetUpdateStrategy that will be
	// employed to update Pods in the StatefulSet when a revision is made to
	// Template.
	UpdateStrategy appsv1.StatefulSetUpdateStrategy `json:"updateStrategy,omitempty" protobuf:"bytes,7,opt,name=updateStrategy"`

	// revisionHistoryLimit is the maximum number of revisions that will
	// be maintained in the StatefulSet's revision history. The revision history
	// consists of all revisions not represented by a currently applied
	// XStatefulSetSpec version. The default value is 10.
	RevisionHistoryLimit *int32 `json:"revisionHistoryLimit,omitempty" protobuf:"varint,8,opt,name=revisionHistoryLimit"`

	// Minimum number of seconds for which a newly created pod should be ready
	// without any of its container crashing for it to be considered available.
	// Defaults to 0 (pod will be considered available as soon as it is ready)
	// +optional
	MinReadySeconds int32 `json:"minReadySeconds,omitempty" protobuf:"varint,9,opt,name=minReadySeconds"`

	// persistentVolumeClaimRetentionPolicy describes the lifecycle of persistent
	// volume claims created from volumeClaimTemplates. By default, all persistent
	// volume claims are created as needed and retained until manually deleted. This
	// policy allows the lifecycle to be altered, for example by deleting persistent
	// volume claims when their stateful set is deleted, or when their pod is scaled
	// down.
	// +optional
	PersistentVolumeClaimRetentionPolicy *appsv1.StatefulSetPersistentVolumeClaimRetentionPolicy `json:"persistentVolumeClaimRetentionPolicy,omitempty" protobuf:"bytes,10,opt,name=persistentVolumeClaimRetentionPolicy"`

	// ordinals controls the numbering of replica indices in a StatefulSet. The
	// default ordinals behavior assigns a "0" index to the first replica and
	// increments the index by one for each additional replica requested.
	// +optional
	Ordinals *appsv1.StatefulSetOrdinals `json:"ordinals,omitempty" protobuf:"bytes,11,opt,name=ordinals"`
}

// XStatefulSetStatus represents the current state of a StatefulSet.
type XStatefulSetStatus struct {
	// observedGeneration is the most recent generation observed for this StatefulSet. It corresponds to the
	// StatefulSet's generation, which is updated on mutation by the API Server.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty" protobuf:"varint,1,opt,name=observedGeneration"`

	// replicas is the number of Pods created by the StatefulSet controller.
	Replicas int32 `json:"replicas" protobuf:"varint,2,opt,name=replicas"`

	// readyReplicas is the number of pods created for this StatefulSet with a Ready Condition.
	ReadyReplicas int32 `json:"readyReplicas,omitempty" protobuf:"varint,3,opt,name=readyReplicas"`

	// currentReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version
	// indicated by currentRevision.
	CurrentReplicas int32 `json:"currentReplicas,omitempty" protobuf:"varint,4,opt,name=currentReplicas"`

	// updatedReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version
	// indicated by updateRevision.
	UpdatedReplicas int32 `json:"updatedReplicas,omitempty" protobuf:"varint,5,opt,name=updatedReplicas"`

	// currentRevision, if not empty, indicates the version of the StatefulSet used to generate Pods in the
	// sequence [0,currentReplicas).
	CurrentRevision string `json:"currentRevision,omitempty" protobuf:"bytes,6,opt,name=currentRevision"`

	// updateRevision, if not empty, indicates the version of the StatefulSet used to generate Pods in the sequence
	// [replicas-updatedReplicas,replicas)
	UpdateRevision string `json:"updateRevision,omitempty" protobuf:"bytes,7,opt,name=updateRevision"`

	// collisionCount is the count of hash collisions for the StatefulSet. The StatefulSet controller
	// uses this field as a collision avoidance mechanism when it needs to create the name for the
	// newest ControllerRevision.
	// +optional
	CollisionCount *int32 `json:"collisionCount,omitempty" protobuf:"varint,9,opt,name=collisionCount"`

	// Represents the latest available observations of a xstatefulset's current state.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []appsv1.StatefulSetCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,10,rep,name=conditions"`

	// Total number of available pods (ready for at least minReadySeconds) targeted by this xstatefulset.
	// +optional
	AvailableReplicas int32 `json:"availableReplicas" protobuf:"varint,11,opt,name=availableReplicas"`

	// Selector is the label selector in string format for the pods managed by this xstatefulset.
	// This field is required for the scale subresource to work with HPA.
	// +optional
	Selector string `json:"selector,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +k8s:prerelease-lifecycle-gen:introduced=1.9
// +kubebuilder:object:root=true

// XStatefulSetList is a collection of StatefulSets.
type XStatefulSetList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	// Items is the list of stateful sets.
	Items []XStatefulSet `json:"items" protobuf:"bytes,2,rep,name=items"`
}
