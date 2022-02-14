/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1

import (
	v12 "github.com/kyverno/kyverno/api/kyverno/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// GenerateRequestLister helps list GenerateRequests.
// All objects returned here must be treated as read-only.
type GenerateRequestLister interface {
	// List lists all GenerateRequests in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v12.GenerateRequest, err error)
	// GenerateRequests returns an object that can list and get GenerateRequests.
	GenerateRequests(namespace string) GenerateRequestNamespaceLister
	GenerateRequestListerExpansion
}

// generateRequestLister implements the GenerateRequestLister interface.
type generateRequestLister struct {
	indexer cache.Indexer
}

// NewGenerateRequestLister returns a new GenerateRequestLister.
func NewGenerateRequestLister(indexer cache.Indexer) GenerateRequestLister {
	return &generateRequestLister{indexer: indexer}
}

// List lists all GenerateRequests in the indexer.
func (s *generateRequestLister) List(selector labels.Selector) (ret []*v12.GenerateRequest, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v12.GenerateRequest))
	})
	return ret, err
}

// GenerateRequests returns an object that can list and get GenerateRequests.
func (s *generateRequestLister) GenerateRequests(namespace string) GenerateRequestNamespaceLister {
	return generateRequestNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// GenerateRequestNamespaceLister helps list and get GenerateRequests.
// All objects returned here must be treated as read-only.
type GenerateRequestNamespaceLister interface {
	// List lists all GenerateRequests in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v12.GenerateRequest, err error)
	// Get retrieves the GenerateRequest from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v12.GenerateRequest, error)
	GenerateRequestNamespaceListerExpansion
}

// generateRequestNamespaceLister implements the GenerateRequestNamespaceLister
// interface.
type generateRequestNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all GenerateRequests in the indexer for a given namespace.
func (s generateRequestNamespaceLister) List(selector labels.Selector) (ret []*v12.GenerateRequest, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v12.GenerateRequest))
	})
	return ret, err
}

// Get retrieves the GenerateRequest from the indexer for a given namespace and name.
func (s generateRequestNamespaceLister) Get(name string) (*v12.GenerateRequest, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v12.Resource("generaterequest"), name)
	}
	return obj.(*v12.GenerateRequest), nil
}
