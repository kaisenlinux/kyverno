package resolvers

import (
	"context"
	"reflect"
	"testing"
	"time"

	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeinformers "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	kubefake "k8s.io/client-go/kubernetes/fake"
	corev1listers "k8s.io/client-go/listers/core/v1"
)

func newEmptyFakeClient() *kubefake.Clientset {
	return kubefake.NewSimpleClientset()
}

func createConfigMaps(ctx context.Context, client *kubefake.Clientset, addLabel bool) error {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      TEST_CONFIGMAP,
			Namespace: TEST_NAMESPACE,
		},
		Data: map[string]string{"configmapkey": "key1"},
	}
	if addLabel {
		cm.ObjectMeta.Labels = map[string]string{LabelCacheKey: "true"}
	}
	_, err := client.CoreV1().ConfigMaps(TEST_NAMESPACE).Create(
		ctx, cm, metav1.CreateOptions{})
	return err
}

func initialiseInformer(client *kubefake.Clientset) kubeinformers.SharedInformerFactory {
	selector, err := GetCacheSelector()
	if err != nil {
		return nil
	}
	labelOptions := kubeinformers.WithTweakListOptions(func(opts *metav1.ListOptions) {
		opts.LabelSelector = selector.String()
	})
	kubeResourceInformer := kubeinformers.NewSharedInformerFactoryWithOptions(client, 15*time.Minute, labelOptions)
	return kubeResourceInformer
}

func Test_InformerCacheSuccess(t *testing.T) {
	client := newEmptyFakeClient()
	ctx := context.TODO()
	err := createConfigMaps(ctx, client, true)
	assert.NilError(t, err, "error while creating configmap")
	informer := initialiseInformer(client)
	informerResolver, err := NewInformerBasedResolver(informer.Core().V1().ConfigMaps().Lister())
	assert.NilError(t, err)
	informer.Start(make(<-chan struct{}))
	time.Sleep(10 * time.Second)
	_, err = informerResolver.Get(ctx, TEST_NAMESPACE, TEST_CONFIGMAP)
	assert.NilError(t, err, "informer didn't have expected configmap")
}

func Test_InformerCacheFailure(t *testing.T) {
	client := newEmptyFakeClient()
	ctx := context.TODO()
	err := createConfigMaps(ctx, client, false)
	assert.NilError(t, err, "error while creating configmap")
	informer := initialiseInformer(client)
	resolver, err := NewInformerBasedResolver(informer.Core().V1().ConfigMaps().Lister())
	assert.NilError(t, err)
	informer.Start(make(<-chan struct{}))
	time.Sleep(10 * time.Second)
	_, err = resolver.Get(ctx, TEST_NAMESPACE, TEST_CONFIGMAP)
	assert.Equal(t, err.Error(), "configmap \"myconfigmap\" not found")
}

func Test_ClientBasedResolver(t *testing.T) {
	client := newEmptyFakeClient()
	ctx := context.TODO()
	err := createConfigMaps(ctx, client, false)
	assert.NilError(t, err, "error while creating configmap")
	resolver, err := NewClientBasedResolver(client)
	assert.NilError(t, err)
	_, err = resolver.Get(ctx, TEST_NAMESPACE, TEST_CONFIGMAP)
	assert.NilError(t, err, "error while getting configmap from client")
}

func Test_ResolverChainWithExistingConfigMap(t *testing.T) {
	client := newEmptyFakeClient()
	informer := initialiseInformer(client)
	lister := informer.Core().V1().ConfigMaps().Lister()
	informerBasedResolver, err := NewInformerBasedResolver(lister)
	assert.NilError(t, err)
	clientBasedResolver, err := NewClientBasedResolver(client)
	assert.NilError(t, err)
	resolvers, err := NewResolverChain(informerBasedResolver, clientBasedResolver)
	assert.NilError(t, err)
	ctx := context.TODO()
	err = createConfigMaps(ctx, client, true)
	assert.NilError(t, err, "error while creating configmap")
	_, err = resolvers.Get(ctx, TEST_NAMESPACE, TEST_CONFIGMAP)
	assert.NilError(t, err, "error while getting configmap")
}

func Test_ResolverChainWithNonExistingConfigMap(t *testing.T) {
	client := newEmptyFakeClient()
	informer := initialiseInformer(client)
	lister := informer.Core().V1().ConfigMaps().Lister()
	informerBasedResolver, err := NewInformerBasedResolver(lister)
	assert.NilError(t, err)
	clientBasedResolver, err := NewClientBasedResolver(client)
	assert.NilError(t, err)
	resolvers, err := NewResolverChain(informerBasedResolver, clientBasedResolver)
	assert.NilError(t, err)
	ctx := context.TODO()
	_, err = resolvers.Get(ctx, TEST_NAMESPACE, TEST_CONFIGMAP)
	assert.Error(t, err, "configmaps \"myconfigmap\" not found")
}

func TestNewInformerBasedResolver(t *testing.T) {
	type args struct {
		lister corev1listers.ConfigMapLister
	}
	client := newEmptyFakeClient()
	informer := initialiseInformer(client)
	lister := informer.Core().V1().ConfigMaps().Lister()
	tests := []struct {
		name    string
		args    args
		want    ConfigmapResolver
		wantErr bool
	}{{
		name:    "nil shoud return an error",
		wantErr: true,
	}, {
		name: "not nil",
		args: args{lister},
		want: &informerBasedResolver{lister},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewInformerBasedResolver(tt.args.lister)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewInformerBasedResolver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewInformerBasedResolver() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewClientBasedResolver(t *testing.T) {
	type args struct {
		client kubernetes.Interface
	}
	client := newEmptyFakeClient()
	tests := []struct {
		name    string
		args    args
		want    ConfigmapResolver
		wantErr bool
	}{{
		name:    "nil shoud return an error",
		wantErr: true,
	}, {
		name: "not nil",
		args: args{client},
		want: &clientBasedResolver{client},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClientBasedResolver(tt.args.client)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClientBasedResolver() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewClientBasedResolver() = %v, want %v", got, tt.want)
			}
		})
	}
}

type dummyResolver struct{}

func (c dummyResolver) Get(context.Context, string, string) (*corev1.ConfigMap, error) {
	return nil, nil
}

func TestNewResolverChain(t *testing.T) {
	type args struct {
		resolvers []ConfigmapResolver
	}
	tests := []struct {
		name    string
		args    args
		want    ConfigmapResolver
		wantErr bool
	}{{
		name:    "nil shoud return an error",
		wantErr: true,
	}, {
		name:    "empty list shoud return an error",
		args:    args{[]ConfigmapResolver{}},
		wantErr: true,
	}, {
		name:    "one nil in the list shoud return an error",
		args:    args{[]ConfigmapResolver{dummyResolver{}, nil}},
		wantErr: true,
	}, {
		name: "no nil",
		args: args{[]ConfigmapResolver{dummyResolver{}, dummyResolver{}, dummyResolver{}}},
		want: resolverChain{dummyResolver{}, dummyResolver{}, dummyResolver{}},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewResolverChain(tt.args.resolvers...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewResolverChain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewResolverChain() = %v, want %v", got, tt.want)
			}
		})
	}
}
