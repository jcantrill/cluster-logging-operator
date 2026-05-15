package input_test

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo/v2"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/input/container"
	utilyaml "github.com/openshift/cluster-logging-operator/internal/utils/yaml"
)

var _ = Describe("Input", func() {
	It("should create a file log", func() {
		// Create an InputSpec for application logs
		spec := obs.InputSpec{
			Name: "my-application",
			Type: obs.InputTypeApplication,
			Application: &obs.Application{
				Includes: []obs.NamespaceContainerSpec{
					{Namespace: "my-namespace"},
					{Namespace: "another-namespace", Container: "web-server"},
				},
				Excludes: []obs.NamespaceContainerSpec{
					{Namespace: "test-*"},
				},
			},
		}

		// Generate the receivers
		id, receiver := container.NewSource(spec, obs.InputTypeApplication)
		fmt.Println(id)
		fmt.Println("")
		fmt.Println(utilyaml.MustMarshal(receiver))
		os.Exit(1)
	})
})
