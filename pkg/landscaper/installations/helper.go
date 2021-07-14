// SPDX-FileCopyrightText: 2020 SAP SE or an SAP affiliate company and Gardener contributors.
//
// SPDX-License-Identifier: Apache-2.0

package installations

import (
	"context"
	"encoding/json"
	"fmt"

	cdv2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	lsv1alpha1 "github.com/gardener/landscaper/apis/core/v1alpha1"
	"github.com/gardener/landscaper/apis/core/v1alpha1/helper"
	lscheme "github.com/gardener/landscaper/pkg/api"
	"github.com/gardener/landscaper/pkg/landscaper/blueprints"
	"github.com/gardener/landscaper/pkg/landscaper/dataobjects"
	"github.com/gardener/landscaper/pkg/utils/kubernetes"
)

var componentInstallationGVK schema.GroupVersionKind

func init() {
	var err error
	componentInstallationGVK, err = apiutil.GVKForObject(&lsv1alpha1.Installation{}, lscheme.LandscaperScheme)
	runtime.Must(err)
}

// IsRootInstallation returns if the installation is a root element.
func IsRootInstallation(inst *lsv1alpha1.Installation) bool {
	_, isOwned := kubernetes.OwnerOfGVK(inst.OwnerReferences, componentInstallationGVK)
	return !isOwned
}

// GetParentInstallationName returns the name of parent installation that encompasses the given installation.
func GetParentInstallationName(inst *lsv1alpha1.Installation) string {
	name, _ := kubernetes.OwnerOfGVK(inst.OwnerReferences, componentInstallationGVK)
	return name
}

// CreateInternalInstallations creates internal installations for a list of ComponentInstallations
func CreateInternalInstallations(ctx context.Context, compResolver ctf.ComponentResolver, installations ...*lsv1alpha1.Installation) ([]*Installation, error) {
	internalInstallations := make([]*Installation, len(installations))
	for i, inst := range installations {
		inInst, err := CreateInternalInstallation(ctx, compResolver, inst)
		if err != nil {
			return nil, err
		}
		internalInstallations[i] = inInst
	}
	return internalInstallations, nil
}

// CreateInternalInstallationBases creates internal installation bases for a list of ComponentInstallations
func CreateInternalInstallationBases(installations ...*lsv1alpha1.Installation) ([]*InstallationBase, error) {
	internalInstallations := make([]*InstallationBase, len(installations))
	for i, inst := range installations {
		inInst := CreateInternalInstallationBase(inst)
		internalInstallations[i] = inInst
	}
	return internalInstallations, nil
}

// ResolveComponentDescriptor resolves the component descriptor of a installation.
// Inline Component Descriptors take precedence
func ResolveComponentDescriptor(ctx context.Context, compRepo ctf.ComponentResolver, inst *lsv1alpha1.Installation) (*cdv2.ComponentDescriptor, ctf.BlobResolver, error) {
	if inst.Spec.ComponentDescriptor == nil || (inst.Spec.ComponentDescriptor.Reference == nil && inst.Spec.ComponentDescriptor.Inline == nil) {
		return nil, nil, nil
	}
	var (
		repoCtx *cdv2.UnstructuredTypedObject
		ref     cdv2.ObjectMeta
	)
	//case inline component descriptor
	if inst.Spec.ComponentDescriptor.Inline != nil {
		repoCtx = inst.Spec.ComponentDescriptor.Inline.GetEffectiveRepositoryContext()
		ref = inst.Spec.ComponentDescriptor.Inline.ObjectMeta
	}
	// case remote reference
	if inst.Spec.ComponentDescriptor.Reference != nil {
		repoCtx = inst.Spec.ComponentDescriptor.Reference.RepositoryContext
		ref = inst.Spec.ComponentDescriptor.Reference.ObjectMeta()
	}
	return compRepo.ResolveWithBlobResolver(ctx, repoCtx, ref.GetName(), ref.GetVersion())
}

// CreateInternalInstallation creates an internal installation for an Installation
func CreateInternalInstallation(ctx context.Context, compResolver ctf.ComponentResolver, inst *lsv1alpha1.Installation) (*Installation, error) {
	cdRef := GetReferenceFromComponentDescriptorDefinition(inst.Spec.ComponentDescriptor)
	blue, err := blueprints.Resolve(ctx, compResolver, cdRef, inst.Spec.Blueprint)
	if err != nil {
		return nil, fmt.Errorf("unable to resolve blueprint for %s/%s: %w", inst.Namespace, inst.Name, err)
	}
	return New(inst, blue)
}

// CreateInternalInstallationBase creates an internal installation base for an Installation
func CreateInternalInstallationBase(inst *lsv1alpha1.Installation) *InstallationBase {
	return NewInstallationBase(inst)
}

// GetDataImport fetches the data import from the cluster.
func GetDataImport(ctx context.Context, kubeClient client.Client, contextName string, inst *InstallationBase,
	dataImport lsv1alpha1.DataImport) (*dataobjects.DataObject, *v1.OwnerReference, error) {
	var rawDataObject *lsv1alpha1.DataObject
	// get deploy item from current context
	if len(dataImport.DataRef) != 0 {
		rawDataObject = &lsv1alpha1.DataObject{}
		doName := helper.GenerateDataObjectName(contextName, dataImport.DataRef)
		if err := kubeClient.Get(ctx, kubernetes.ObjectKey(doName, inst.Info.Namespace), rawDataObject); err != nil {
			return nil, nil, fmt.Errorf("unable to fetch data object %s (%s/%s): %w", doName, contextName, dataImport.DataRef, err)
		}
	}
	if dataImport.SecretRef != nil {
		_, data, gen, err := ResolveSecretReference(ctx, kubeClient, dataImport.SecretRef)
		if err != nil {
			return nil, nil, err
		}
		rawDataObject = &lsv1alpha1.DataObject{}
		rawDataObject.Data.RawMessage = data
		// set the generation as it is used to detect outdated imports.
		rawDataObject.SetGeneration(gen)
	}
	if dataImport.ConfigMapRef != nil {
		_, data, gen, err := ResolveConfigMapReference(ctx, kubeClient, dataImport.ConfigMapRef)
		if err != nil {
			return nil, nil, err
		}
		rawDataObject = &lsv1alpha1.DataObject{}
		rawDataObject.Data.RawMessage = []byte(data)
		// set the generation as it is used to detect outdated imports.
		rawDataObject.SetGeneration(gen)
	}

	do, err := dataobjects.NewFromDataObject(rawDataObject)
	if err != nil {
		return nil, nil, err
	}

	owner := kubernetes.GetOwner(do.Raw.ObjectMeta)
	return do, owner, nil
}

// GetTargetImport fetches the target import from the cluster.
func GetTargetImport(ctx context.Context, kubeClient client.Client, contextName string, inst *Installation, targetName string) (*dataobjects.Target, error) {
	// get deploy item from current context
	raw := &lsv1alpha1.Target{}
	targetName = helper.GenerateDataObjectName(contextName, targetName)
	if err := kubeClient.Get(ctx, kubernetes.ObjectKey(targetName, inst.Info.Namespace), raw); err != nil {
		return nil, err
	}

	target, err := dataobjects.NewFromTarget(raw)
	if err != nil {
		return nil, fmt.Errorf("unable to create internal target for %s: %w", targetName, err)
	}
	return target, nil
}

// GetTargetListImportByNames fetches the target imports from the cluster, based on a list of target names.
func GetTargetListImportByNames(ctx context.Context, kubeClient client.Client, contextName string, inst *Installation, targetNames []string) (*dataobjects.TargetList, error) {
	targets := make([]lsv1alpha1.Target, len(targetNames))
	for i, targetName := range targetNames {
		// get deploy item from current context
		raw := &lsv1alpha1.Target{}
		targetName = helper.GenerateDataObjectName(contextName, targetName)
		if err := kubeClient.Get(ctx, kubernetes.ObjectKey(targetName, inst.Info.Namespace), raw); err != nil {
			return nil, err
		}
		targets[i] = *raw
	}
	targetList, err := dataobjects.NewFromTargetList(targets)
	if err != nil {
		return nil, err
	}

	return targetList, nil
}

// GetTargetListImportBySelector fetches the target imports from the cluster, based on a label selector.
func GetTargetListImportBySelector(ctx context.Context, kubeClient client.Client, contextName string, inst *Installation, selector map[string]string) (*dataobjects.TargetList, error) {
	targets := &lsv1alpha1.TargetList{}
	// add context to selector
	selectorWithContext := map[string]string{}
	for k, v := range selector {
		selectorWithContext[k] = v
	}
	if len(contextName) != 0 {
		// top-level targets probably don't have an empty context set, so only add the selector if there actually is a context
		selectorWithContext[lsv1alpha1.DataObjectContextLabel] = contextName
	}
	if err := kubeClient.List(ctx, targets, client.MatchingLabels(selectorWithContext), client.InNamespace(inst.Info.Namespace)); err != nil {
		return nil, err
	}
	if len(contextName) == 0 {
		// if we didn't add the context label selector, we now need to sort out all targets which have a non-empty context label
		newTargets := &lsv1alpha1.TargetList{}
		newTargets.Items = []lsv1alpha1.Target{}
		for _, t := range targets.Items {
			cls, ok := t.Labels[lsv1alpha1.DataObjectContextLabel]
			if !ok || len(cls) == 0 {
				newTargets.Items = append(newTargets.Items, t)
			}
		}
		targets = newTargets
	}
	targetList, err := dataobjects.NewFromTargetList(targets.Items)
	if err != nil {
		return nil, err
	}
	return targetList, nil
}

// GetComponentDescriptorImport fetches the component descriptor import from the cluster/registry.
func GetComponentDescriptorImport(ctx context.Context, kubeClient client.Client, contextName string, op *Operation, imp lsv1alpha1.ComponentDescriptorImport, refType dataobjects.CDReferenceType) (*dataobjects.ComponentDescriptor, error) {
	// default refType if empty
	if len(refType) == 0 {
		if imp.ConfigMapRef != nil {
			refType = dataobjects.ConfigMapReference
		} else if imp.SecretRef != nil {
			refType = dataobjects.SecretReference
		} else if imp.CDRef != nil {
			refType = dataobjects.RegistryReference
		} else if len(imp.DataRef) != 0 {
			refType = dataobjects.DataReference
		} else {
			return nil, fmt.Errorf("invalid component descriptor import '%s': none of dataRef, configMapRef, secretRef, and componentDescriptorRef is specified", imp.Name)
		}
	}

	res := dataobjects.NewComponentDescriptor()
	owner := kubernetes.GetOwner(op.Inst.Info.ObjectMeta)
	if owner != nil && owner.Kind == "Installation" {
		res.SetOwner(owner)
	}
	switch refType {
	case dataobjects.DataReference:
		// resolving data references is hard at this point, therefore they are replaced during the subinstallation template rendering
		// this means that there shouldn't be any data reference at this point
		return nil, fmt.Errorf("unsupported reference type '%s'", string(refType))
	case dataobjects.RegistryReference:
		// fetch component descriptor from registry
		if imp.CDRef == nil {
			return nil, fmt.Errorf("reference type mismatch: reference type is '%s', but CDRef is nil", string(refType))
		}
		cd, err := op.ComponentsRegistry().Resolve(ctx, imp.CDRef.RepositoryContext, imp.CDRef.ComponentName, imp.CDRef.Version)
		if err != nil {
			return nil, fmt.Errorf("unable to get component descriptor from registry (%v): %w", imp.CDRef, err)
		}
		res.SetRegistryReference(imp.CDRef).SetDescriptor(cd)
	case dataobjects.ConfigMapReference:
		_, data, _, err := ResolveConfigMapReference(ctx, kubeClient, imp.ConfigMapRef)
		if err != nil {
			return nil, fmt.Errorf("unable to get component descriptor from configmap %s: %w", imp.ConfigMapRef.NamespacedName().String(), err)
		}
		cd := &cdv2.ComponentDescriptor{}
		err = json.Unmarshal([]byte(data), cd)
		if err != nil {
			return nil, fmt.Errorf("unable to convert data into component descriptor: %w", err)
		}
		res.SetConfigMapReference(imp.ConfigMapRef).SetDescriptor(cd)
	case dataobjects.SecretReference:
		_, data, _, err := ResolveSecretReference(ctx, kubeClient, imp.SecretRef)
		if err != nil {
			return nil, fmt.Errorf("unable to get component descriptor from secret %s: %w", imp.SecretRef.NamespacedName().String(), err)
		}
		cd := &cdv2.ComponentDescriptor{}
		err = json.Unmarshal(data, cd)
		if err != nil {
			return nil, fmt.Errorf("unable to convert data into component descriptor: %w", err)
		}
		res.SetSecretReference(imp.SecretRef).SetDescriptor(cd)
	default:
		return nil, fmt.Errorf("unknown reference type '%s' for component descriptor import", string(refType))
	}

	return res, nil
}

// GeReferenceFromComponentDescriptorDefinition tries to extract a component descriptor reference from a given component descriptor definition
// GetReferenceFromComponentDescriptorDefinition tries to extract a component descriptor reference from a given component descriptor definition
func GetReferenceFromComponentDescriptorDefinition(cdDef *lsv1alpha1.ComponentDescriptorDefinition) *lsv1alpha1.ComponentDescriptorReference {
	if cdDef == nil {
		return nil
	}

	if cdDef.Inline != nil {
		repoCtx := cdDef.Inline.GetEffectiveRepositoryContext()
		return &lsv1alpha1.ComponentDescriptorReference{
			RepositoryContext: repoCtx,
			ComponentName:     cdDef.Inline.Name,
			Version:           cdDef.Inline.Version,
		}
	}

	return cdDef.Reference
}

// ResolveSecretReference is an auxiliary function that fetches the content of a secret as specified by the given SecretReference
// The first returned value is the complete secret content, the second one the specified key (if set), the third one is the generation of the secret
func ResolveSecretReference(ctx context.Context, kubeClient client.Client, secretRef *lsv1alpha1.SecretReference) (map[string][]byte, []byte, int64, error) {
	secret := &corev1.Secret{}
	if err := kubeClient.Get(ctx, secretRef.NamespacedName(), secret); err != nil {
		return nil, nil, 0, err
	}
	completeData := secret.Data
	var data []byte
	var ok bool
	if len(secretRef.Key) != 0 {
		data, ok = secret.Data[secretRef.Key]
		if !ok {
			return nil, nil, 0, fmt.Errorf("key %s in secret %s does not exist", secretRef.Key, secretRef.NamespacedName().String())
		}
	}
	return completeData, data, secret.Generation, nil
}

// ResolveConfigMapReference is an auxiliary function that fetches the content of a configmap as specified by the given ConfigMapReference
// The first returned value is the complete configmap content, the second one the specified key (if set), the third one is the generation of the configmap
func ResolveConfigMapReference(ctx context.Context, kubeClient client.Client, configMapRef *lsv1alpha1.ConfigMapReference) (map[string]string, string, int64, error) {
	cm := &corev1.ConfigMap{}
	if err := kubeClient.Get(ctx, configMapRef.NamespacedName(), cm); err != nil {
		return nil, "", 0, err
	}
	completeData := cm.Data
	var data string
	var ok bool
	if len(configMapRef.Key) != 0 {
		data, ok = cm.Data[configMapRef.Key]
		if !ok {
			return nil, "", 0, fmt.Errorf("key %s in configmap %s does not exist", configMapRef.Key, configMapRef.NamespacedName().String())
		}
	}
	return completeData, data, cm.Generation, nil
}
