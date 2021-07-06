package output

import (
	"github.com/openshift/cluster-logging-operator/internal/builders/config/fluentd"
	fout "github.com/openshift/cluster-logging-operator/internal/builders/config/fluentd/output"
	clffluentd "github.com/openshift/cluster-logging-operator/internal/builders/logforwarder/fluentd"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	corev1 "k8s.io/api/core/v1"
)

type ForwardOutputBuilder struct{
	outSpec logging.OutputSpec
	outForward *fout.OutForward
	secret *corev1.Secret
}

func NewForwardOutputBuilder(outspec logging.OutputSpec, secret *corev1.Secret) *ForwardOutputBuilder {
	builder := fout.NewOutForwardBuilder("**")
	builder.WithHeartBeatType("none").
		WithKeepAlive(true)
	return &ForwardOutputBuilder{
		outSpec: outspec,
		outForward: builder,
		secret: secret,
	}
}

func  setSecurity(b *fout.OutForward, secret *corev1.Secret) {
	if secret != nil {
		security := b.WithSecurity()
		security.WithHostname("#{ENV['NODE_NAME']}")
		security.WithShardKey(string(secret.Data["shared_key"]))
	}
}

func (b *ForwardOutputBuilder) AsList() []string {
	setSecurity(b.outForward, b.secret)
	return fluentd.Label(clffluentd.FormatLabelName(b.outSpec.Name), b.outForward.AsList())
}