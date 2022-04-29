//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright (c) 2021 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file

SPDX-License-Identifier: Apache-2.0
*/
// Code generated by conversion-gen. DO NOT EDIT.

package v1alpha1

import (
	unsafe "unsafe"

	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"

	corev1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	manifest "github.com/gardener/landscaper/apis/deployer/manifest"
)

func init() {
	localSchemeBuilder.Register(RegisterConversions)
}

// RegisterConversions adds conversion functions to the given scheme.
// Public to allow building arbitrary schemes.
func RegisterConversions(s *runtime.Scheme) error {
	if err := s.AddGeneratedConversionFunc((*Configuration)(nil), (*manifest.Configuration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Configuration_To_manifest_Configuration(a.(*Configuration), b.(*manifest.Configuration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*manifest.Configuration)(nil), (*Configuration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_manifest_Configuration_To_v1alpha1_Configuration(a.(*manifest.Configuration), b.(*Configuration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*Controller)(nil), (*manifest.Controller)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_Controller_To_manifest_Controller(a.(*Controller), b.(*manifest.Controller), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*manifest.Controller)(nil), (*Controller)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_manifest_Controller_To_v1alpha1_Controller(a.(*manifest.Controller), b.(*Controller), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*ExportConfiguration)(nil), (*manifest.ExportConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ExportConfiguration_To_manifest_ExportConfiguration(a.(*ExportConfiguration), b.(*manifest.ExportConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddGeneratedConversionFunc((*manifest.ExportConfiguration)(nil), (*ExportConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_manifest_ExportConfiguration_To_v1alpha1_ExportConfiguration(a.(*manifest.ExportConfiguration), b.(*ExportConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*manifest.ProviderConfiguration)(nil), (*ProviderConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_manifest_ProviderConfiguration_To_v1alpha1_ProviderConfiguration(a.(*manifest.ProviderConfiguration), b.(*ProviderConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*manifest.ProviderStatus)(nil), (*ProviderStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_manifest_ProviderStatus_To_v1alpha1_ProviderStatus(a.(*manifest.ProviderStatus), b.(*ProviderStatus), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*ProviderConfiguration)(nil), (*manifest.ProviderConfiguration)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ProviderConfiguration_To_manifest_ProviderConfiguration(a.(*ProviderConfiguration), b.(*manifest.ProviderConfiguration), scope)
	}); err != nil {
		return err
	}
	if err := s.AddConversionFunc((*ProviderStatus)(nil), (*manifest.ProviderStatus)(nil), func(a, b interface{}, scope conversion.Scope) error {
		return Convert_v1alpha1_ProviderStatus_To_manifest_ProviderStatus(a.(*ProviderStatus), b.(*manifest.ProviderStatus), scope)
	}); err != nil {
		return err
	}
	return nil
}

func autoConvert_v1alpha1_Configuration_To_manifest_Configuration(in *Configuration, out *manifest.Configuration, s conversion.Scope) error {
	out.Identity = in.Identity
	out.TargetSelector = *(*[]corev1alpha1.TargetSelector)(unsafe.Pointer(&in.TargetSelector))
	if err := Convert_v1alpha1_ExportConfiguration_To_manifest_ExportConfiguration(&in.Export, &out.Export, s); err != nil {
		return err
	}
	if err := Convert_v1alpha1_Controller_To_manifest_Controller(&in.Controller, &out.Controller, s); err != nil {
		return err
	}
	return nil
}

// Convert_v1alpha1_Configuration_To_manifest_Configuration is an autogenerated conversion function.
func Convert_v1alpha1_Configuration_To_manifest_Configuration(in *Configuration, out *manifest.Configuration, s conversion.Scope) error {
	return autoConvert_v1alpha1_Configuration_To_manifest_Configuration(in, out, s)
}

func autoConvert_manifest_Configuration_To_v1alpha1_Configuration(in *manifest.Configuration, out *Configuration, s conversion.Scope) error {
	out.Identity = in.Identity
	out.TargetSelector = *(*[]corev1alpha1.TargetSelector)(unsafe.Pointer(&in.TargetSelector))
	if err := Convert_manifest_ExportConfiguration_To_v1alpha1_ExportConfiguration(&in.Export, &out.Export, s); err != nil {
		return err
	}
	if err := Convert_manifest_Controller_To_v1alpha1_Controller(&in.Controller, &out.Controller, s); err != nil {
		return err
	}
	return nil
}

// Convert_manifest_Configuration_To_v1alpha1_Configuration is an autogenerated conversion function.
func Convert_manifest_Configuration_To_v1alpha1_Configuration(in *manifest.Configuration, out *Configuration, s conversion.Scope) error {
	return autoConvert_manifest_Configuration_To_v1alpha1_Configuration(in, out, s)
}

func autoConvert_v1alpha1_Controller_To_manifest_Controller(in *Controller, out *manifest.Controller, s conversion.Scope) error {
	out.CommonControllerConfig = in.CommonControllerConfig
	return nil
}

// Convert_v1alpha1_Controller_To_manifest_Controller is an autogenerated conversion function.
func Convert_v1alpha1_Controller_To_manifest_Controller(in *Controller, out *manifest.Controller, s conversion.Scope) error {
	return autoConvert_v1alpha1_Controller_To_manifest_Controller(in, out, s)
}

func autoConvert_manifest_Controller_To_v1alpha1_Controller(in *manifest.Controller, out *Controller, s conversion.Scope) error {
	out.CommonControllerConfig = in.CommonControllerConfig
	return nil
}

// Convert_manifest_Controller_To_v1alpha1_Controller is an autogenerated conversion function.
func Convert_manifest_Controller_To_v1alpha1_Controller(in *manifest.Controller, out *Controller, s conversion.Scope) error {
	return autoConvert_manifest_Controller_To_v1alpha1_Controller(in, out, s)
}

func autoConvert_v1alpha1_ExportConfiguration_To_manifest_ExportConfiguration(in *ExportConfiguration, out *manifest.ExportConfiguration, s conversion.Scope) error {
	out.DefaultTimeout = (*corev1alpha1.Duration)(unsafe.Pointer(in.DefaultTimeout))
	return nil
}

// Convert_v1alpha1_ExportConfiguration_To_manifest_ExportConfiguration is an autogenerated conversion function.
func Convert_v1alpha1_ExportConfiguration_To_manifest_ExportConfiguration(in *ExportConfiguration, out *manifest.ExportConfiguration, s conversion.Scope) error {
	return autoConvert_v1alpha1_ExportConfiguration_To_manifest_ExportConfiguration(in, out, s)
}

func autoConvert_manifest_ExportConfiguration_To_v1alpha1_ExportConfiguration(in *manifest.ExportConfiguration, out *ExportConfiguration, s conversion.Scope) error {
	out.DefaultTimeout = (*corev1alpha1.Duration)(unsafe.Pointer(in.DefaultTimeout))
	return nil
}

// Convert_manifest_ExportConfiguration_To_v1alpha1_ExportConfiguration is an autogenerated conversion function.
func Convert_manifest_ExportConfiguration_To_v1alpha1_ExportConfiguration(in *manifest.ExportConfiguration, out *ExportConfiguration, s conversion.Scope) error {
	return autoConvert_manifest_ExportConfiguration_To_v1alpha1_ExportConfiguration(in, out, s)
}
