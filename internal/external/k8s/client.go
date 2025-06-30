package k8s

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Client is a kubernetes client interface used internally. It copies functions from
// sigs.k8s.io/controller-runtime/pkg/client
//
//counterfeiter:generate . Client
type Client interface {
	Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error
	Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error

	Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error
	Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error
	DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error
	List(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) error
	Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error

	RESTMapper() meta.RESTMapper
	Scheme() *runtime.Scheme

	GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error)
	IsObjectNamespaced(obj runtime.Object) (bool, error)
	Status() client.StatusWriter
	SubResource(subResource string) client.SubResourceClient
}

// StatusWriter is a kubernetes status writer interface used internally. It copies functions from
// sigs.k8s.io/controller-runtime/pkg/client
//
//counterfeiter:generate . StatusWriter
type StatusWriter interface {
	Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error
	Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error
	Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error
}

// SubResourceClient is a kubernetes status writer interface used internally. It copies functions from
// sigs.k8s.io/controller-runtime/pkg/client
//
//counterfeiter:generate . SubResourceClient
type SubResourceClient interface {
	Get(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceGetOption) error

	Create(ctx context.Context, obj client.Object, subResource client.Object, opts ...client.SubResourceCreateOption) error
	Update(ctx context.Context, obj client.Object, opts ...client.SubResourceUpdateOption) error
	Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.SubResourcePatchOption) error
}

// k8sClient is a wrapper around controller-runtime client that implements our Client interface
type k8sClient struct {
	client client.Client
}

// NewClient creates a new k8s.Client from a controller-runtime client
func NewClient(c client.Client) Client {
	return &k8sClient{client: c}
}

// Create implements Client.Create
func (k *k8sClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return k.client.Create(ctx, obj, opts...)
}

// Get implements Client.Get
func (k *k8sClient) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return k.client.Get(ctx, key, obj, opts...)
}

// Update implements Client.Update
func (k *k8sClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return k.client.Update(ctx, obj, opts...)
}

// Delete implements Client.Delete
func (k *k8sClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return k.client.Delete(ctx, obj, opts...)
}

// DeleteAllOf implements Client.DeleteAllOf
func (k *k8sClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return k.client.DeleteAllOf(ctx, obj, opts...)
}

// List implements Client.List
func (k *k8sClient) List(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) error {
	return k.client.List(ctx, obj, opts...)
}

// Patch implements Client.Patch
func (k *k8sClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return k.client.Patch(ctx, obj, patch, opts...)
}

// RESTMapper implements Client.RESTMapper
func (k *k8sClient) RESTMapper() meta.RESTMapper {
	return k.client.RESTMapper()
}

// Scheme implements Client.Scheme
func (k *k8sClient) Scheme() *runtime.Scheme {
	return k.client.Scheme()
}

// GroupVersionKindFor implements Client.GroupVersionKindFor
func (k *k8sClient) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return k.client.GroupVersionKindFor(obj)
}

// IsObjectNamespaced implements Client.IsObjectNamespaced
func (k *k8sClient) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return k.client.IsObjectNamespaced(obj)
}

// Status implements Client.Status
func (k *k8sClient) Status() client.StatusWriter {
	return k.client.Status()
}

// SubResource implements Client.SubResource
func (k *k8sClient) SubResource(subResource string) client.SubResourceClient {
	return k.client.SubResource(subResource)
}
