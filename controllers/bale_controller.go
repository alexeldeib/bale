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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrav1alpha1 "github.com/alexeldeib/bale/api/v1alpha1"
)

// BaleReconciler reconciles a Bale object
type BaleReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *BaleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1alpha1.Bale{}).
		Owns(&infrav1alpha1.Turtle{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=infra.alexeldeib.xyz,resources=bales,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.alexeldeib.xyz,resources=bales/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infra.alexeldeib.xyz,resources=turtles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.alexeldeib.xyz,resources=turtles/status,verbs=get;update;patch

func (r *BaleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("bale", req.NamespacedName)

	var bale infrav1alpha1.Bale
	if err := r.Get(ctx, req.NamespacedName, &bale); err != nil {
		log.Error(err, "unable to fetch")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	selector, err := metav1.LabelSelectorAsSelector(bale.Spec.Selector)
	if err != nil {
		log.Error(err, "failed to convert selector")
		return ctrl.Result{}, err

	}

	var armada infrav1alpha1.TurtleList
	if err := r.List(context.Background(), &armada, client.MatchingLabelsSelector{Selector: selector}); err != nil {
		log.Error(err, "unable to fetch bale list")
		return ctrl.Result{}, err
	}

	diff := bale.Spec.Replicas - int32(len(armada.Items))
	if diff <= 0 {
		log.Info(fmt.Sprintf("found %d replicas, required %d, diff of %d. returning early", len(armada.Items), bale.Spec.Replicas, diff))
	}

	for i := 0; int32(i) < diff; i++ {
		name := "acecap-" + RandomLowercaseString(6)
		turtle := new(infrav1alpha1.Turtle)
		turtle.Namespace = bale.Namespace
		turtle.Name = name
		turtle.Spec = *bale.Spec.Template.DeepCopy()
		turtle.Spec.ResourceGroup = name
		want := turtle.DeepCopy()

		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, turtle, func() error {
			if err := controllerutil.SetControllerReference(turtle, want, r.Scheme); err != nil {
				return err
			}
			turtle = want
			return nil
		})

		if err != nil {
			return ctrl.Result{}, errors.Wrap(err, "failed to create/update turtle: %w")
		}
	}

	return ctrl.Result{}, nil
}
