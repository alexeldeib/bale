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

// HatchlingSpec defines the desired state of Hatchling
type HatchlingSpec struct {
	Name string `json:"name"`
	// +kubebuilder:default=512
	OSDiskSizeGB int32 `json:"osDiskSizeGB,omitempty"`
	// +kubebuilder:default=1
	Replicas int32  `json:"replicas,omitempty"`
	Version  string `json:"version,omitempty"`
	// +kubebuilder:default=Standard_D8s_v3
	VMSize string `json:"vmSize,omitempty"`
}

// HatchlingStatus defines the observed state of Hatchling
type HatchlingStatus struct{}

// // kubebuilder:object:root=true

// // Hatchling is the Schema for the hatchlings API
// type Hatchling struct {
// 	metav1.TypeMeta   `json:",inline"`
// 	metav1.ObjectMeta `json:"metadata,omitempty"`

// 	Spec   HatchlingSpec   `json:"spec,omitempty"`
// 	Status HatchlingStatus `json:"status,omitempty"`
// }

// // kubebuilder:object:root=true

// // HatchlingList contains a list of Hatchling
// type HatchlingList struct {
// 	metav1.TypeMeta `json:",inline"`
// 	metav1.ListMeta `json:"metadata,omitempty"`
// 	Items           []Hatchling `json:"items"`
// }

// func init() {
// 	SchemeBuilder.Register(&Hatchling{}, &HatchlingList{})
// }
