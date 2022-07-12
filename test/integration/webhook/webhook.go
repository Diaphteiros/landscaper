// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"context"
	"path/filepath"

	lsv1alpha1helper "github.com/gardener/landscaper/apis/core/v1alpha1/helper"
	utils2 "github.com/gardener/landscaper/pkg/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/gardener/landscaper/apis/config"
	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	hdv1alpha1 "github.com/gardener/landscaper/apis/deployer/helm/v1alpha1"
	kutil "github.com/gardener/landscaper/controller-utils/pkg/kubernetes"
	"github.com/gardener/landscaper/test/framework"
	"github.com/gardener/landscaper/test/utils"
)

// RegisterTests registers all tests of the package
func RegisterTests(f *framework.Framework) {
	WebhookTest(f)
}

func WebhookTest(f *framework.Framework) {
	_ = Describe("WebhookTest", func() {
		var (
			ctx   context.Context
			state = f.Register()
		)

		BeforeEach(func() {
			ctx = context.Background()
		})

		AfterEach(func() {
			ctx.Done()
		})

		It("should have created a ValidatingWebhookConfiguration", func() {
			vwc := admissionregistrationv1.ValidatingWebhookConfiguration{}
			utils.ExpectNoError(f.Client.Get(ctx, kutil.ObjectKey("landscaper-validation-webhook", ""), &vwc))
		})

		It("should block invalid Installation resources", func() {
			instResource := filepath.Join(f.RootPath, "/docs/tutorials/resources/ingress-nginx", "installation.yaml")

			// load nginx installation from tutorial
			inst := &lsv1alpha1.Installation{}
			Expect(utils.ReadResourceFromFile(inst, instResource)).To(Succeed())
			inst.SetNamespace(state.Namespace)
			if utils2.IsNewReconcile() {
				lsv1alpha1helper.SetOperation(&inst.ObjectMeta, lsv1alpha1.ReconcileOperation)
			}

			// make installation invalid by duplicating the first export
			inst.Spec.Exports.Data = append(inst.Spec.Exports.Data, inst.Spec.Exports.Data[0])
			err := state.Create(ctx, inst)
			Expect(err).To(HaveOccurred()) // validation webhook should have denied this
			Expect(err.Error()).To(HavePrefix("admission webhook \"installations.validation.landscaper.gardener.cloud\" denied the request"))
		})

		It("should block invalid Execution resources", func() {
			conf := hdv1alpha1.Configuration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ProviderConfiguration",
					APIVersion: "helm.deployer.landscaper.gardener.cloud/v1alpha1",
				},
				OCI: &config.OCIConfiguration{
					AllowPlainHttp: false,
				},
				TargetSelector: []lsv1alpha1.TargetSelector{},
			}
			obj := conf.DeepCopyObject()
			raw := runtime.RawExtension{}
			utils.ExpectNoError(runtime.Convert_runtime_Object_To_runtime_RawExtension(&obj, &raw, nil))

			// create invalid execution (cyclic dependencies)
			exec := &lsv1alpha1.Execution{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-execution",
					Namespace: state.Namespace,
				},
				Spec: lsv1alpha1.ExecutionSpec{
					DeployItems: lsv1alpha1.DeployItemTemplateList{
						{
							Name:          "a",
							Type:          "landscaper.gardener.cloud/helm",
							DependsOn:     []string{"b"},
							Configuration: &raw,
						},
						{
							Name:          "b",
							Type:          "landscaper.gardener.cloud/helm",
							DependsOn:     []string{"a"},
							Configuration: &raw,
						},
					},
				},
			}

			err := state.Create(ctx, exec)
			Expect(err).To(HaveOccurred()) // validation webhook should have denied this
			Expect(err.Error()).To(HavePrefix("admission webhook \"executions.validation.landscaper.gardener.cloud\" denied the request"))
		})

		It("should block invalid DeployItem resources", func() {
			conf := hdv1alpha1.Configuration{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ProviderConfiguration",
					APIVersion: "helm.deployer.landscaper.gardener.cloud/v1alpha1",
				},
				OCI: &config.OCIConfiguration{
					AllowPlainHttp: false,
				},
				TargetSelector: []lsv1alpha1.TargetSelector{},
			}
			obj := conf.DeepCopyObject()
			raw := runtime.RawExtension{}
			utils.ExpectNoError(runtime.Convert_runtime_Object_To_runtime_RawExtension(&obj, &raw, nil))

			di := &lsv1alpha1.DeployItem{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployitem",
					Namespace: state.Namespace,
				},
				Spec: lsv1alpha1.DeployItemSpec{
					Type: "landscaper.gardener.cloud/helm",
					Target: &lsv1alpha1.ObjectReference{
						Name:      "", // invalid
						Namespace: state.Namespace,
					},
					Configuration: &raw,
				},
			}

			err := state.Create(ctx, di)
			Expect(err).To(HaveOccurred()) // validation webhook should have denied this
			Expect(err.Error()).To(HavePrefix("admission webhook \"deployitems.validation.landscaper.gardener.cloud\" denied the request"))
		})

		It("should block a DeployItem type update", func() {
			di := &lsv1alpha1.DeployItem{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-deployitem",
					Namespace: state.Namespace,
				},
				Spec: lsv1alpha1.DeployItemSpec{
					Type: "some-type",
				},
			}

			err := state.Create(ctx, di)
			Expect(err).ToNot(HaveOccurred())

			updated := di.DeepCopy()
			updated.Spec.Type = "other-type"
			err = f.Client.Update(ctx, updated)
			Expect(err).To(HaveOccurred()) // validation webhook should have denied this
			Expect(err.Error()).To(HavePrefix("admission webhook \"deployitems.validation.landscaper.gardener.cloud\" denied the request"))
		})
	})
}
