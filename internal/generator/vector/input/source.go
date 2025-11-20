package input

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	vectorapisources "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/source"
	"k8s.io/utils/set"
)

// NewSource creates an input adapter to generate config for ViaQ sources to collect logs excluding the
// collector container logs from the namespace where the collector is deployed
func NewSource(input internalobs.Input, resNames factory.ForwarderResourceNames, secrets internalobs.Secrets, op framework.Options) ([]framework.Element, []string) {
	els := []framework.Element{}
	ids := []string{}
	switch {
	case input.LogType() == obs.InputTypeApplication.String():
		ib := source.NewContainerPathGlobBuilder()
		eb := source.NewContainerPathGlobBuilder()
		appIncludes := []string{}
		v1Input := input.AsV1InputSpec()
		if v1Input.Application != nil {
			if len(v1Input.Application.Includes) > 0 {
				for _, in := range v1Input.Application.Includes {
					ncs := source.NamespaceContainer{
						Namespace: in.Namespace,
						Container: in.Container,
					}
					ib.AddCombined(ncs)
					appIncludes = append(appIncludes, ncs.Namespace)
				}
			}
			// Need to remove any of the default excluded infra namespaces if they are part of the includes
			excludesList := pruneInfraNS(appIncludes)
			for _, ns := range excludesList {
				ncs := source.NamespaceContainer{
					Namespace: ns,
				}
				eb.AddCombined(ncs)
			}
			if len(v1Input.Application.Excludes) > 0 {
				for _, ex := range v1Input.Application.Excludes {
					ncs := source.NamespaceContainer{
						Namespace: ex.Namespace,
						Container: ex.Container,
					}
					eb.AddCombined(ncs)
				}
			}
		} else {
			// Need to remove any of the default excluded infra namespaces if they are part of the includes
			excludesList := pruneInfraNS(appIncludes)
			for _, ns := range excludesList {
				ncs := source.NamespaceContainer{
					Namespace: ns,
				}
				eb.AddCombined(ncs)
			}
		}
		eb.AddExtensions(excludeExtensions...)
		includes := ib.Build()
		excludes := eb.Build(infraNamespaces...)
		return NewContainerSource(input, includes, excludes, obs.InputTypeApplication, obs.InfrastructureSourceContainer)
	case input.LogType() == obs.InputTypeInfrastructure.String():
		sources := set.Set[obs.InfrastructureSource]{}
		v1Input := input.AsV1InputSpec()
		if v1Input.Infrastructure == nil {
			sources.Insert(obs.InfrastructureSources...)
		} else {
			sources = set.New(v1Input.Infrastructure.Sources...)
			if sources.Len() == 0 {
				sources.Insert(obs.InfrastructureSources...)
			}
		}
		if sources.Has(obs.InfrastructureSourceContainer) {
			infraIncludes := source.NewContainerPathGlobBuilder().AddNamespaces(infraNamespaces...).Build()
			cels, cids := NewContainerSource(input, infraIncludes, loggingExcludes, obs.InputTypeInfrastructure, obs.InfrastructureSourceContainer)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.InfrastructureSourceNode) {
			jels, jids := NewJournalSource(input.AsV1InputSpec())
			els = append(els, jels...)
			ids = append(ids, jids...)
		}
		return els, ids
	case input.LogType() == obs.InputTypeAudit.String():
		sources := set.Set[obs.AuditSource]{}
		v1Input := input.AsV1InputSpec()
		if v1Input.Audit == nil || len(v1Input.Audit.Sources) == 0 {
			sources.Insert(obs.AuditSources...)
		} else {
			sources = set.New(v1Input.Audit.Sources...)
			if sources.Len() == 0 {
				sources.Insert(obs.AuditSources...)
			}
		}
		if sources.Has(obs.AuditSourceAuditd) {
			cels, cids := NewAuditAuditdSource(input.AsV1InputSpec(), op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.AuditSourceKube) {
			cels, cids := NewK8sAuditSource(input.AsV1InputSpec(), op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.AuditSourceOpenShift) {
			cels, cids := NewOpenshiftAuditSource(input.AsV1InputSpec(), op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
		if sources.Has(obs.AuditSourceOVN) {
			cels, cids := NewOVNAuditSource(input.AsV1InputSpec(), op)
			els = append(els, cels...)
			ids = append(ids, cids...)
		}
	case input.LogType() == obs.InputTypeReceiver.String():
		return NewViaqReceiverSource(input.AsV1InputSpec(), resNames, secrets, op)
	case input.ComponentType() == vectorapisources.ComponentTypeVector:
		return NewVectorSource(input)
	}
	return els, ids
}
