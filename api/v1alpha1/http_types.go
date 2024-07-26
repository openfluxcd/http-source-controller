/*
Copyright 2024.

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

// HttpSpec defines the desired state of Http
type HttpSpec struct {
	// URL defines where to get the archive from.
	URL string `json:"url"`
}

// HttpStatus defines the observed state of Http
type HttpStatus struct {
	// ObservedGeneration is the last reconciled generation.
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// The last successfully applied revision.
	// Equals the Revision of the applied Artifact from the referenced Source.
	// +optional
	LastAppliedRevision string `json:"lastAppliedRevision,omitempty"`

	// LastAttemptedRevision is the revision of the last reconciliation attempt.
	// +optional
	LastAttemptedRevision string `json:"lastAttemptedRevision,omitempty"`

	// ArtifactName present what the name of the generated artifact is.
	ArtifactName string `json:"artifactName,omitempty"`
}

// GetConditions returns the status conditions of the object.
func (in Http) GetConditions() []metav1.Condition {
	return in.Status.Conditions
}

// SetConditions sets the status conditions on the object.
func (in *Http) SetConditions(conditions []metav1.Condition) {
	in.Status.Conditions = conditions
}

func (in *Http) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *Http) GetKind() string {
	return "Http"
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Http is the Schema for the https API
type Http struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HttpSpec   `json:"spec,omitempty"`
	Status HttpStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// HttpList contains a list of Http
type HttpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Http `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Http{}, &HttpList{})
}
