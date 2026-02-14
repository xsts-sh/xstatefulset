# API Reference

## Packages
- [apps.x-k8s.io/v1](#appsx-k8siov1)


## apps.x-k8s.io/v1


### Resource Types
- [XStatefulSet](#xstatefulset)
- [XStatefulSetList](#xstatefulsetlist)



#### XStatefulSet



XStatefulSet represents a set of pods with consistent identities.
Identities are defined as:
  - Network: A single stable DNS and hostname.
  - Storage: As many VolumeClaims as requested.

The StatefulSet guarantees that a given network identity will always
map to the same storage identity.



_Appears in:_
- [XStatefulSetList](#xstatefulsetlist)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `apps.x-k8s.io/v1` | | |
| `kind` _string_ | `XStatefulSet` | | |
| `spec` _[XStatefulSetSpec](#xstatefulsetspec)_ | Spec defines the desired identities of pods in this set. |  |  |
| `status` _[XStatefulSetStatus](#xstatefulsetstatus)_ | Status is the current status of Pods in this StatefulSet. This data<br />may be out of date by some window of time. |  |  |


#### XStatefulSetList



XStatefulSetList is a collection of StatefulSets.





| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `apiVersion` _string_ | `apps.x-k8s.io/v1` | | |
| `kind` _string_ | `XStatefulSetList` | | |
| `items` _[XStatefulSet](#xstatefulset) array_ | Items is the list of stateful sets. |  |  |


#### XStatefulSetSpec



A XStatefulSetSpec is the specification of a StatefulSet.



_Appears in:_
- [XStatefulSet](#xstatefulset)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `replicas` _integer_ | replicas is the desired number of replicas of the given Template.<br />These are replicas in the sense that they are instantiations of the<br />same Template, but individual replicas also have a consistent identity.<br />If unspecified, defaults to 1. |  |  |
| `template` _[PodTemplateSpec](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#podtemplatespec-v1-core)_ | template is the object that describes the pod that will be created if<br />insufficient replicas are detected. Each pod stamped out by the StatefulSet<br />will fulfill this Template, but have a unique identity from the rest<br />of the StatefulSet. Each pod will be named with the format<br /><statefulsetname>-<podindex>. For example, a pod in a StatefulSet named<br />"web" with index number "3" would be named "web-3".<br />The only allowed template.spec.restartPolicy value is "Always". |  |  |
| `volumeClaimTemplates` _[PersistentVolumeClaim](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#persistentvolumeclaim-v1-core) array_ | volumeClaimTemplates is a list of claims that pods are allowed to reference.<br />The StatefulSet controller is responsible for mapping network identities to<br />claims in a way that maintains the identity of a pod. Every claim in<br />this list must have at least one matching (by name) volumeMount in one<br />container in the template. A claim in this list takes precedence over<br />any volumes in the template, with the same name. |  |  |
| `serviceName` _string_ | serviceName is the name of the service that governs this StatefulSet.<br />This service must exist before the StatefulSet, and is responsible for<br />the network identity of the set. Pods get DNS/hostnames that follow the<br />pattern: pod-specific-string.serviceName.default.svc.cluster.local<br />where "pod-specific-string" is managed by the StatefulSet controller. |  |  |
| `podManagementPolicy` _[PodManagementPolicyType](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#podmanagementpolicytype-v1-apps)_ | podManagementPolicy controls how pods are created during initial scale up,<br />when replacing pods on nodes, or when scaling down. The default policy is<br />`OrderedReady`, where pods are created in increasing order (pod-0, then<br />pod-1, etc) and the controller will wait until each pod is ready before<br />continuing. When scaling down, the pods are removed in the opposite order.<br />The alternative policy is `Parallel` which will create pods in parallel<br />to match the desired scale without waiting, and on scale down will delete<br />all pods at once. |  |  |
| `updateStrategy` _[StatefulSetUpdateStrategy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#statefulsetupdatestrategy-v1-apps)_ | updateStrategy indicates the StatefulSetUpdateStrategy that will be<br />employed to update Pods in the StatefulSet when a revision is made to<br />Template. |  |  |
| `revisionHistoryLimit` _integer_ | revisionHistoryLimit is the maximum number of revisions that will<br />be maintained in the StatefulSet's revision history. The revision history<br />consists of all revisions not represented by a currently applied<br />XStatefulSetSpec version. The default value is 10. |  |  |
| `minReadySeconds` _integer_ | Minimum number of seconds for which a newly created pod should be ready<br />without any of its container crashing for it to be considered available.<br />Defaults to 0 (pod will be considered available as soon as it is ready) |  |  |
| `persistentVolumeClaimRetentionPolicy` _[StatefulSetPersistentVolumeClaimRetentionPolicy](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#statefulsetpersistentvolumeclaimretentionpolicy-v1-apps)_ | persistentVolumeClaimRetentionPolicy describes the lifecycle of persistent<br />volume claims created from volumeClaimTemplates. By default, all persistent<br />volume claims are created as needed and retained until manually deleted. This<br />policy allows the lifecycle to be altered, for example by deleting persistent<br />volume claims when their stateful set is deleted, or when their pod is scaled<br />down. |  |  |
| `ordinals` _[StatefulSetOrdinals](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#statefulsetordinals-v1-apps)_ | ordinals controls the numbering of replica indices in a StatefulSet. The<br />default ordinals behavior assigns a "0" index to the first replica and<br />increments the index by one for each additional replica requested. |  |  |


#### XStatefulSetStatus



XStatefulSetStatus represents the current state of a StatefulSet.



_Appears in:_
- [XStatefulSet](#xstatefulset)

| Field | Description | Default | Validation |
| --- | --- | --- | --- |
| `observedGeneration` _integer_ | observedGeneration is the most recent generation observed for this StatefulSet. It corresponds to the<br />StatefulSet's generation, which is updated on mutation by the API Server. |  |  |
| `replicas` _integer_ | replicas is the number of Pods created by the StatefulSet controller. |  |  |
| `readyReplicas` _integer_ | readyReplicas is the number of pods created for this StatefulSet with a Ready Condition. |  |  |
| `currentReplicas` _integer_ | currentReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version<br />indicated by currentRevision. |  |  |
| `updatedReplicas` _integer_ | updatedReplicas is the number of Pods created by the StatefulSet controller from the StatefulSet version<br />indicated by updateRevision. |  |  |
| `currentRevision` _string_ | currentRevision, if not empty, indicates the version of the StatefulSet used to generate Pods in the<br />sequence [0,currentReplicas). |  |  |
| `updateRevision` _string_ | updateRevision, if not empty, indicates the version of the StatefulSet used to generate Pods in the sequence<br />[replicas-updatedReplicas,replicas) |  |  |
| `collisionCount` _integer_ | collisionCount is the count of hash collisions for the StatefulSet. The StatefulSet controller<br />uses this field as a collision avoidance mechanism when it needs to create the name for the<br />newest ControllerRevision. |  |  |
| `conditions` _[StatefulSetCondition](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.33/#statefulsetcondition-v1-apps) array_ | Represents the latest available observations of a xstatefulset's current state. |  |  |
| `availableReplicas` _integer_ | Total number of available pods (ready for at least minReadySeconds) targeted by this xstatefulset. |  |  |
| `selector` _string_ | Selector is the label selector in string format for the pods managed by this xstatefulset.<br />This field is required for the scale subresource to work with HPA. |  |  |


