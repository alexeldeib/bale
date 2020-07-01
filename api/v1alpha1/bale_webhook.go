// Copyright 2020 Alexander Eldeib
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS
// OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
// IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
// CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
// TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
// SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package v1alpha1

import (
	"fmt"

	"github.com/blang/semver"
	apierr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var balelog = logf.Log.WithName("bale-resource")

func (r *Bale) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-infra-alexeldeib-xyz-v1alpha1-bale,mutating=true,failurePolicy=fail,groups=infra.alexeldeib.xyz,resources=bales,verbs=create;update,versions=v1alpha1,name=mbale.kb.io

var _ webhook.Defaulter = &Bale{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Bale) Default() {
	balelog.Info("default", "name", r.Name)

	controlPlaneVersion := r.Spec.Template.Version
	for i := range r.Spec.Template.Hatchlings {
		hatchling := r.Spec.Template.Hatchlings[i]
		if hatchling.Version == "" {
			hatchling.Version = controlPlaneVersion
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-infra-alexeldeib-xyz-v1alpha1-bale,mutating=false,failurePolicy=fail,groups=infra.alexeldeib.xyz,resources=bales,versions=v1alpha1,name=vbale.kb.io

var _ webhook.Validator = &Bale{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Bale) ValidateCreate() error {
	balelog.Info("validate create", "name", r.Name)

	// validate control plane has higher version than all workers
	controlPlaneVersion := r.Spec.Template.Version
	for i := range r.Spec.Template.Hatchlings {
		hatchling := r.Spec.Template.Hatchlings[i]
		if hatchling.Version != "" {
			var cpSemver, hatchlingSemver semver.Version
			var err error

			cpSemver, err = semver.Make(controlPlaneVersion)
			if err != nil {
				return apierr.NewInternalError(err)
			}

			hatchlingSemver, err = semver.Make(hatchling.Version)
			if err != nil {
				return apierr.NewInternalError(err)
			}

			if cpSemver.LT(hatchlingSemver) {
				return apierr.NewBadRequest(
					fmt.Sprintf(
						"control plane versin %s cannot be less than hatchling version %s",
						controlPlaneVersion,
						hatchling.Version,
					),
				)
			}
		}
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Bale) ValidateUpdate(old runtime.Object) error {
	balelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Bale) ValidateDelete() error {
	balelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
