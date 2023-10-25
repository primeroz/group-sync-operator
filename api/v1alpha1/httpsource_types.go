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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HttpSourceSpec defines the desired state of HttpSource
type HttpSourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Source URL where to fetch the file from
	// +operator-sdk:csv:customresourcedefinitions:type=spec
  // +kubebuilder:validation:Required
	SourceUrl string `json:"sourceUrl"`

	// Format of the file
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +kubebuilder:validation:Enum=plaintext;json
  // +kubebuilder:validation:Required
	Format string `json:"format"`

	// Validation Regular Expression to validate each entry before syncing as group subject
	// +operator-sdk:csv:customresourcedefinitions:type=spec
  // +kubebuilder:validation:Required
	ValidationRegex string `json:"validationRegex"`

	// List of Transformers to apply to each element
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Transformers []Transformer `json:"transformers,omitempty"`
}

type Transformer struct {
	// Type of Transformer
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +kubebuilder:validation:Enum=prefix;suffix;regexReplace;regexRemove;regexKeep;camelCase;jsonPathExtract
  // +kubebuilder:validation:Required
	Type string `json:"type"`

	// Value of Transformer
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Value string `json:"value,omnitempty"`
}


// HttpSourceStatus defines the observed state of HttpSource
type HttpSourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Conditions store the status conditions of the HttpSource Syncs
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// HttpSource is the Schema for the httpsources API
type HttpSource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HttpSourceSpec   `json:"spec,omitempty"`
	Status HttpSourceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HttpSourceList contains a list of HttpSource
type HttpSourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HttpSource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HttpSource{}, &HttpSourceList{})
}
