package splunk

import (
	"bytes"
	"fmt"
	log "github.com/ViaQ/logerr/v2/log/static"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test/client"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
	"github.com/openshift/cluster-logging-operator/test/helpers/rand"
	v1 "k8s.io/api/core/v1"
	"strings"
	"text/template"
)

const (
	SplunkImage   = "quay.io/openshift-logging/splunk:9.0.0"
	SplunkHecPort = 8088
)

var (
	HecToken      = rand.Word(16)
	AdminPassword = rand.Word(16)

	configTemplateName = "splunkserver"
	ConfigTemplate     = `
  splunk:
    hec:
      ssl: false
      token: "{{ string .Token }}"
    password: "{{ string .Password }}"
    idxc_secret: "{{ string .IdxcSecret }}"
    shc_secret: "{{ string .SHCSecret }}"
`
	SplunkEndpointHTTP = fmt.Sprintf("http://localhost:%d", SplunkHecPort)
)

func AddOutput(kubeclient *client.Client, b *runtime.PodBuilder, output logging.OutputSpec) error {
	data, err := GenerateConfigmapData()
	if err != nil {
		return err
	}
	config := runtime.NewConfigMap(b.Pod.Namespace, logging.OutputTypeSplunk, data)
	log.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name)
	if err := kubeclient.Create(config); err != nil {
		return err
	}
	cb := b.AddContainer(logging.OutputTypeSplunk, SplunkImage).
		AddContainerPort("http-splunkweb", 8000).
		AddContainerPort("http-hec", SplunkHecPort).
		AddContainerPort("https-splunkd", 8089).
		AddContainerPort("tcp-s2s", 9097).
		AddEnvVar("SPLUNK_DECLARATIVE_ADMIN_PASSWORD", "true").
		AddEnvVar("SPLUNK_DEFAULTS_URL", "/mnt/splunk-secrets/default.yml").
		AddEnvVar("SPLUNK_DECLARATIVE_ADMIN_PASSWORD", "true").
		AddEnvVar("SPLUNK_HOME_OWNERSHIP_ENFORCEMENT", "false").
		AddEnvVar("SPLUNK_ROLE", "splunk_standalone").
		AddEnvVar("SPLUNK_START_ARGS", "--accept-license").
		AddVolumeMount(config.Name, "/mnt/splunk-secrets", "", true).
		AddVolumeMount("optvar", "/opt/splunk/var", "", false).
		AddVolumeMount("optetc", "/opt/splunk/etc", "", false).
		WithPrivilege()
	cb.Container.LivenessProbe = &v1.Probe{
		ProbeHandler: v1.ProbeHandler{
			Exec: &v1.ExecAction{
				Command: []string{
					"/sbin/checkstate.sh",
				},
			},
		},
		InitialDelaySeconds: 300,
		TimeoutSeconds:      30,
		PeriodSeconds:       30,
		SuccessThreshold:    1,
		FailureThreshold:    3,
	}
	cb.End()
	b.AddConfigMapVolume(config.Name, config.Name)
	b.AddEmptyDirVolume("optvar")
	b.AddEmptyDirVolume("optetc")
	return nil
}

func GenerateConfigmapData() (data map[string]string, err error) {
	b := &bytes.Buffer{}
	t := template.Must(
		template.New(configTemplateName).
			Funcs(template.FuncMap{
				"string": func(arg []byte) string {
					return string(arg)
				},
			}).
			Parse(ConfigTemplate),
	)
	if err = t.Execute(b,
		struct {
			Token        []byte
			Password     []byte
			Pass4SymmKey []byte
			IdxcSecret   []byte
			SHCSecret    []byte
		}{
			Token:        HecToken,
			Password:     AdminPassword,
			Pass4SymmKey: []byte("o4a9itWyG1YECvxpyVV9faUO"),
			IdxcSecret:   []byte("5oPyAqIlod4sxH1Xk7fZpNe4"),
			SHCSecret:    []byte("77mwFNOSUzmQLG9EGa2ZVEFq"),
		},
	); err != nil {
		log.V(3).Error(err, "Error executing template")
		return data, err
	}
	data = map[string]string{
		"default.yml": b.String(),
	}

	return data, nil
}

func ReadLogs(kubeclient *client.Client, namespace, name, logType string) (results []string, err error) {
	cmd := fmt.Sprintf(`/opt/splunk/bin/splunk search logtype=%s -auth "admin:%s"`, logType, AdminPassword)
	output, err := oc.Exec().WithNamespace(namespace).Pod(name).Container(logging.OutputTypeSplunk).WithCmd("/bin/sh", "-c", cmd).Run()
	if err != nil {
		return nil, err
	}
	if output == "" {
		return nil, fmt.Errorf("No logs were found for logType: %s", logType)
	}
	results = strings.Split(output, "\n")
	return results, nil
}
