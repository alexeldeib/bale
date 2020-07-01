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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	capzv1alpha3 "sigs.k8s.io/cluster-api-provider-azure/api/v1alpha3"
	capiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	capbkv1alpha3 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1alpha3"
	kcpv1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	"sigs.k8s.io/cluster-api/util/secret"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infrav1alpha1 "github.com/alexeldeib/bale/api/v1alpha1"
	"github.com/alexeldeib/bale/pkg/remote"
)

// TurtleReconciler reconciles a Turtle object
type TurtleReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	AzureSettings map[string]string
}

// +kubebuilder:rbac:groups=infra.alexeldeib.xyz,resources=turtles,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infra.alexeldeib.xyz,resources=turtles/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=controlplane.cluster.x-k8s.io,resources=kubeadmcontrolplanes;kubeadmcontrolplanes/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters;clusters/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=azureclusters;azuremachinetemplates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=azureclusters/status;azuremachinetemplates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=bootstrap.cluster.x-k8s.io,resources=kubeadmconfigs;kubeadmconfigs/status;kubeadmconfigtemplates;kubeadmconfigtemplates/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=machinedeployments;machinedeployments/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;patch

// kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io;bootstrap.cluster.x-k8s.io;controlplane.cluster.x-k8s.io,resources=*,verbs=get;list;watch;create;update;patch;delete

func (r *TurtleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1alpha1.Turtle{}).
		Owns(&capiv1alpha3.Cluster{}).
		Owns(&kcpv1alpha3.KubeadmControlPlane{}).
		Owns(&capzv1alpha3.AzureCluster{}).
		Owns(&capbkv1alpha3.KubeadmConfigTemplate{}).
		Owns(&capiv1alpha3.MachineDeployment{}).
		Owns(&capzv1alpha3.AzureMachineTemplate{}).
		Complete(r)
}

func (r *TurtleReconciler) Reconcile(req ctrl.Request) (_ ctrl.Result, reterr error) {
	ctx := context.Background()
	log := r.Log.WithValues("turtle", req.NamespacedName)

	var turtle infrav1alpha1.Turtle
	if err := r.Get(ctx, req.NamespacedName, &turtle); err != nil {
		log.Error(err, "unable to fetch")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	reconcilers := []func(context.Context, *infrav1alpha1.Turtle) error{
		r.reconcileCluster,
		r.reconcileKubeadmConfigTemplate,
		r.reconcileKubeadmControlPlane,
		r.reconcileMachineTemplates,
		r.reconcileMachineDeployments,
		r.reconcileAzureCluster,
		r.reconcileExternal,
	}

	for _, reconcileFn := range reconcilers {
		reconcileFn := reconcileFn
		if err := reconcileFn(ctx, &turtle); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to execute reconcile function: %w", err)
		}
	}

	defer func() {
		if err := r.Status().Update(ctx, &turtle); err != nil && reterr == nil {
			log.Error(err, "failed to update turtle status")
			reterr = err
		}
	}()

	return ctrl.Result{}, nil
}

func (r *TurtleReconciler) reconcileCluster(ctx context.Context, turtle *infrav1alpha1.Turtle) error {
	template := getCluster(turtle.Namespace, turtle.Name, turtle.Spec.Location)

	// TODO(ace): Verify -- I believe this is necessary because CreateOrUpdate does a get
	// into the object it receives, so we need to save a copy and capture it
	// into the closure context.
	want := template.DeepCopy()

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, template, func() error {
		if err := controllerutil.SetControllerReference(turtle, want, r.Scheme); err != nil {
			return err
		}
		template = want
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create/update cluster: %w", err)
	}

	return nil
}

func (r *TurtleReconciler) reconcileKubeadmControlPlane(ctx context.Context, turtle *infrav1alpha1.Turtle) error {
	template, err := getKubeadmControlPlane(
		turtle.Namespace,
		turtle.Name,
		turtle.Spec.Location,
		turtle.Spec.Version,
		turtle.Spec.ControlPlaneReplicas,
		r.AzureSettings,
	)

	if err != nil {
		return fmt.Errorf("failed to get azure settings: %w", err)
	}

	// TODO(ace): Verify -- I believe this is necessary because CreateOrUpdate does a get
	// into the object it receives, so we need to save a copy and capture it
	// into the closure context.
	want := template.DeepCopy()

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, template, func() error {
		template = want
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create/update kubeadm control plane: %w", err)
	}

	return nil
}

func (r *TurtleReconciler) reconcileKubeadmConfigTemplate(ctx context.Context, turtle *infrav1alpha1.Turtle) error {
	template, err := getKubeadmConfigTemplate(turtle.Namespace, turtle.Name, turtle.Spec.Location, r.AzureSettings)
	if err != nil {
		return fmt.Errorf("failed to get azure settings: %w", err)
	}

	// TODO(ace): Verify -- I believe this is necessary because CreateOrUpdate does a get
	// into the object it receives, so we need to save a copy and capture it
	// into the closure context.
	want := template.DeepCopy()

	_, err = controllerutil.CreateOrUpdate(ctx, r.Client, template, func() error {
		template = want
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create/update kubeadm config template: %w", err)
	}

	return nil
}

func (r *TurtleReconciler) reconcileMachineTemplates(ctx context.Context, turtle *infrav1alpha1.Turtle) error {
	for _, hatchling := range turtle.Spec.Hatchlings {
		template := getMachineTemplate(turtle.Namespace, hatchling.Name, turtle.Spec.Location, hatchling.VMSize, hatchling.OSDiskSizeGB)
		// TODO(ace): Verify -- I believe this is necessary because CreateOrUpdate does a get
		// into the object it receives, so we need to save a copy and capture it
		// into the closure context.
		want := template.DeepCopy()

		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, template, func() error {
			template = want
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to create/update machine template: %w", err)
		}
	}

	return nil
}

func (r *TurtleReconciler) reconcileMachineDeployments(ctx context.Context, turtle *infrav1alpha1.Turtle) error {
	for _, hatchling := range turtle.Spec.Hatchlings {
		template := getMachineDeployment(turtle.Namespace, hatchling.Name, hatchling.Version, hatchling.Replicas)

		// TODO(ace): Verify -- I believe this is necessary because CreateOrUpdate does a get
		// into the object it receives, so we need to save a copy and capture it
		// into the closure context.
		want := template.DeepCopy()

		_, err := controllerutil.CreateOrUpdate(ctx, r.Client, template, func() error {
			template = want
			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to create/update machine deployment: %w", err)
		}
	}

	return nil
}

func (r *TurtleReconciler) reconcileAzureCluster(ctx context.Context, turtle *infrav1alpha1.Turtle) error {
	template := getAzureCluster(turtle.Namespace, turtle.Name, turtle.Spec.Location)

	// TODO(ace): Verify -- I believe this is necessary because CreateOrUpdate does a get
	// into the object it receives, so we need to save a copy and capture it
	// into the closure context.
	want := template.DeepCopy()

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, template, func() error {
		template = want
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to create/update azure cluster: %w", err)
	}

	return nil
}

func (r *TurtleReconciler) reconcileExternal(ctx context.Context, turtle *infrav1alpha1.Turtle) error {
	// TODO(ace): don't hardcode, specify as reference in object
	azureSecret := &corev1.Secret{}
	azureKey := types.NamespacedName{
		Name:      "bale-manager-credentials",
		Namespace: "bale-system",
	}

	// Fetch azure manager credentials to transfer to remote cluster
	if err := r.Get(ctx, azureKey, azureSecret); err != nil {
		return fmt.Errorf("failed to get azure manager secret to apply to cluster: %w", err)
	}

	// Fetch remove kubeconfig
	kubeconfigSecret := &corev1.Secret{}
	kubeconfigKey := types.NamespacedName{
		Name:      fmt.Sprintf("%s-kubeconfig", turtle.Name),
		Namespace: turtle.Namespace,
	}

	if err := r.Get(ctx, kubeconfigKey, kubeconfigSecret); err != nil {
		return fmt.Errorf("failed to get remote kubeconfig to apply to cluster: %w", err)
	}

	data, ok := kubeconfigSecret.Data[secret.KubeconfigDataName]
	if !ok {
		return fmt.Errorf("missing key %q in secret data", secret.KubeconfigDataName)
	}

	// Construct a kubeclient with it
	remoteClient, err := remote.NewClient(data)
	if err != nil {
		return fmt.Errorf("failed to create REST configuration for turtle %s/%s : %w", turtle.Namespace, turtle.Name, err)
	}

	// Ensure existence of remote namespace
	remoteNamespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: azureKey.Namespace,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, remoteClient, remoteNamespace, func() error {
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create remote azure manager namespace")
	}

	// Create fresh copy to avoid copying stuff like UID, resourceVersion
	remoteSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      azureKey.Name,
			Namespace: azureKey.Namespace,
		},
		Data: azureSecret.Data,
	}
	want := remoteSecret.DeepCopy()
	_, err = controllerutil.CreateOrUpdate(ctx, remoteClient, remoteSecret, func() error {
		remoteSecret = want
		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create remote azure manager secret")
	}

	_, _, err = remoteClient.Apply("https://raw.githubusercontent.com/kubernetes-sigs/cluster-api-provider-azure/master/templates/addons/calico.yaml")

	if err != nil {
		return fmt.Errorf("failed to apply calico config: %w", err)
	}

	return nil
}
