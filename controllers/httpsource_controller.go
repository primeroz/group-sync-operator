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
	"io"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	groupsyncv1alpha1 "github.com/primeroz/group-sync-operator/api/v1alpha1"
	"github.com/primeroz/group-sync-operator/pkg/format"
	"github.com/primeroz/group-sync-operator/pkg/transformer"
	"github.com/primeroz/group-sync-operator/pkg/validation"
)

// HttpSourceReconciler reconciles a HttpSource object
type HttpSourceReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

func getFileFromHTTP(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// TODO: ADd extra rbac here for group sybjects management
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

	log.V(1).Info("Syncing HttpSource")

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
	body, err := getFileFromHTTP(httpSource.Spec.SourceUrl)

	if err != nil {
		log.Error(err, "Failed to fetch file")

		meta.SetStatusCondition(
			&httpSource.Status.Conditions,
			metav1.Condition{
				Type:               "Failed",
				Status:             metav1.ConditionTrue,
				Reason:             "Fetching",
				LastTransitionTime: metav1.Now(),
				Message:            fmt.Sprintf("Failed to fetch file from %s", httpSource.Spec.SourceUrl)})

		if err := r.Status().Update(ctx, httpSource); err != nil {
			log.Error(err, "Failed to update HttpSource status")
		}

		// Wait 1 minute before requeing
		// exit without error so the requeue after works
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	// Parse Users
	// TODO: Should have a plugin interaface here
	var users []string

	if httpSource.Spec.Format == "plaintext" {
		users, err = format.ParseUsersFromPlaintext(body)
	}

	if err != nil {
		log.Error(err, "Failed to Parse Users")

		meta.SetStatusCondition(
			&httpSource.Status.Conditions,
			metav1.Condition{
				Type:               "Failed",
				Status:             metav1.ConditionTrue,
				Reason:             "Parsing",
				LastTransitionTime: metav1.Now(),
				Message:            "Failed to Parse Users from fetched file"})

		if err := r.Status().Update(ctx, httpSource); err != nil {
			log.Error(err, "Failed to update HttpSource status")
		}

		// Wait 1 minute before requeing
		// exit without error so the requeue after works
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	// TODO: Should have a plugin interaface here
	// Apply transformations
	if len(httpSource.Spec.Transformers) > 0 {
		for _, t := range httpSource.Spec.Transformers {
			if t.Type == "regexKeep" {
				users, err = transformer.RegexKeep(users, t.Value)
			} else if t.Type == "regexRemove" {
				users, err = transformer.RegexRemove(users, t.Value)
			} else if t.Type == "prefix" {
				users, err = transformer.Prefix(users, t.Value)
			} else if t.Type == "suffix" {
				users, err = transformer.Suffix(users, t.Value)
			}

			if err != nil {
				log.Error(err, "Failed to Apply Transformer")

				meta.SetStatusCondition(
					&httpSource.Status.Conditions,
					metav1.Condition{
						Type:               "Failed",
						Status:             metav1.ConditionTrue,
						Reason:             "Transformer",
						LastTransitionTime: metav1.Now(),
						Message:            "Failed to apply transformer"})

				if err := r.Status().Update(ctx, httpSource); err != nil {
					log.Error(err, "Failed to update Transformer")
				}

				// Wait 1 minute before requeing
				// exit without error so the requeue after works
				return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
			}
		}

	}

	// Validate - All elements must match
	err = validation.ValidateUsersRegex(users, httpSource.Spec.ValidationRegex)
	// XXX: Should we apply a 0 list of subjects ?

	if err != nil {
		log.Error(err, "Failed to Validate Users")

		meta.SetStatusCondition(
			&httpSource.Status.Conditions,
			metav1.Condition{
				Type:               "Failed",
				Status:             metav1.ConditionTrue,
				Reason:             "Validate",
				LastTransitionTime: metav1.Now(),
				Message:            "Failed to Validate Users against regex"})

		if err := r.Status().Update(ctx, httpSource); err != nil {
			log.Error(err, "Failed to update HttpSource status")
		}

		// Wait 1 minute before requeing
		// exit without error so the requeue after works
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	// Update Group

	for _, u := range users {
		fmt.Println("DEBUG", httpSource.Name, "USER : ", u)
	}

	// XXX: How to force the status to always update LastTransitionTime ?
	meta.SetStatusCondition(
		&httpSource.Status.Conditions,
		metav1.Condition{
			Type:               "Failed",
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             "Success",
			Message:            "Successfully Synced Group"})

	if err := r.Status().Update(ctx, httpSource); err != nil {
		log.Error(err, "Failed to update HttpSource status")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HttpSourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&groupsyncv1alpha1.HttpSource{}).
		WithOptions(controller.Options{MaxConcurrentReconciles: 1}).
		Complete(r)
}
