package vector

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	"github.com/openshift/cluster-logging-operator/test/helpers"

	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Testing Config Generation", func() {
	var f = func(clspec logging.CollectionSpec, secrets map[string]*corev1.Secret, clfspec logging.ClusterLogForwarderSpec, op generator.Options) []generator.Element {
		return generator.MergeElements(
			Inputs(&clfspec, op),
			Pipelines(&clfspec, op),
		)
	}
	DescribeTable("Source(s) to Pipeline(s)", helpers.TestGenerateConfWith(f),
		Entry("Send all log types to output by name", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

# Set log_type to "audit"
[transforms.audit]
type = "remap"
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs","ovn_audit_logs"]
source = '''
  .log_type = "audit"
  .hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""
  ."@timestamp" = del(.timestamp)
'''

[transforms.pipeline]
type = "remap"
inputs = ["application","infrastructure","audit"]
source = '''
  .
'''
`,
		}),
		Entry("Send same logtype to multiple output", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs: []string{
							logging.InputNameApplication,
							logging.InputNameInfrastructure,
							logging.InputNameAudit,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline1",
					},
					{
						InputRefs: []string{
							logging.InputNameApplication,
						},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline2",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

# Set log_type to "audit"
[transforms.audit]
type = "remap"
inputs = ["host_audit_logs","k8s_audit_logs","openshift_audit_logs","ovn_audit_logs"]
source = '''
  .log_type = "audit"
  .hostname = get_env_var("VECTOR_SELF_NODE_NAME") ?? ""
  ."@timestamp" = del(.timestamp)
'''

[transforms.pipeline1]
type = "remap"
inputs = ["application","infrastructure","audit"]
source = '''
  .
'''

[transforms.pipeline2]
type = "remap"
inputs = ["application"]
source = '''
  .
'''
`,
		}),
		Entry("Route Logs by Namespace(s)", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "myapplogs",
						Application: &logging.Application{
							Namespaces: []string{"test-ns1", "test-ns2"},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{"myapplogs"},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

[transforms.route_application_logs]
type = "route"
inputs = ["application"]
route.myapplogs = '(.kubernetes.namespace_name == "test-ns1") || (.kubernetes.namespace_name == "test-ns2")'

[transforms.pipeline]
type = "remap"
inputs = ["route_application_logs.myapplogs"]
source = '''
  .
'''
`,
		}),
		Entry("Route Logs by Namespaces(s), and Labels(s)", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Inputs: []logging.InputSpec{
					{
						Name: "myapplogs",
						Application: &logging.Application{
							Namespaces: []string{"myapp1", "myapp2"},
							Selector: &v1.LabelSelector{
								MatchLabels: map[string]string{
									"key1": "value1",
									"key2": "value2",
								},
							},
						},
					},
				},
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{"myapplogs"},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

[transforms.route_application_logs]
type = "route"
inputs = ["application"]
route.myapplogs = '((.kubernetes.namespace_name == "myapp1") || (.kubernetes.namespace_name == "myapp2")) && ((.kubernetes.labels.key1 == "value1") && (.kubernetes.labels.key2 == "value2"))'

[transforms.pipeline]
type = "remap"
inputs = ["route_application_logs.myapplogs"]
source = '''
  .
'''
`,
		}),
		Entry("Add Openshift Label(s)", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication, logging.InputNameInfrastructure},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
						Labels: map[string]string{
							"label1": "value1",
						},
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

[transforms.pipeline]
type = "remap"
inputs = ["application","infrastructure"]
source = '''
  .openshift.labels = {"label1":"value1"}
'''
`,
		}),
		Entry("Parse log message as Jaon", helpers.ConfGenerateTest{
			CLFSpec: logging.ClusterLogForwarderSpec{
				Pipelines: []logging.PipelineSpec{
					{
						InputRefs:  []string{logging.InputNameApplication, logging.InputNameInfrastructure},
						OutputRefs: []string{logging.OutputNameDefault},
						Name:       "pipeline",
						Parse:      "json",
					},
				},
			},
			ExpectedConf: `
[transforms.route_container_logs]
type = "route"
inputs = ["container_logs"]
route.app = '!((starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube"))'
route.infra = '(starts_with!(.kubernetes.namespace_name,"kube-")) || (starts_with!(.kubernetes.namespace_name,"openshift-")) || (.kubernetes.namespace_name == "default") || (.kubernetes.namespace_name == "openshift") || (.kubernetes.namespace_name == "kube")'

# Set log_type to "application"
[transforms.application]
type = "remap"
inputs = ["route_container_logs.app"]
source = '''
  .log_type = "application"
'''

# Set log_type to "infrastructure"
[transforms.infrastructure]
type = "remap"
inputs = ["route_container_logs.infra","journal_logs"]
source = '''
  .log_type = "infrastructure"
'''

[transforms.pipeline]
type = "remap"
inputs = ["application","infrastructure"]
source = '''
  parsed, err = parse_json(.message)
  if err == null {
    .structured = parsed
    del(.message)
  }
'''
`,
		}),
	)
})
