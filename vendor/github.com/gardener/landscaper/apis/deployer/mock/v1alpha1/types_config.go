// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Configuration is the helm deployer configuration that configures the controller
type Configuration struct {
	metav1.TypeMeta `json:",inline"`
	// Identity identity describes the unique identity of the deployer.
	// +optional
	Identity string `json:"identity,omitempty"`
	// TargetSelector describes all selectors the deployer should depend on.
	TargetSelector []lsv1alpha1.TargetSelector `json:"targetSelector,omitempty"`
}
