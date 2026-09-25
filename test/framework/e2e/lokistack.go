package e2e

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	clolog "github.com/ViaQ/logerr/v2/log/static"
	"github.com/onsi/ginkgo/v2"
	testerrors "github.com/openshift/cluster-logging-operator/test/helpers/errors"
	"github.com/pkg/errors"

	openshiftv1 "github.com/openshift/api/route/v1"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/test"
	lokitesthelper "github.com/openshift/cluster-logging-operator/test/helpers/loki"
	"github.com/openshift/cluster-logging-operator/test/helpers/oc"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	lokistackURI     = "apis/loki.grafana.com/v1/namespaces/%s/lokistacks/%s"
	operatorGroupURI = "apis/operators.coreos.com/v1/namespaces/%s/operatorgroups/%s"
	subscriptionURI  = "apis/operators.coreos.com/v1alpha1/namespaces/%s/subscriptions/%s"

	lokistackCRDURI = "apis/apiextensions.k8s.io/v1/customresourcedefinitions/%s"

	lokiOperatorDeploymentName = "loki-operator-controller-manager"

	// Stateful sets
	lokistackCompactor    = "-compactor"
	lokistackIndexGateway = "-index-gateway"
	lokistackIngester     = "-ingester"
	lokistackRuler        = "-ruler"

	// Deployments
	lokistackDistributor   = "-distributor"
	lokistackGateway       = "-gateway"
	lokistackQuerier       = "-querier"
	lokistackQueryFrontend = "-query-frontend"

	LokistackName = "lokistack-dev"

	// garageName is the name of the Garage S3 storage StatefulSet/Service and the label
	// value used by its manifests (app.kubernetes.io/name=garage).
	garageName = "garage"

	// garageManifestDir is the kustomize directory (relative to the git root) that deploys
	// Garage as an S3-compatible storage backend for the LokiStack.
	garageManifestDir = "hack/manifests/observability-operators/garage"

	// The following credentials/bucket are seeded by the Garage manifests via its
	// --single-node --default-bucket bootstrap and the "test" storage secret.
	garageBucket          = "loki"
	garageRegion          = "garage"
	garageAccessKeyID     = "GKgarageaccesskey0000"
	garageAccessKeySecret = "garagesecretkey1234567890abcdefgh"
)

var lokiOperatorChannel = "stable-6.6"

func init() {
	if value := os.Getenv("LOKI_OPERATOR_CHANNEL"); value != "" {
		lokiOperatorChannel = value
	}
}

type LokistackLogStore struct {
	Name      string
	Namespace string
	tc        *E2ETestFramework
}

// DeployGarage deploys Garage as an S3-compatible storage backend for the LokiStack by
// applying the kustomize manifests at garageManifestDir via the oc wrapper. Garage is started
// with --single-node --default-bucket so it self-provisions its cluster layout, the "loki"
// bucket and access keys on first boot; no manual bootstrap is required.
func (tc *E2ETestFramework) DeployGarage(namespace string) error {
	ginkgo.By("Deploying Garage from manifest: " + garageManifestDir + " to namespace:" + namespace)
	manifestDir := test.GitRoot(garageManifestDir)
	clolog.V(1).Info("deploying garage", "namespace", namespace, "manifests", manifestDir)

	if _, err := oc.Literal().From("oc apply -k %s -n %s", manifestDir, namespace).Run(); err != nil {
		return err
	}
	tc.AddCleanup(func() error {
		_, err := oc.Literal().From("oc delete -k %s -n %s --ignore-not-found", manifestDir, namespace).Run()
		return err
	})

	// Wait for the garage statefulset to become ready before the LokiStack tries to use it.
	return oc.Literal().From("oc -n %s rollout status statefulset/%s --timeout=%s",
		namespace, garageName, defaultTimeout).Output()
}

func (tc *E2ETestFramework) DeployLokiOperator() error {
	ginkgo.By("Deploying Loki Operator to namespace: " + test.OpenshiftOperatorsRedhatNS + " channel: " + lokiOperatorChannel)
	clolog.V(1).Info("deploying loki operator", "namespace", test.OpenshiftOperatorsRedhatNS, "channel", lokiOperatorChannel)
	operatorGroupYaml := `
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: loki-operator-group
  namespace: openshift-operators-redhat
spec:
  targetNamespaces: [ ]
  upgradeStrategy: Default
`
	subscriptionYaml := fmt.Sprintf(`
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: loki-operator
  namespace: %s
spec:
  channel: %s
  name: loki-operator
  source: redhat-operators
  sourceNamespace: openshift-marketplace
  installPlanApproval: Automatic
`, test.OpenshiftOperatorsRedhatNS, lokiOperatorChannel)
	// Create namespace
	tc.CreateNamespace(test.OpenshiftOperatorsRedhatNS)
	tc.AddCleanup(func() error {
		return tc.KubeClient.CoreV1().Namespaces().Delete(context.TODO(), test.OpenshiftOperatorsRedhatNS, metav1.DeleteOptions{})
	})

	// Create operator group
	ogUri := fmt.Sprintf(operatorGroupURI, test.OpenshiftOperatorsRedhatNS, "loki-operator-group")
	err := tc.Client().RESTClient().Post().
		RequestURI(ogUri).
		SetHeader("Content-Type", "application/yaml").
		Body([]byte(operatorGroupYaml)).
		Do(context.TODO()).Error()

	if err != nil {
		return err
	}

	// Create subscription
	subUri := fmt.Sprintf(subscriptionURI, test.OpenshiftOperatorsRedhatNS, "loki-operator")
	err = tc.Client().RESTClient().Post().
		RequestURI(subUri).
		SetHeader("Content-Type", "application/yaml").
		Body([]byte(subscriptionYaml)).
		Do(context.TODO()).Error()

	// Add CRD clean up
	crdURI := fmt.Sprintf(lokistackCRDURI, "lokistacks.loki.grafana.com")
	tc.AddCleanup(func() error {
		return tc.Client().RESTClient().Delete().
			RequestURI(crdURI).
			Do(context.TODO()).Error()
	})

	if err != nil {
		return err
	}

	return tc.WaitForDeployment(test.OpenshiftOperatorsRedhatNS, lokiOperatorDeploymentName, defaultRetryInterval, defaultTimeout)
}

func (tc *E2ETestFramework) DeployLokistackInNamespace(namespace string) (ls *LokistackLogStore, err error) {
	ginkgo.By("Deploying Lokistack in namespace: " + namespace + "/" + LokistackName)
	clolog.V(1).Info("deploying lokistack", "namespace", namespace, "name", LokistackName)
	logStore := &LokistackLogStore{
		Name:      LokistackName,
		Namespace: namespace,
		tc:        tc,
	}

	// Create log reader role
	apiGroups := []string{"loki.grafana.com"}
	resources := []string{"application", "audit", "infrastructure"}
	resourceNames := []string{"logs"}
	verbs := []string{"get"}

	if err := tc.createClusterRole(ClusterRoleAllLogsReader, apiGroups, resources, resourceNames, verbs); err != nil {
		return nil, err
	}

	storageSecretName := garageName + "-secret"

	yaml := fmt.Sprintf(`
apiVersion: loki.grafana.com/v1
kind: LokiStack
metadata:
  name: %s
  namespace: %s
spec:
  size: 1x.demo
  storage:
    schemas:
    - version: v13
      effectiveDate: 2024-10-25
    secret:
      name: %s
      type: s3
  storageClassName: gp3-csi
  tenants:
    mode: openshift-logging
  rules:
    enabled: true
    selector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
    namespaceSelector:
      matchLabels:
        openshift.io/cluster-monitoring: "true"
  limits:
    global:
      ingestion:
        ingestionBurstSize: 10
        ingestionRate: 10
`, LokistackName, namespace, storageSecretName)

	uri := fmt.Sprintf(lokistackURI, namespace, LokistackName)

	// Create the garage storage secret in the lokistack namespace. Garage is deployed into the
	// same namespace as the lokistack, so the endpoint uses the in-namespace service DNS.
	clolog.V(1).Info("creating garage secret for lokistack", "namespace", namespace)
	data := map[string][]byte{
		"endpoint":          []byte(fmt.Sprintf("http://%s.%s.svc:3900", garageName, namespace)),
		"bucketnames":       []byte(garageBucket),
		"region":            []byte(garageRegion),
		"access_key_id":     []byte(garageAccessKeyID),
		"access_key_secret": []byte(garageAccessKeySecret),
	}
	storageSecret := runtime.NewSecret(namespace, storageSecretName, data)
	_, err = tc.KubeClient.CoreV1().Secrets(namespace).Create(context.TODO(), storageSecret, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	tc.AddCleanup(func() error {
		return tc.KubeClient.CoreV1().Secrets(namespace).Delete(context.TODO(), storageSecretName, metav1.DeleteOptions{})
	})

	if err = tc.Client().RESTClient().Post().
		RequestURI(uri).
		SetHeader("Content-Type", "application/yaml").
		Body([]byte(yaml)).
		Do(context.TODO()).Error(); err != nil {
		return nil, err
	}

	return logStore, tc.WaitForLokistackReady(namespace)
}

// ExternalURL returns the URL of the external route.
func (ls LokistackLogStore) ExternalURL(path string) (*url.URL, error) {
	route := openshiftv1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ls.Name,
			Namespace: ls.Namespace,
		},
	}
	if err := ls.tc.Test.Get(&route); err != nil {
		return nil, err
	}
	return &url.URL{Scheme: "https", Host: route.Spec.Host, Path: path}, nil
}

func (ls LokistackLogStore) Query(logQL string, orgID, tenant, saName string, limit int) ([]lokitesthelper.StreamValues, error) {
	u, err := ls.ExternalURL(fmt.Sprintf("/api/logs/v1/%s/loki/api/v1/query_range", tenant))
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Add("query", logQL)
	q.Add("limit", strconv.Itoa(limit))
	q.Add("direction", "FORWARD")
	u.RawQuery = q.Encode()
	clolog.V(3).Info("Loki Query", "url", u.String(), "org-id", orgID)
	header := http.Header{}
	if orgID != "" {
		header.Add("X-Scope-OrgID", orgID)
	}

	// Get SA token secret
	saTokenSecret, err := ls.tc.KubeClient.CoreV1().Secrets(ls.Namespace).Get(context.TODO(), saName+"-token", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if saTokenSecret != nil {
		if token, ok := saTokenSecret.Data[constants.TokenKey]; ok {
			header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
		}
	}

	req := &http.Request{
		Method: "GET",
		URL:    u,
		Header: header,
	}

	// Create client that skips certificate verification
	skipTLSClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := skipTLSClient.Do(req)
	if err == nil {
		err = test.HTTPError(resp)
	}

	if err != nil {
		clolog.V(3).Error(err, "Loki Query", "url", u.String())
		return nil, fmt.Errorf("%w\nURL: %v", err, u)
	}
	defer func() { testerrors.LogIfError(resp.Body.Close()) }()
	qr := lokitesthelper.QueryResponse{}
	if err = json.NewDecoder(resp.Body).Decode(&qr); err != nil {
		return nil, err
	}
	if qr.Status != "success" {
		return nil, fmt.Errorf("expected 'status: success' in %v", qr)
	}
	if qr.Data.ResultType != "streams" {
		return nil, fmt.Errorf("expected 'resultType: streams' in %v", qr)
	}
	clolog.V(3).Info("Loki Query done", "result", test.JSONString(qr.Data.Result))
	return qr.Data.Result, nil
}

// QueryUntil repeats the query until at least n lines are received.
func (ls LokistackLogStore) QueryUntil(logQL string, orgID, tenant, saName string, n int, timeToWait time.Duration) (values []lokitesthelper.StreamValues, err error) {
	clolog.V(2).Info("Loki QueryUntil", "query", logQL, "n", n)
	err = wait.PollUntilContextTimeout(context.TODO(), time.Second, timeToWait, true, func(cxt context.Context) (done bool, err error) {
		values, err = ls.Query(logQL, orgID, tenant, saName, n)
		if err != nil {
			return false, err
		}
		got := 0
		for _, v := range values {
			got += len(v.Values)
		}
		return got >= n, nil
	})
	return values, errors.Wrap(err, fmt.Sprintf("waiting for loki query %q, orgID %q", logQL, orgID))
}

func (ls LokistackLogStore) GetApplicationLogs(saName string, timeToWait time.Duration) ([]lokitesthelper.StreamValues, error) {
	query := fmt.Sprintf(`{log_type=%q}`, obs.InputTypeApplication)
	result, err := ls.QueryUntil(query, "", string(obs.InputTypeApplication), saName, 1, defaultTimeout)
	return result, errors.Wrap(err, "error getting application logs")
}

func (ls LokistackLogStore) GetApplicationLogsByKeyValue(saName, key, value string, timeToWait time.Duration) ([]lokitesthelper.StreamValues, error) {
	query := fmt.Sprintf(`{log_type=%q, %s=%q}`, obs.InputTypeApplication, key, value)
	result, err := ls.QueryUntil(query, "", string(obs.InputTypeApplication), saName, 1, defaultTimeout)
	return result, errors.Wrap(err, "error getting application logs with key, value")
}

func (ls LokistackLogStore) GetApplicationLogsWithPipeline(saName, expression string, timeToWait time.Duration) ([]lokitesthelper.StreamValues, error) {
	query := fmt.Sprintf(`{log_type=%q} %s`, obs.InputTypeApplication, expression)
	result, err := ls.QueryUntil(query, "", string(obs.InputTypeApplication), saName, 1, defaultTimeout)
	return result, errors.Wrap(err, "error getting application logs with log pipeline expression")
}

func (ls LokistackLogStore) HasApplicationLogs(saName string, timeToWait time.Duration) (bool, error) {
	query := fmt.Sprintf(`{log_type=%q}`, obs.InputTypeApplication)
	result, err := ls.QueryUntil(query, "", string(obs.InputTypeApplication), saName, 1, timeToWait)
	return len(result) > 0, errors.Wrap(err, "error determining if logstore has application logs")
}

func (ls LokistackLogStore) HasInfrastructureLogs(saName string, timeToWait time.Duration) (bool, error) {
	result, err := ls.InfrastructureLogs(saName, timeToWait, 1)
	return len(result) > 0, errors.Wrap(err, "error determining if logstore has infrastructure logs")
}

func (ls LokistackLogStore) InfrastructureLogs(saName string, timeToWait time.Duration, limit int) ([]lokitesthelper.StreamValues, error) {
	query := fmt.Sprintf(`{log_type=%q}`, obs.InputTypeInfrastructure)
	result, err := ls.QueryUntil(query, "", string(obs.InputTypeInfrastructure), saName, limit, timeToWait)
	clolog.V(3).Info("Loki Query", "result", test.JSONString(result))
	return result, errors.Wrap(err, "error determining if logstore has infrastructure logs")
}

func (ls LokistackLogStore) HasAuditLogs(saName string, timeToWait time.Duration) (bool, error) {
	query := fmt.Sprintf(`{log_type=%q}`, obs.InputTypeAudit)
	result, err := ls.QueryUntil(query, "", string(obs.InputTypeAudit), saName, 1, timeToWait)
	return len(result) > 0, errors.Wrap(err, "error determining if logstore has audit logs")
}

func (tc *E2ETestFramework) WaitForLokistackReady(namespace string) error {
	var err error

	// Wait stateful sets
	if err = tc.waitForStatefulSet(namespace, LokistackName+lokistackCompactor, defaultRetryInterval, defaultTimeout); err != nil {
		return errors.Wrap(err, "error waiting for lokistack-compactor")
	}
	if err = tc.waitForStatefulSet(namespace, LokistackName+lokistackIndexGateway, defaultRetryInterval, defaultTimeout); err != nil {
		return errors.Wrap(err, "error waiting for lokistack-index-gateway")
	}
	if err = tc.waitForStatefulSet(namespace, LokistackName+lokistackIngester, defaultRetryInterval, defaultTimeout); err != nil {
		return errors.Wrap(err, "error waiting for lokistack-ingester")
	}
	if err = tc.waitForStatefulSet(namespace, LokistackName+lokistackRuler, defaultRetryInterval, defaultTimeout); err != nil {
		return errors.Wrap(err, "error waiting for lokistack-ruler")
	}

	// Wait for deployments
	if err = tc.WaitForDeployment(namespace, LokistackName+lokistackDistributor, defaultRetryInterval, defaultTimeout); err != nil {
		return errors.Wrap(err, "error waiting for lokistack-distributor")
	}
	if err = tc.WaitForDeployment(namespace, LokistackName+lokistackGateway, defaultRetryInterval, defaultTimeout); err != nil {
		return errors.Wrap(err, "error waiting for lokistack-gateway")
	}
	if err = tc.WaitForDeployment(namespace, LokistackName+lokistackQuerier, defaultRetryInterval, defaultTimeout); err != nil {
		return errors.Wrap(err, "error waiting for lokistack-querier")
	}
	if err = tc.WaitForDeployment(namespace, LokistackName+lokistackQueryFrontend, defaultRetryInterval, defaultTimeout); err != nil {
		return errors.Wrap(err, "error waiting for lokistack-query-frontend")
	}

	return err
}
