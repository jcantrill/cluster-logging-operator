package certificates

import (
	"fmt"
	"os/exec"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
)

func GenerateCertificates(namespace, scriptsDir, logStoreName, workDir string) (err error) {
	script := fmt.Sprintf("%s/cert_generation.sh", scriptsDir)
	return RunCertificatesScript(namespace, logStoreName, workDir, script)
}

func RunCertificatesScript(namespace, logStoreName, workDir, script string) (err error) {
	logger.Debugf("Running script '%s %s %s %s'", script, workDir, namespace, logStoreName)
	cmd := exec.Command(script, workDir, namespace, logStoreName)
	result, err := cmd.Output()
	if logger.IsDebugEnabled() {
		logger.Debugf("cert_generation output: %s", string(result))
		logger.Debugf("err: %v", err)
	}
	return err
}
