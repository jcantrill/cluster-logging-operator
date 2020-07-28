package api

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type ApiGateway interface {

	Create(object runtime.Object) error
	Get(objectName string, object runtime.Object) error
	Update(object runtime.Object) (err error)
}