package handlers

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/kyverno/kyverno/pkg/config"
	"github.com/kyverno/kyverno/pkg/userinfo"
	admissionutils "github.com/kyverno/kyverno/pkg/utils/admission"
	kubeutils "github.com/kyverno/kyverno/pkg/utils/kube"
	admissionv1 "k8s.io/api/admission/v1"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	rbacv1listers "k8s.io/client-go/listers/rbac/v1"
)

func (inner AdmissionHandler) WithDump(
	enabled bool,
	rbLister rbacv1listers.RoleBindingLister,
	crbLister rbacv1listers.ClusterRoleBindingLister,
	configuration config.Configuration,
) AdmissionHandler {
	if !enabled {
		return inner
	}
	return inner.withDump(rbLister, crbLister, configuration).WithTrace("DUMP")
}

func (inner AdmissionHandler) withDump(
	rbLister rbacv1listers.RoleBindingLister,
	crbLister rbacv1listers.ClusterRoleBindingLister,
	configuration config.Configuration,
) AdmissionHandler {
	return func(ctx context.Context, logger logr.Logger, request *admissionv1.AdmissionRequest, startTime time.Time) *admissionv1.AdmissionResponse {
		response := inner(ctx, logger, request, startTime)
		dumpPayload(logger, rbLister, crbLister, configuration, request, response)
		return response
	}
}

func dumpPayload(
	logger logr.Logger,
	rbLister rbacv1listers.RoleBindingLister,
	crbLister rbacv1listers.ClusterRoleBindingLister,
	configuration config.Configuration,
	request *admissionv1.AdmissionRequest,
	response *admissionv1.AdmissionResponse,
) {
	reqPayload, err := newAdmissionRequestPayload(request, rbLister, crbLister, configuration)
	if err != nil {
		logger.Error(err, "Failed to extract resources")
	} else {
		if response != nil {
			logger = logger.WithValues("AdmissionResponse", *response)
		}
		logger.Info("Logging admission request and response payload ", "AdmissionRequest", reqPayload)
	}
}

// admissionRequestPayload holds a copy of the AdmissionRequest payload
type admissionRequestPayload struct {
	UID                types.UID                    `json:"uid"`
	Kind               metav1.GroupVersionKind      `json:"kind"`
	Resource           metav1.GroupVersionResource  `json:"resource"`
	SubResource        string                       `json:"subResource,omitempty"`
	RequestKind        *metav1.GroupVersionKind     `json:"requestKind,omitempty"`
	RequestResource    *metav1.GroupVersionResource `json:"requestResource,omitempty"`
	RequestSubResource string                       `json:"requestSubResource,omitempty"`
	Name               string                       `json:"name,omitempty"`
	Namespace          string                       `json:"namespace,omitempty"`
	Operation          string                       `json:"operation"`
	UserInfo           authenticationv1.UserInfo    `json:"userInfo"`
	Roles              []string                     `json:"roles"`
	ClusterRoles       []string                     `json:"clusterRoles"`
	Object             unstructured.Unstructured    `json:"object,omitempty"`
	OldObject          unstructured.Unstructured    `json:"oldObject,omitempty"`
	DryRun             *bool                        `json:"dryRun,omitempty"`
	Options            unstructured.Unstructured    `json:"options,omitempty"`
}

func newAdmissionRequestPayload(request *admissionv1.AdmissionRequest, rbLister rbacv1listers.RoleBindingLister, crbLister rbacv1listers.ClusterRoleBindingLister, configuration config.Configuration) (*admissionRequestPayload, error) {
	newResource, oldResource, err := admissionutils.ExtractResources(nil, request)
	if err != nil {
		return nil, err
	}
	options := new(unstructured.Unstructured)
	if request.Options.Raw != nil {
		options, err = kubeutils.BytesToUnstructured(request.Options.Raw)
		if err != nil {
			return nil, err
		}
	}
	var roles, clusterRoles []string
	if rbLister != nil && crbLister != nil && configuration != nil {
		if r, cr, err := userinfo.GetRoleRef(rbLister, crbLister, request, configuration); err != nil {
			return nil, err
		} else {
			roles = r
			clusterRoles = cr
		}
	}
	return redactPayload(&admissionRequestPayload{
		UID:                request.UID,
		Kind:               request.Kind,
		Resource:           request.Resource,
		SubResource:        request.SubResource,
		RequestKind:        request.RequestKind,
		RequestResource:    request.RequestResource,
		RequestSubResource: request.RequestSubResource,
		Name:               request.Name,
		Namespace:          request.Namespace,
		Operation:          string(request.Operation),
		UserInfo:           request.UserInfo,
		Roles:              roles,
		ClusterRoles:       clusterRoles,
		Object:             newResource,
		OldObject:          oldResource,
		DryRun:             request.DryRun,
		Options:            *options,
	})
}

func redactPayload(payload *admissionRequestPayload) (*admissionRequestPayload, error) {
	if strings.EqualFold(payload.Kind.Kind, "Secret") {
		if payload.Object.Object != nil {
			obj, err := kubeutils.RedactSecret(&payload.Object)
			if err != nil {
				return nil, err
			}
			payload.Object = obj
		}
		if payload.OldObject.Object != nil {
			oldObj, err := kubeutils.RedactSecret(&payload.OldObject)
			if err != nil {
				return nil, err
			}
			payload.OldObject = oldObj
		}
	}
	return payload, nil
}
