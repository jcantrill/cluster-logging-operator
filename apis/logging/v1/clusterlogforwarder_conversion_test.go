package v1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/apis/logging/v2beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("MigrateForward", func() {

	const (
		deploymentNS = "aNamespace"
		saName       = "myserviceaccount"
		name         = "aname"
	)
	It("should convert v1 to v2beta1", func() {
		exp := v2beta1.ClusterLogForwarder{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Namespace:   deploymentNS,
				Annotations: map[string]string{"foo": "bar"},
				Labels:      map[string]string{"xya": "123"},
			},
			Spec: v2beta1.ClusterLogForwarderSpec{
				ServiceAccountName: saName,
				Inputs: []v2beta1.InputSpec{
					{
						Name: "infraName",
						Infrastructure: &v2beta1.Infrastructure{
							Sources: []string{
								v2beta1.InfrastructureSourceContainer,
								v2beta1.InfrastructureSourceNode,
							},
						},
					},
					{
						Name: "auditName",
						Audit: &v2beta1.Audit{
							Sources: []string{
								v2beta1.AuditSourceAuditd,
								v2beta1.AuditSourceKube,
								v2beta1.AuditSourceOVN,
								v2beta1.AuditSourceOpenShift,
							},
						},
					},
					{
						Name: "appName",
						Application: &v2beta1.Application{
							Namespaces: &v2beta1.InclusionSpec{
								Include: []string{"abc", "123"},
								Exclude: []string{"xyz", "987"},
							},
							Containers: &v2beta1.InclusionSpec{
								Include: []string{"abc", "123"},
								Exclude: []string{"xyz", "987"},
							},
						},
					},
				},
			},
		}
		from := &ClusterLogForwarder{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: deploymentNS,
				Name:      "aname",
			},
			Spec: ClusterLogForwarderSpec{
				Inputs: []InputSpec{
					{
						Name: "infraName",
						Infrastructure: &Infrastructure{
							Sources: []string{
								InfrastructureSourceContainer,
								InfrastructureSourceNode,
							},
						},
					},
					{
						Name: "auditName",
						Audit: &Audit{
							Sources: []string{
								AuditSourceAuditd,
								AuditSourceKube,
								AuditSourceOVN,
								AuditSourceOpenShift,
							},
						},
					},
					{
						Name: "appName",
						Application: &Application{
							Namespaces:        []string{"abc", "123"},
							ExcludeNamespaces: []string{"xyz", "987"},
							Selector:          &LabelSelector{},
							ContainerLimit: &LimitSpec{
								MaxRecordsPerSecond: 20,
							},
							Containers: &InclusionSpec{
								Include: []string{"abc", "123"},
								Exclude: []string{"xyz", "987"},
							},
						},
					},
				},
				Pipelines: []PipelineSpec{},
				Outputs:   []OutputSpec{},
				Filters:   []FilterSpec{},
			},
		}
		to := &v2beta1.ClusterLogForwarder{}
		Expect(from.ConvertTo(to)).To(Succeed())
		Expect(to).To(Equal(exp))
	})
})
