/*
Copyright 2023.

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

package controllers

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	groupsyncv1alpha1 "github.com/primeroz/group-sync-operator/api/v1alpha1"
)

// HttpSourceReconciler reconciles a HttpSource object
type HttpSourceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=groupsync.primeroz.xyz,resources=httpsources,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=groupsync.primeroz.xyz,resources=httpsources/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=groupsync.primeroz.xyz,resources=httpsources/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the HttpSource object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *HttpSourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	httpSource := &groupsyncv1alpha1.HttpSource{}
	err := r.Get(ctx, req.NamespacedName, httpSource)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("HttpSource resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get HttpSource")

		return ctrl.Result{}, err
	}

  // Fetch the file from the remote Url
  resp, err := http.Get(httpSource.Spec.SourceUrl)
	defer resp.Body.Close()

	if err != nil {

		meta.SetStatusCondition( 
      &httpSource.Status.Conditions,  
      metav1.Condition{   
        Type: "Failed", 
        Status: metav1.ConditionTrue,    
        Reason: "Fetching", 
        Message: fmt.Sprintf("Failed to fetch file from %s", httpSource.Spec.SourceUrl)})

		return ctrl.Result{}, err
	}

	meta.SetStatusCondition( 
    &httpSource.Status.Conditions,  
    metav1.Condition{   
      Type: "Failed", 
      Status: metav1.ConditionFalse,    
      Reason: "Fetching", 
      Message: fmt.Sprintf("successfully fetched file from %s", httpSource.Spec.SourceUrl)})

		if err := r.Status().Update(ctx, httpSource); err != nil {
			log.Error(err, "Failed to update HttpSource status")
			return ctrl.Result{}, err
		}

		// // The following implementation will raise an event
		// r.Recorder.Event(cr, "Warning", "Deleting",
		// 	fmt.Sprintf("Custom Resource %s is being deleted from the namespace %s",
		// 		cr.Name,
		// 		cr.Namespace))

	log.Info("HttpSource: ")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HttpSourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&groupsyncv1alpha1.HttpSource{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}
