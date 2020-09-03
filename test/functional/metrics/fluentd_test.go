package metrics

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/openshift/cluster-logging-operator/internal/pkg/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/pkg/certificates"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/openshift/cluster-logging-operator/test/builder"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"
)

var _ = Describe("[Metrics] Fluentd", func() {

	const (
		clusterLogForwarder = `
apiVersion: "logging.openshift.io/v1"
kind: "ClusterLogForwarder"
metadata:
  name: "instance"
spec:
  outputs:
  - name: forward
    type: fluentdForward
    url:  https://localhost
  pipelines:
  - name: test-pipeline
    inputrefs: ["application"]
    outputrefs: ["forward"]
`
		cmdTemplate = "curl -ks https://%s.%s:24231/metrics"
	)

	var (
		testName       string
		namespace      string
		fluentConf     string
		err            error
		image          = utils.GetComponentImage(constants.FluentdName)
		labels         map[string]string
		maxDuration, _ = time.ParseDuration("2m")
	)

	BeforeEach(func() {

		labels = map[string]string{
			"testtype": "functional",
			"testname": testName,
		}
		testName = fmt.Sprintf("test-fluent-%d", rand.Intn(1000))
		namespace = testName
		if err := oc.Literal().From(fmt.Sprintf("oc create ns %s", namespace)).Output(); err != nil {
			Fail(fmt.Sprintf("Error creating test ns: %v", err))
		}
		//generate empty config
		if fluentConf, err = forwarder.Generate(clusterLogForwarder, false); err != nil {
			Fail(fmt.Sprintf("Error generating configuration %v", err))
		}
		if err = certificates.GenerateCertificates(namespace,
			utils.GetScriptsDir(), "elasticsearch",
			utils.DefaultWorkingDir); err != nil {
			Fail(fmt.Sprintf("Error generating secrets %v", err))
		}
		if _, err := builder.NewConfigMapBuilder(namespace, testName).
			Add("fluent.conf", fluentConf).
			Add("run.sh", string(utils.GetFileContents(utils.GetShareDir()+"/fluentd/run.sh"))).
			Create(); err != nil {
			Fail(fmt.Sprintf("Error creating fluent configmap: %v", err))
		}
		certsName := "certs-" + testName
		if _, err := builder.NewConfigMapBuilder(namespace, certsName).
			Add("tls.key", string(utils.GetWorkingDirFileContents("system.logging.fluentd.key"))).
			Add("tls.crt", string(utils.GetWorkingDirFileContents("system.logging.fluentd.crt"))).
			Create(); err != nil {
			Fail(fmt.Sprintf("Error creating fluent configmap: %v", err))
		}

		if _, err := builder.NewPodBuilder(namespace, testName).
			WithLabels(labels).
			AddConfigMapVolume("config", testName).
			AddConfigMapVolume("entrypoint", testName).
			AddConfigMapVolume("certs", certsName).
			AddContainer(testName, image).
			AddVolumeMount("config", "/etc/fluent/configs.d/user", "", true).
			AddVolumeMount("entrypoint", "/opt/app-root/src/run.sh", "run.sh", true).
			AddVolumeMount("certs", "/etc/fluent/metrics", "", true).
			End().
			Create(); err != nil {
			Fail(fmt.Sprintf("Error creating pod: %v", err))
		}
		if _, err := builder.NewServiceBuilder(namespace, testName).
			AddServicePort(24231, 24231).WithSelector(labels).Create(); err != nil {
			Fail(fmt.Sprintf("Error creating service: %v", err))
		}
		if err := oc.Literal().From(fmt.Sprintf("oc wait -n %s pod/%s --timeout=60s --for=condition=Ready", namespace, testName)).Output(); err != nil {
			Fail(fmt.Sprintf("Error waiting for pod to start: %v", err))
		}
	})

	AfterEach(func() {
		oc.Literal().From(fmt.Sprintf("oc delete ns %s --force - --grace-period=0 --ignore-not-found", namespace)).Output()
	})

	Context("when using a service address", func() {
		It("should return successfully", func() {
			cmd := strings.Split(fmt.Sprintf(cmdTemplate, testName, namespace), " ")
			Expect(oc.Exec().Pod(testName).WithNamespace(namespace).WithCmd(cmd[0], cmd[1:]...).RunFor(maxDuration)).ToNot(BeEmpty())
		})
	})
	// Context("when using the podIP", func() {
	// 	It("should return successfully", func() {

	// 	})
	// })

})
