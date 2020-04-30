package k8shandler

import (
	"encoding/json"
	"fmt"
	"strings"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (clusterRequest *ClusterLoggingRequest) generateCollectorConfig() (config string, err error) {
	switch clusterRequest.cluster.Spec.Collection.Logs.Type {
	case logging.LogCollectionTypeFluentd:
		break
	default:
		return "", fmt.Errorf("%s collector does not support pipelines feature", clusterRequest.cluster.Spec.Collection.Logs.Type)
	}

	spec, status := clusterRequest.normalizeForwarder()
	clusterRequest.ForwarderSpec = *spec
	clusterRequest.ForwarderRequest.Status = *status

	// TODO(alanconway) get rid of legacy/old stuff.
	generator, err := forwarding.NewConfigGenerator(clusterRequest.cluster.Spec.Collection.Logs.Type, clusterRequest.includeLegacyForwardConfig(), clusterRequest.includeLegacySyslogConfig(), clusterRequest.useOldRemoteSyslogPlugin())
	if err != nil {
		logger.Warnf("Unable to create collector config generator: %v", err)
		return "",
			clusterRequest.UpdateCondition(
				logging.CollectorDeadEnd,
				"Unable to generate collector configuration",
				"No defined logstore destination",
				v1.ConditionTrue,
			)
	}
	generatedConfig, err := generator.Generate(&clusterRequest.ForwarderSpec)
	if err != nil {
		logger.Warnf("Unable to generate log confguraiton: %v", err)
		return "",
			clusterRequest.UpdateCondition(
				logging.CollectorDeadEnd,
				"Collectors are defined but there is no defined LogStore or LogForward destinations",
				"No defined logstore destination",
				corev1.ConditionTrue,
			)
	}
	// else
	err = clusterRequest.UpdateCondition(
		logging.CollectorDeadEnd,
		"",
		"",
		corev1.ConditionFalse,
	)

	return generatedConfig, err
}

func jsonString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(b)
}

// normalizeForwarder normalizes the clusterRequest.ForwarderSpec, returns a normalized spec and status.
func (clusterRequest *ClusterLoggingRequest) normalizeForwarder() (*logging.ClusterLogForwarderSpec, *logging.ClusterLogForwarderStatus) {
	logger.Debugf("Normalizing ClusterLogForwarder from request: %v", jsonString(clusterRequest))

	// Default configuration for empty/missing forwarder, forward to the default store
	if len(clusterRequest.ForwarderSpec.Pipelines) == 0 {
		logger.Debug("Configuring forwarder to use the default log store")
		clusterRequest.ForwarderSpec.Pipelines = []logging.PipelineSpec{
			{
				InputRefs:  logging.ReservedInputNames.List(),
				OutputRefs: []string{logging.OutputNameDefault},
			},
		}
	}

	spec := &logging.ClusterLogForwarderSpec{}
	status := &logging.ClusterLogForwarderStatus{}

	clusterRequest.verifyInputs(spec, status)
	clusterRequest.verifyOutputs(spec, status)
	clusterRequest.verifyPipelines(spec, status)

	routes := logging.NewRoutes(spec.Pipelines) // Compute used inputs/outputs

	// Add Ready=true status for all surviving inputs.
	status.Inputs = logging.NamedConditions{}
	inRefs := sets.StringKeySet(routes.ByInput).List()
	for _, inRef := range inRefs {
		status.Inputs.Get(inRef).SetNew(logging.ConditionReady, true, "", "")
	}

	// Determine overall health
	degraded := []string{}
	unready := []string{}
	for name, conds := range status.Pipelines {
		if !conds[logging.ConditionReady].IsTrue() {
			unready = append(unready, name)
		}
		if conds[logging.ConditionDegraded].IsTrue() {
			degraded = append(degraded, name)
		}
	}
	status.Conditions = logging.Conditions{}
	if len(unready) == len(status.Pipelines) {
		setInvalid(status.Conditions, "all pipelines invalid: %v", unready)
	} else {
		if len(unready)+len(degraded) > 0 {
			setWarn(status.Conditions, logging.ConditionDegraded, true, logging.ReasonInvalid, "bad pipelines: unready %v, degraded %v", unready, degraded)
		}
		status.Conditions.SetNew(logging.ConditionReady, true, "", "")
		logger.Infof("ClusterLogForwarder is ready")
	}
	return spec, status
}

func setError(conds logging.Conditions, t logging.ConditionType, status bool, r logging.ConditionReason, format string, args ...interface{}) {
	conds.SetNew(t, status, r, format, args...)
	logger.Errorf(format, args...)
}

func setWarn(conds logging.Conditions, t logging.ConditionType, status bool, r logging.ConditionReason, format string, args ...interface{}) {
	conds.SetNew(t, status, r, format, args...)
	logger.Warnf(format, args...)
}

func setInvalid(conds logging.Conditions, format string, args ...interface{}) {
	setError(conds, logging.ConditionReady, false, logging.ReasonInvalid, format, args...)
}

// verifyRefs returns the set of valid refs and a slice of error messages for bad refs.
func verifyRefs(what string, refs []string, allowed sets.String) (sets.String, []string) {
	good, bad := sets.NewString(), sets.NewString()
	for _, ref := range refs {
		if allowed.Has(ref) {
			good.Insert(ref)
		} else {
			bad.Insert(ref)
		}
	}
	msg := []string{}
	if len(bad) > 0 {
		msg = append(msg, fmt.Sprintf("unrecognized %s: %v", what, bad.List()))
	}
	if len(good) == 0 {
		msg = append(msg, fmt.Sprintf("no valid %s", what))
	}
	return good, msg
}

func (clusterRequest *ClusterLoggingRequest) verifyPipelines(spec *logging.ClusterLogForwarderSpec, status *logging.ClusterLogForwarderStatus) {
	// Validate each pipeline and add a status object.
	status.Pipelines = logging.NamedConditions{}
	names := sets.NewString() // Collect pipeline names

	// Known output names, note if "default" is enabled it will already be in the OutputMap()
	outputs := sets.StringKeySet(spec.OutputMap())
	// Known input names, reserved names not in InputMap() we don't expose default inputs.
	inputs := sets.StringKeySet(spec.InputMap()).Union(logging.ReservedInputNames)

	for i, pipeline := range clusterRequest.ForwarderSpec.Pipelines {
		if pipeline.Name == "" {
			pipeline.Name = fmt.Sprintf("pipeline[%v]", i)
		}
		conds := status.Pipelines.Get(pipeline.Name)
		if names.Has(pipeline.Name) {
			pipeline.Name = fmt.Sprintf("pipeline[%v]", i)
			conds = status.Pipelines.Get(pipeline.Name)
			setInvalid(conds, "duplicate pipeline name: %q", pipeline.Name)
			continue
		}
		names.Insert(pipeline.Name)

		goodIn, msgIn := verifyRefs("inputs", pipeline.InputRefs, inputs)
		goodOut, msgOut := verifyRefs("outputs", pipeline.OutputRefs, outputs)
		if msgs := append(msgIn, msgOut...); len(msgs) > 0 { // Something wrong
			msg := strings.Join(msgs, ", ")
			if len(goodIn) == 0 || len(goodOut) == 0 { // All bad, disabled
				setInvalid(conds, "pipeline %q: %v", pipeline.Name, msg)
				continue
			} else { // Some bad, degraded
				setWarn(conds, logging.ConditionDegraded, true, logging.ReasonInvalid, "pipeline %q: %v", pipeline.Name, msg)
			}
		}
		conds.SetNew(logging.ConditionReady, true, "", "")
		spec.Pipelines = append(spec.Pipelines, logging.PipelineSpec{
			Name: pipeline.Name, InputRefs: goodIn.List(), OutputRefs: goodOut.List(),
		})
	}
}

// verifyInputs and set status.Inputs conditions
func (clusterRequest *ClusterLoggingRequest) verifyInputs(spec *logging.ClusterLogForwarderSpec, status *logging.ClusterLogForwarderStatus) {
	// Collect input conditions
	status.Inputs = logging.NamedConditions{}
	for i, input := range clusterRequest.ForwarderSpec.Inputs {
		conds := status.Inputs.Get(input.Name)
		badName := func(format string, args ...interface{}) {
			input.Name = fmt.Sprintf("input[%v]", i)
			conds = status.Inputs.Get(input.Name)
			setInvalid(conds, format, args...)
		}
		switch {
		case input.Name == "":
			badName("input must have a name")
		case logging.ReservedInputNames.Has(input.Name):
			badName("input name %q is reserved", input.Name)
		case len(conds) > 0:
			badName("duplicate name: %q", input.Name)
		case !logging.IsInputTypeName(input.Type):
			setInvalid(conds, "unknown input type: %q", input.Type)
		}
		if len(conds) == 0 {
			conds.SetNew(logging.ConditionReady, true, "", "")
			spec.Inputs = append(spec.Inputs, input)
		} else {
			conds.SetNew(logging.ConditionReady, false, logging.ReasonInvalid, "")
		}
	}
}

func (clusterRequest *ClusterLoggingRequest) verifyOutputs(spec *logging.ClusterLogForwarderSpec, status *logging.ClusterLogForwarderStatus) {
	status.Outputs = logging.NamedConditions{}
	for i, output := range clusterRequest.ForwarderSpec.Outputs {
		conds := status.Outputs.Get(output.Name)
		badName := func(format string, args ...interface{}) {
			output.Name = fmt.Sprintf("output[%v]", i)
			conds = status.Outputs.Get(output.Name)
			setInvalid(conds, format, args...)
		}

		switch {
		case output.Name == "":
			badName("output must have a name")
		case logging.ReservedOutputNames.Has(output.Name):
			badName("output name %q is reserved", output.Name)
		case len(conds) > 0:
			badName("duplicate name: %q", output.Name)
		case !logging.IsOutputTypeName(output.Type):
			setInvalid(conds, "output %q: unknown output type %q", output.Name, output.Type)
		case output.URL == "":
			setInvalid(conds, "output %q: missing URL", output.Name)
		default:
			clusterRequest.verifyOutputSecret(&output, conds)
		}
		if len(conds) == 0 {
			conds.SetNew(logging.ConditionReady, true, "", "")
			spec.Outputs = append(spec.Outputs, output)
		}
	}
	// Add the default output if required and available.
	routes := logging.NewRoutes(clusterRequest.ForwarderSpec.Pipelines)
	if _, ok := routes.ByOutput[logging.OutputNameDefault]; ok {
		conds := status.Outputs.Get(logging.OutputNameDefault)
		if clusterRequest.cluster.Spec.LogStore == nil {
			conds.SetNew(logging.ConditionReady, false, logging.ReasonMissingResource, "no default log store specified")
		} else {
			spec.Outputs = append(spec.Outputs, logging.OutputSpec{
				Name:   logging.OutputNameDefault,
				Type:   logging.OutputTypeElasticsearch,
				URL:    constants.LogStoreURL,
				Secret: &logging.OutputSecretSpec{Name: constants.CollectorSecretName},
			})
			conds.SetNew(logging.ConditionReady, true, "", "")
		}
	}
}

func (clusterRequest *ClusterLoggingRequest) verifyOutputSecret(output *logging.OutputSpec, conds logging.Conditions) {
	if output.Secret == nil {
		return
	}
	name := output.Secret.Name
	if name == "" {
		setInvalid(conds, "secretRef must have a name")
		return
	}
	if _, err := clusterRequest.GetSecret(name); err != nil {
		setError(conds, logging.ConditionReady, false, logging.ReasonMissingResource, "output %q: secret %q not found", output.Name, name)
	}
}
