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

package controller

import (
	"context"
	"math/rand"
	"time"

	clientset "github.com/xsts-sh/xstatefulset/client-go/clientset/versioned"
	xStatefulSetInformers "github.com/xsts-sh/xstatefulset/client-go/informers/externalversions"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

// ControllerContext
type ControllerContext struct {
	// InformerFactory gives access to informers for the controller.
	KubeInformerFactory informers.SharedInformerFactory

	XStatefulsetInformerFactory xStatefulSetInformers.SharedInformerFactory

	// InformersStarted is closed after all of the controllers have been initialized and are running.  After this point it is safe,
	// for an individual controller to start the shared informers. Before it is closed, they should not.
	InformersStarted chan struct{}

	// ResyncPeriod generates a duration each time it is invoked; this is so that
	// multiple controllers don't get into lock-step and all hammer the apiserver
	// with list requests simultaneously.
	ResyncPeriod func() time.Duration
}

func NewControllerContext(ctx context.Context, versionedClient kubernetes.Interface, xStatefulSetClient clientset.Interface) *ControllerContext {
	// Informer transform to trim ManagedFields for memory efficiency.
	trim := func(obj interface{}) (interface{}, error) {
		if accessor, err := meta.Accessor(obj); err == nil {
			if accessor.GetManagedFields() != nil {
				accessor.SetManagedFields(nil)
			}
		}
		return obj, nil
	}
	kubeSharedInformers := informers.NewSharedInformerFactoryWithOptions(versionedClient, 0, informers.WithTransform(trim))
	xStafulsetInformer := xStatefulSetInformers.NewSharedInformerFactoryWithOptions(xStatefulSetClient, 0, xStatefulSetInformers.WithTransform(trim))
	return &ControllerContext{
		KubeInformerFactory:         kubeSharedInformers,
		XStatefulsetInformerFactory: xStafulsetInformer,
		InformersStarted:            make(chan struct{}),
	}
}

// ResyncPeriod returns a function which generates a duration each time it is
// invoked; this is because that multiple controllers don't get into lock-step.
func ResyncPeriod(s time.Duration) func() time.Duration {
	return func() time.Duration {
		factor := rand.Float64() + 1
		return time.Duration(float64(s.Nanoseconds()) * factor)
	}
}
