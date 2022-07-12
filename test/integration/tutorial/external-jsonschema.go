// SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package tutorial

import (
	"context"
	"path/filepath"
	"time"

	lsv1alpha1helper "github.com/gardener/landscaper/apis/core/v1alpha1/helper"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	kutil "github.com/gardener/landscaper/controller-utils/pkg/kubernetes"
	lsutils "github.com/gardener/landscaper/pkg/utils/landscaper"
	"github.com/gardener/landscaper/test/framework"
	"github.com/gardener/landscaper/test/utils"
)

// ExternalJSONSchemaTest tests the jsonschema tutorial.
func ExternalJSONSchemaTest(f *framework.Framework) {
	_ = Describe("ExternalJSONSchemaTest", func() {
		state := f.Register()

		It("should deploy an echo server with resources defined by an external jsonschema", func() {
			var (
				tutorialResourcesRootDir = filepath.Join(f.RootPath, "/docs/tutorials/resources/external-jsonschema")
				targetResource           = filepath.Join(tutorialResourcesRootDir, "my-target.yaml")
				importResource           = filepath.Join(tutorialResourcesRootDir, "configmap.yaml")
				instResource             = filepath.Join(tutorialResourcesRootDir, "installation.yaml")
			)
			ctx := context.Background()
			defer ctx.Done()

			By("Create Target for the installation")
			target := &lsv1alpha1.Target{}
			utils.ExpectNoError(utils.ReadResourceFromFile(target, targetResource))
			target, err := utils.BuildInternalKubernetesTarget(ctx, f.Client, state.Namespace, target.Name, f.RestConfig, true)
			utils.ExpectNoError(err)
			utils.ExpectNoError(state.Create(ctx, target))

			By("Create ConfigMap with imports for the installation")
			cm := &corev1.ConfigMap{}
			utils.ExpectNoError(utils.ReadResourceFromFile(cm, importResource))
			cm.SetNamespace(state.Namespace)
			utils.ExpectNoError(state.Create(ctx, cm))

			By("Create Installation")
			inst := &lsv1alpha1.Installation{}
			Expect(utils.ReadResourceFromFile(inst, instResource)).To(Succeed())
			inst.SetNamespace(state.Namespace)
			utils.ExpectNoError(state.Create(ctx, inst))

			// wait for installation to finish
			utils.ExpectNoError(lsutils.WaitForInstallationToBeHealthy(ctx, f.Client, inst, 2*time.Minute))

			deployItems, err := lsutils.GetDeployItemsOfInstallation(ctx, f.Client, inst)
			utils.ExpectNoError(err)
			Expect(deployItems).To(HaveLen(1))
			Expect(deployItems[0].Status.Phase).To(Equal(lsv1alpha1.ExecutionPhaseSucceeded))

			// todo: make namespace configurable for deployed resources
			// expect that the echo server deployment is successfully running
			echoServerDeploymentName := "echo-server"
			echoServerDeploymentObjectKey := kutil.ObjectKey(echoServerDeploymentName, "default")
			utils.ExpectNoError(utils.WaitForDeploymentToBeReady(ctx, f.TestLog(), f.Client, echoServerDeploymentObjectKey, 2*time.Minute))
			// expect that the deployment has the correct resource requests and limits
			echoServerDeploy := &appsv1.Deployment{}
			utils.ExpectNoError(f.Client.Get(ctx, echoServerDeploymentObjectKey, echoServerDeploy))
			Expect(echoServerDeploy.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(echoServerDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()).To(Equal("50Mi"))
			Expect(echoServerDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()).To(Equal("100Mi"))

			By("Delete installation")
			utils.ExpectNoError(f.Client.Delete(ctx, inst))
			utils.ExpectNoError(utils.WaitForObjectDeletion(ctx, f.Client, inst, 2*time.Minute))

			// expect that the nginx deployment is alread deleted or has an deletion timestamp
			err = f.Client.Get(ctx, echoServerDeploymentObjectKey, echoServerDeploy)
			if err != nil && !apierrors.IsNotFound(err) {
				utils.ExpectNoError(err)
			} else if err == nil {
				Expect(echoServerDeploy.DeletionTimestamp.IsZero()).To(BeTrue())
			}
		})
	})
}

// ExternalJSONSchemaTest tests the jsonschema tutorial.
func ExternalJSONSchemaTestForNewReconcile(f *framework.Framework) {
	_ = Describe("ExternalJSONSchemaTest", func() {
		state := f.Register()

		It("should deploy an echo server with resources defined by an external jsonschema", func() {
			var (
				tutorialResourcesRootDir = filepath.Join(f.RootPath, "/docs/tutorials/resources/external-jsonschema")
				targetResource           = filepath.Join(tutorialResourcesRootDir, "my-target.yaml")
				importResource           = filepath.Join(tutorialResourcesRootDir, "configmap.yaml")
				instResource             = filepath.Join(tutorialResourcesRootDir, "installation.yaml")
			)
			ctx := context.Background()
			defer ctx.Done()

			By("Create Target for the installation")
			target := &lsv1alpha1.Target{}
			utils.ExpectNoError(utils.ReadResourceFromFile(target, targetResource))
			target, err := utils.BuildInternalKubernetesTarget(ctx, f.Client, state.Namespace, target.Name, f.RestConfig, true)
			utils.ExpectNoError(err)
			utils.ExpectNoError(state.Create(ctx, target))

			By("Create ConfigMap with imports for the installation")
			cm := &corev1.ConfigMap{}
			utils.ExpectNoError(utils.ReadResourceFromFile(cm, importResource))
			cm.SetNamespace(state.Namespace)
			utils.ExpectNoError(state.Create(ctx, cm))

			By("Create Installation")
			inst := &lsv1alpha1.Installation{}
			Expect(utils.ReadResourceFromFile(inst, instResource)).To(Succeed())
			inst.SetNamespace(state.Namespace)
			lsv1alpha1helper.SetOperation(&inst.ObjectMeta, lsv1alpha1.ReconcileOperation)
			utils.ExpectNoError(state.Create(ctx, inst))

			// wait for installation to finish
			utils.ExpectNoError(lsutils.WaitForInstallationToFinish(ctx, f.Client, inst, lsv1alpha1.InstallationPhaseSucceeded, 2*time.Minute))

			deployItems, err := lsutils.GetDeployItemsOfInstallation(ctx, f.Client, inst)
			utils.ExpectNoError(err)
			Expect(deployItems).To(HaveLen(1))
			Expect(deployItems[0].Status.DeployItemPhase).To(Equal(lsv1alpha1.DeployItemPhaseSucceeded))
			Expect(deployItems[0].Status.JobIDFinished).To(Equal(deployItems[0].Status.JobID))

			// todo: make namespace configurable for deployed resources
			// expect that the echo server deployment is successfully running
			echoServerDeploymentName := "echo-server"
			echoServerDeploymentObjectKey := kutil.ObjectKey(echoServerDeploymentName, "default")
			utils.ExpectNoError(utils.WaitForDeploymentToBeReady(ctx, f.TestLog(), f.Client, echoServerDeploymentObjectKey, 2*time.Minute))
			// expect that the deployment has the correct resource requests and limits
			echoServerDeploy := &appsv1.Deployment{}
			utils.ExpectNoError(f.Client.Get(ctx, echoServerDeploymentObjectKey, echoServerDeploy))
			Expect(echoServerDeploy.Spec.Template.Spec.Containers).To(HaveLen(1))
			Expect(echoServerDeploy.Spec.Template.Spec.Containers[0].Resources.Requests.Memory().String()).To(Equal("50Mi"))
			Expect(echoServerDeploy.Spec.Template.Spec.Containers[0].Resources.Limits.Memory().String()).To(Equal("100Mi"))

			By("Delete installation")
			utils.ExpectNoError(f.Client.Delete(ctx, inst))
			utils.ExpectNoError(utils.WaitForObjectDeletion(ctx, f.Client, inst, 2*time.Minute))

			// expect that the nginx deployment is alread deleted or has an deletion timestamp
			err = f.Client.Get(ctx, echoServerDeploymentObjectKey, echoServerDeploy)
			if err != nil && !apierrors.IsNotFound(err) {
				utils.ExpectNoError(err)
			} else if err == nil {
				Expect(echoServerDeploy.DeletionTimestamp.IsZero()).To(BeTrue())
			}
		})
	})
}
