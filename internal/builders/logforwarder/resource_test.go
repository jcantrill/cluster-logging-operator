package logforwarder

import loggingv1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"

var (
	OutputForwardSample = loggingv1.OutputSpec{
		Name: "special-output",
		Type: loggingv1.OutputTypeFluentdForward,
		URL: "tcp://fluentdserver.security.example.com:24224",
	}
)
const (
	example = `
apiVersion: logging.openshift.io/v1
 kind: ClusterLogForwarder
 metadata:
   name: instance 
   namespace: openshift-logging 
 spec:
   inputs:
    - name: special
      application:
       namespaces:  
        - my-devel 
   outputs:
    - name: special-output
      type: fluentdForward
      url: 'tcp://fluentdserver.security.example.com:24224'
    - name: leftover
      type: fluentdForward
      url: 'tcp://fluentdserver.home.example.com:24224'
   pipelines:
    - name: special-to-special
      inputRefs:  
      - special
      - infrastructure
      outputRefs:
      - special-output
      labels:
        clusterId: C1234 
    - name: everythingelse
      inputRefs:
      - infrastructure
      - application
      outputRefs:
      - leftover
`
)
