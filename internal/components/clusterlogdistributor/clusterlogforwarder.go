package clusterlogdistributor

import (
	"context"
	"fmt"

	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obsv1beta1 "github.com/openshift/cluster-logging-operator/api/observability/v1beta1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/components/clusterlogforwarder"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ReconcileClusterLogForwarderForDistributor(k8sClient client.Client, cld obsv1beta1.ClusterLogDistributor, forwarders []internalobs.LogForwarder) (err error) {

	if err = ServiceAccount(k8sClient, cld); err != nil {
		return err
	}
	var clf *obsv1.ClusterLogForwarder
	clf, err = clusterlogforwarder.Fetch(k8sClient, constants.OpenshiftNS, cldObjectName(cld))
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	log.V(cldDefaultLogLevel).Info("fetched", "clusterlogforwarder", clf)
	clf = initClusterLogForwarder(clf, cld, forwarders)
	var result controllerutil.OperationResult
	log.V(cldDefaultLogLevel).Info("controllerutil.CreateOrUpdate...", "clusterlogforwarder", clf)
	result, err = controllerutil.CreateOrUpdate(context.TODO(), k8sClient, clf, func() error {
		clf = initClusterLogForwarder(clf, cld, forwarders)
		return nil
	})
	if err != nil {
		log.V(cldDefaultLogLevel).Error(err, "Error creating clusterlogforwarder", "clusterlogforwarder", clf)
		return err
	}
	log.V(cldDefaultLogLevel).Info("ReconcileClusterLogForwarderForDistributor", "clusterlogforwarder", clf, "result", result)
	return nil
}

func initClusterLogForwarder(clf *obsv1.ClusterLogForwarder, cld obsv1beta1.ClusterLogDistributor, forwarders []internalobs.LogForwarder) *obsv1.ClusterLogForwarder {
	clfName := cldObjectName(cld)
	if clf == nil {
		log.V(cldDefaultLogLevel).Info("Initializing new clusterlogforwarder")
		clf = observability.NewClusterLogForwarder(constants.OpenshiftNS, clfName, obsruntime.Initialize, func(clf *obsv1.ClusterLogForwarder) {
			clf.Spec = obsv1.ClusterLogForwarderSpec{
				ServiceAccount: obsv1.ServiceAccount{
					Name: clfName,
				},
			}
		})
		obsruntime.SetCommonLabels(clf, cldLabelName, clfName, cldLabelComponent)
		utils.AddOwnerRefToObject(clf, utils.AsOwner(&cld))
	}
	clf.Spec.Collector = cld.Spec.Collector
	clf.Spec.Inputs = []obsv1.InputSpec{}
	clf.Spec.Pipelines = []obsv1.PipelineSpec{}
	clf.Spec.Outputs = []obsv1.OutputSpec{}
	for _, f := range forwarders {
		name := fmt.Sprintf("lf-%s-%s", f.Namespace(), f.Name())
		input := obsv1.InputSpec{
			Name: name,
			Type: obsv1.InputTypeApplication,
			Application: &obsv1.Application{
				Includes: newNamespaceContainerSpecs(f.Namespace(), f.Spec.Input.Container.Includes, "*"),
				Excludes: newNamespaceContainerSpecs(f.Namespace(), f.Spec.Input.Container.Excludes, ""),
				Selector: f.Spec.Input.Container.Selector,
			},
		}
		clf.Spec.Inputs = append(clf.Spec.Inputs, input)

		output := obsv1.OutputSpec{
			Name: name,
			Type: obsv1.OutputTypeHTTP,
			HTTP: &obsv1.HTTP{
				URLSpec: obsv1.URLSpec{
					URL: fmt.Sprintf("http://%s.%s.svc:%d", f.DeployedName(), f.Namespace(), DefaultLogForwarderPort),
				},
			},
			//TLS: &obsv1.OutputTLSSpec{
			//	InsecureSkipVerify: true,
			//},
		}
		clf.Spec.Outputs = append(clf.Spec.Outputs, output)

		pipeline := obsv1.PipelineSpec{
			Name:       name,
			InputRefs:  []string{name},
			OutputRefs: []string{name},
		}
		clf.Spec.Pipelines = append(clf.Spec.Pipelines, pipeline)
	}
	log.V(cldDefaultLogLevel).Info("Initialized clf for cld", "clusterlogforwarder", clf, "cld", cld.Name)
	return clf
}

func newNamespaceContainerSpecs(ns string, containers []string, all string) (res []obsv1.NamespaceContainerSpec) {
	if all != "" {
		containers = append(containers, "*")
	}
	for _, container := range containers {
		res = append(res, obsv1.NamespaceContainerSpec{
			Namespace: ns,
			Container: container,
		})
	}
	return res
}
