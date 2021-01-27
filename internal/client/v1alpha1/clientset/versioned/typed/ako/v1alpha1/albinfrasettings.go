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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/apis/ako/v1alpha1"
	scheme "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/client/v1alpha1/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// AlbInfraSettingsesGetter has a method to return a AlbInfraSettingsInterface.
// A group's client should implement this interface.
type AlbInfraSettingsesGetter interface {
	AlbInfraSettingses(namespace string) AlbInfraSettingsInterface
}

// AlbInfraSettingsInterface has methods to work with AlbInfraSettings resources.
type AlbInfraSettingsInterface interface {
	Create(ctx context.Context, albInfraSettings *v1alpha1.AlbInfraSettings, opts v1.CreateOptions) (*v1alpha1.AlbInfraSettings, error)
	Update(ctx context.Context, albInfraSettings *v1alpha1.AlbInfraSettings, opts v1.UpdateOptions) (*v1alpha1.AlbInfraSettings, error)
	UpdateStatus(ctx context.Context, albInfraSettings *v1alpha1.AlbInfraSettings, opts v1.UpdateOptions) (*v1alpha1.AlbInfraSettings, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.AlbInfraSettings, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.AlbInfraSettingsList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.AlbInfraSettings, err error)
	AlbInfraSettingsExpansion
}

// albInfraSettingses implements AlbInfraSettingsInterface
type albInfraSettingses struct {
	client rest.Interface
	ns     string
}

// newAlbInfraSettingses returns a AlbInfraSettingses
func newAlbInfraSettingses(c *AkoV1alpha1Client, namespace string) *albInfraSettingses {
	return &albInfraSettingses{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the albInfraSettings, and returns the corresponding albInfraSettings object, and an error if there is any.
func (c *albInfraSettingses) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.AlbInfraSettings, err error) {
	result = &v1alpha1.AlbInfraSettings{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("albinfrasettingses").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of AlbInfraSettingses that match those selectors.
func (c *albInfraSettingses) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.AlbInfraSettingsList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.AlbInfraSettingsList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("albinfrasettingses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested albInfraSettingses.
func (c *albInfraSettingses) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("albinfrasettingses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a albInfraSettings and creates it.  Returns the server's representation of the albInfraSettings, and an error, if there is any.
func (c *albInfraSettingses) Create(ctx context.Context, albInfraSettings *v1alpha1.AlbInfraSettings, opts v1.CreateOptions) (result *v1alpha1.AlbInfraSettings, err error) {
	result = &v1alpha1.AlbInfraSettings{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("albinfrasettingses").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(albInfraSettings).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a albInfraSettings and updates it. Returns the server's representation of the albInfraSettings, and an error, if there is any.
func (c *albInfraSettingses) Update(ctx context.Context, albInfraSettings *v1alpha1.AlbInfraSettings, opts v1.UpdateOptions) (result *v1alpha1.AlbInfraSettings, err error) {
	result = &v1alpha1.AlbInfraSettings{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("albinfrasettingses").
		Name(albInfraSettings.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(albInfraSettings).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *albInfraSettingses) UpdateStatus(ctx context.Context, albInfraSettings *v1alpha1.AlbInfraSettings, opts v1.UpdateOptions) (result *v1alpha1.AlbInfraSettings, err error) {
	result = &v1alpha1.AlbInfraSettings{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("albinfrasettingses").
		Name(albInfraSettings.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(albInfraSettings).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the albInfraSettings and deletes it. Returns an error if one occurs.
func (c *albInfraSettingses) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("albinfrasettingses").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *albInfraSettingses) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("albinfrasettingses").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched albInfraSettings.
func (c *albInfraSettingses) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.AlbInfraSettings, err error) {
	result = &v1alpha1.AlbInfraSettings{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("albinfrasettingses").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
