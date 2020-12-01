package logstore

import (
	"fmt"
	"os"

	operatorsv1 "github.com/operator-framework/api/pkg/operators/v1"
	operatorsv1alpha1 "github.com/operator-framework/api/pkg/operators/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	elasticsearchOperatorName = "elasticsearch-operator"
)

type fnCreate func(runtime.Object) error

//CreateOperator creates the OLM objects to deploy the elasticsearch-opeartor
//Note: these objects are not reconciled and do not have ownerref set to ensure
//the operator is not removed if being used by other parties
func CreateOperator(create fnCreate) (err error) {
	ns := os.Getenv("LOGSTORE_OPERATOR_NS")
	if ns == "" {
		return fmt.Errorf("LOGSTORE_OPERATOR_NS is empty. Unable to deploy operator for the logstore")
	}
	//create operator group
	og := newOperatorGroup(ns)
	if err = createObject(og, create); err != nil {
		return err
	}
	//create subscription
	subscription, err := newSubscription(ns)
	if err != nil {
		return err
	}
	return createObject(subscription, create)
}

func newSubscription(ns string) (*operatorsv1alpha1.Subscription, error) {
	channel := os.Getenv("OPERATOR_PACKAGE_CHANNEL")
	if channel == "" {
		return nil, fmt.Errorf("OPERATOR_PACKAGE_CHANNEL is empty. Unable to create subscription for the logstore")
	}
	source := os.Getenv("OPERATOR_CATALOG_SOURCE")
	if channel == "" {
		return nil, fmt.Errorf("OPERATOR_CATALOG_SOURCE is empty. Unable to create subscription for the logstore")
	}
	return &operatorsv1alpha1.Subscription{
		TypeMeta: metav1.TypeMeta{
			Kind:       operatorsv1alpha1.SubscriptionKind,
			APIVersion: operatorsv1alpha1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      elasticsearchOperatorName,
			Namespace: ns,
		},
		Spec: &operatorsv1alpha1.SubscriptionSpec{
			Channel:                channel,
			CatalogSource:          source,
			CatalogSourceNamespace: ns,
			Package:                elasticsearchOperatorName,
		},
	}, nil
}

func newOperatorGroup(ns string) *operatorsv1.OperatorGroup {

	return &operatorsv1.OperatorGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       operatorsv1.OperatorGroupKind,
			APIVersion: operatorsv1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      elasticsearchOperatorName,
			Namespace: ns,
		},
	}
}
func createObject(obj runtime.Object, create fnCreate) error {
	if err := create(obj); err != nil {
		if !errors.IsAlreadyExists(err) {
			return fmt.Errorf("Failure creating kind %s while trying to deploy logstore operator: %v", obj.GetObjectKind(), err)
		}
	}
	return nil
}
