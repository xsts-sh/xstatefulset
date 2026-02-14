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

/*
Copyright 2017 The Kubernetes Authors.

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

package xstatefulset

import (
	"context"

	xstsappv1 "github.com/xsts-sh/xstatefulset/api/apps/v1"
	clientset "github.com/xsts-sh/xstatefulset/client-go/clientset/versioned"
	appslisters "github.com/xsts-sh/xstatefulset/client-go/listers/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

// StatefulSetStatusUpdaterInterface is an interface used to update the StatefulSetStatus associated with a StatefulSet.
// For any use other than testing, clients should create an instance using NewRealStatefulSetStatusUpdater.
type StatefulSetStatusUpdaterInterface interface {
	// UpdateStatefulSetStatus sets the set's Status to status. Implementations are required to retry on conflicts,
	// but fail on other errors. If the returned error is nil set's Status has been successfully set to status.
	UpdateStatefulSetStatus(ctx context.Context, set *xstsappv1.XStatefulSet, status *xstsappv1.XStatefulSetStatus) error
}

// NewRealStatefulSetStatusUpdater returns a StatefulSetStatusUpdaterInterface that updates the Status of a StatefulSet,
// using the supplied client and setLister.
func NewRealStatefulSetStatusUpdater(
	client clientset.Interface,
	setLister appslisters.XStatefulSetLister) StatefulSetStatusUpdaterInterface {
	return &realStatefulSetStatusUpdater{client, setLister}
}

type realStatefulSetStatusUpdater struct {
	client    clientset.Interface
	setLister appslisters.XStatefulSetLister
}

func (ssu *realStatefulSetStatusUpdater) UpdateStatefulSetStatus(
	ctx context.Context,
	set *xstsappv1.XStatefulSet,
	status *xstsappv1.XStatefulSetStatus) error {
	logger := klog.FromContext(ctx)
	// don't wait due to limited number of clients, but backoff after the default number of steps
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		set.Status = *status
		// TODO: This context.TODO should use a real context once we have RetryOnConflictWithContext
		_, updateErr := ssu.client.AppsV1().XStatefulSets(set.Namespace).UpdateStatus(context.TODO(), set, metav1.UpdateOptions{})
		if updateErr == nil {
			return nil
		}
		if updated, err := ssu.setLister.XStatefulSets(set.Namespace).Get(set.Name); err == nil {
			// make a copy so we don't mutate the shared cache
			set = updated.DeepCopy()
		} else {
			utilruntime.HandleErrorWithLogger(logger, nil, "Error getting updated StatefulSet from lister", "StatefulSet", klog.KObj(set))
		}

		return updateErr
	})
}

var _ StatefulSetStatusUpdaterInterface = &realStatefulSetStatusUpdater{}
