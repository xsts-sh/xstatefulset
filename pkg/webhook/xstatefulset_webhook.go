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

	xstsappv1 "github.com/xsts-sh/xstatefulset/api/apps/v1"
	"github.com/xsts-sh/xstatefulset/pkg/controller/legacyscheme"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-apps-x-k8s-io-v1-xstatefulset,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.x-k8s.io,resources=xstatefulsets,verbs=create;update,versions=v1,name=mxstatefulset.kb.io,admissionReviewVersions=v1

// XStatefulSetDefaulter implements a defaulting webhook for XStatefulSet
type XStatefulSetDefaulter struct{}

// SetupWebhookWithManager registers the webhook with the manager
func (w *XStatefulSetDefaulter) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&xstsappv1.XStatefulSet{}).
		WithDefaulter(w).
		Complete()
}

// Default implements webhook.Defaulter
func (w *XStatefulSetDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	set, ok := obj.(*xstsappv1.XStatefulSet)
	if !ok {
		return nil
	}

	// Apply defaults using the generated defaulter functions
	legacyscheme.Scheme.Default(set)

	return nil
}

var _ webhook.CustomDefaulter = &XStatefulSetDefaulter{}
