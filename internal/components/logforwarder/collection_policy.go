package logforwarder

import (
	"context"
	"regexp"
	"sort"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obsv1beta1 "github.com/openshift/cluster-logging-operator/api/observability/v1beta1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LabelKubernetesMetaDataName = "kubernetes.io/metadata.name"
)

// Evaluate all cld to determine if THIS one should service it.
// lf.distributionClass == cld.Name
// cld.excludes[].namespace == lf.ns && cld.excludes[].container incluces ls.excludes
// cld.includes[].namespace == lf.ns && cld.includes[].container incluces ls.includes
// cld.namespaceselector
func AssignDistributor(k8sClient client.Client, lf internalobs.LogForwarder) string {
	log.V(0).Info("Assigning a distributor for LogForwarder", "logforwarder", lf)
	distributors := loadClusterLogDistributors(k8sClient)
	if len(distributors) == 0 {
		log.V(0).Info("No ClusterLogDistributors available to evaluate for LogForwarder", "namespace", lf.Namespace, "logforwarder", lf.Name)
		return ""
	}
	sort.Slice(distributors, func(i, j int) bool {
		return distributors[i].Spec.Priority > distributors[j].Spec.Priority
	})

	for _, cld := range distributors {
		log.V(0).Info("Evaluating distributor for LogForwarder", "ClusterLogdistributor", cld)

		log.V(0).Info("Evaluating distributionClass...")
		if lfCollectionPolicy := lf.Spec.Input.CollectionPolicy; lfCollectionPolicy != nil && lfCollectionPolicy.DistributionClass == cld.Name {
			log.V(0).Info("Matched LogForwarder to distribution class", "ClusterLogForwarder", cld.Name, "namespace", lf.Namespace(), "logforwarder", lf.Name())
			return cld.Name
		}

		log.V(0).Info("Evaluating excludes...")
		collectionPolicy := cld.Spec.CollectionPolicy
		if hasNamespace(collectionPolicy.Container.Excludes, lf.Namespace(), []obsv1.NamespaceContainerSpec{}) {
			log.V(0).Info("Namespace is excluded for LogForwarder", "ClusterLogForwarder", cld.Name, "namespace", lf.Namespace(), "logforwarder", lf.Name())
			continue
		}

		log.V(0).Info("Evaluating includes...")
		if hasNamespace(collectionPolicy.Container.Includes, lf.Namespace(), []obsv1.NamespaceContainerSpec{{Namespace: "*"}}) {
			log.V(0).Info("Matched namespace for LogForwarder", "ClusterLogForwarder", cld.Name, "namespace", lf.Namespace(), "logforwarder", lf.Name())
			return cld.Name
		}
		log.V(0).Info("Evaluating namespace selector ...")
		if namespaceMatchesSelector(k8sClient, collectionPolicy.Container.NamespaceMatchLabels, lf) {
			log.V(0).Info("Matched namespace selector for LogForwarder", "ClusterLogForwarder", cld.Name, "namespace", lf.Namespace(), "logforwarder", lf.Name())
			return cld.Name
		}

	}
	log.V(0).Info("Unable to match LogForwarder to any collection policy", "namespace", lf.Namespace(), "logforwarder", lf.Name())
	return ""
}

func loadClusterLogDistributors(k8sClient client.Client) (clds []obsv1beta1.ClusterLogDistributor) {
	log.V(0).Info("Fetching ClusterLogDistributors...")
	list := &obsv1beta1.ClusterLogDistributorList{}
	if err := k8sClient.List(context.TODO(), list); err != nil {
		log.Error(err, "Unable to list ClusterLogDistributors")
		return clds
	}
	if len(list.Items) > 0 {
		return list.Items
	}
	return clds
}

func namespaceMatchesSelector(k8sClient client.Client, matchLabels client.MatchingLabels, lf internalobs.LogForwarder) bool {
	if matchLabels == nil {
		matchLabels = client.MatchingLabels{}
	}
	matchLabels[LabelKubernetesMetaDataName] = lf.LogForwarder.Namespace
	log.V(0).Info("Listing namespaces", lf.LogForwarder.Namespace, "matchLabels", matchLabels)
	namespaces := &corev1.NamespaceList{}
	if err := k8sClient.List(context.TODO(), namespaces, matchLabels); err != nil {
		log.Error(err, "Unable to evaluate LogForwarder namespace for matching label selectors", "namespace", lf.Namespace(), "logforwarder", lf.Name(), "matchLabels", matchLabels)
		return false
	}
	if len(namespaces.Items) == 1 {
		log.V(0).Info("Matched ClusterLogDistributor matchlabels to LogForwarder namesspace", "namespace", lf.Namespace(), "logforwarder", lf.Name(), "matchLabels", matchLabels)
		return true
	}
	return false
}

func hasNamespace(specs []obsv1.NamespaceContainerSpec, namespace string, whenSpecsEmpty []obsv1.NamespaceContainerSpec) bool {
	log.V(0).Info("#hasNamespace", "specs", specs, "whenSpecsEmpty", whenSpecsEmpty)
	if len(specs) == 0 && len(whenSpecsEmpty) != 0 {
		specs = whenSpecsEmpty
	}
	for _, s := range specs {
		pattern := strings.ReplaceAll(s.Namespace, "*", ".*")
		match, err := regexp.MatchString(pattern, namespace)
		if err != nil {
			log.Error(err, "Error trying to evaluate namespace. PANIC", "namespace", namespace, "pattern", pattern)
			panic(err)
		}
		if match {
			return true
		}
	}
	return false
}
