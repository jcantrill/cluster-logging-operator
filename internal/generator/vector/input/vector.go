package input

import (
	obs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	apisourcesvector "github.com/openshift/cluster-logging-operator/internal/generator/vector/api/sources/vector"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
)

func NewVectorSource(input obs.Input) ([]framework.Element, []string) {
	id := helpers.MakeInputID(input.Name(), "vector")
	els := []framework.Element{
		apisourcesvector.New(id, input.AsInputSpec().ReceiverSpec.Address()),
	}
	return els, []string{id}
}
