/*
Copyright 2022.

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

// ComponentType defines the components to be installed
type ComponentType struct {
	Name     string            `json:"name"`
	Settings map[string]string `json:"settings"`
}

// KymaSpec defines the desired state of Kyma
type KymaSpec struct {
	Release string `json:"release"`
	// Components specifies the list of components to be installed
	Components []ComponentType `json:"components,omitempty"`
}

// KymaStatus defines the observed state of Kyma
// +kubebuilder:subresource:status
type KymaStatus struct {
	// State signifies current state of Kyma.
	// Value can be one of ("Ready", "Processing", "Error", "Deleting").
	State KymaState `json:"state,omitempty"`

	// List of status conditions to indicate the status of a ServiceInstance.
	// +optional
	Conditions []KymaCondition `json:"conditions,omitempty"`

	// Observed generation
	// +optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Active Release
	// +optional
	ActiveRelease string `json:"activeRelease,omitempty"`
}

type KymaState string

// Valid Kyma States
const (
	// KymaStateReady signifies Kyma is ready
	KymaStateReady KymaState = "Ready"

	// KymaStateProcessing signifies Kyma is reconciling
	KymaStateProcessing KymaState = "Processing"

	// KymaStateError signifies an error for Kyma
	KymaStateError KymaState = "Error"

	// KymaStateDeleting signifies Kyma is being deleted
	KymaStateDeleting KymaState = "Deleting"
)

// KymaCondition describes condition information for Kyma.
type KymaCondition struct {
	Type KymaConditionType `json:"type"`

	// Status of the Kyma Condition.
	// Value can be one of ("True", "False", "Unknown").
	Status KymaConditionStatus `json:"status"`

	// Human-readable message indicating details about the last status transition.
	// +optional
	Message string `json:"message,omitempty"`

	// Machine-readable text indicating the reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Timestamp for when Kyma last transitioned from one status to another.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`
}

type KymaConditionType string

const (
	// ConditionTypeReady represents KymaConditionType Ready
	ConditionTypeReady KymaConditionType = "Ready"
)

type KymaConditionStatus string

// Valid KymaCondition Status
const (
	// ConditionStatusTrue signifies KymaConditionStatus true
	ConditionStatusTrue KymaConditionStatus = "True"

	// ConditionStatusFalse signifies KymaConditionStatus false
	ConditionStatusFalse KymaConditionStatus = "False"

	// ConditionStatusUnknown signifies KymaConditionStatus unknown
	ConditionStatusUnknown KymaConditionStatus = "Unknown"
)

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Kyma is the Schema for the kymas API
type Kyma struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KymaSpec   `json:"spec,omitempty"`
	Status KymaStatus `json:"status,omitempty"`
}

func (k *Kyma) IsAlreadyReconciled() bool {
	return (k.Status.State == KymaStateReady || k.Status.State == KymaStateError) && k.Status.ObservedGeneration == k.ObjectMeta.Generation
}

func (k *Kyma) IsReconciling() bool {
	return k.Status.State == KymaStateProcessing
}

//+kubebuilder:object:root=true

// KymaList contains a list of Kyma
type KymaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kyma `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kyma{}, &KymaList{})
}
