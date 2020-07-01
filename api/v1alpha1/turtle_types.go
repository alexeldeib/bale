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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TurtleSpec defines the desired state of Turtle
type TurtleSpec struct {
	// +kubebuilder:default=1
	ControlPlaneReplicas int32           `json:"controlPlaneReplicas,omitempty"`
	Location             string          `json:"location"`
	ResourceGroup        string          `json:"resourceGroup,omitempty"`
	Hatchlings           []HatchlingSpec `json:"hatchlings,omitempty"`
	// Version is the Kubernetes version of the control plane.
	Version string `json:"version"`
}

// TurtleStatus defines the observed state of Turtle
type TurtleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Turtle is the Schema for the turtles API
type Turtle struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TurtleSpec   `json:"spec,omitempty"`
	Status TurtleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// TurtleList contains a list of Turtle
type TurtleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Turtle `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Turtle{}, &TurtleList{})
}
