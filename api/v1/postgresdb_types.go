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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PostgresDBSpec defines the desired state of PostgresDB
type PostgresDBSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PostgresDB. Edit postgresdb_types.go to remove/update
	ParentDB  string `json:"parentDb"`
	Secret    string `json:"secret"`
	OwnerRole string `json:"ownerRole"`
}

// PostgresDBStatus defines the observed state of PostgresDB
type PostgresDBStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Created   string       `json:"created"`
	LastCheck *metav1.Time `json:"lastTry,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// PostgresDB is the Schema for the postgresdbs API
type PostgresDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgresDBSpec   `json:"spec,omitempty"`
	Status PostgresDBStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PostgresDBList contains a list of PostgresDB
type PostgresDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgresDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PostgresDB{}, &PostgresDBList{})
}
