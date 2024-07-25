/*
Copyright 2024.

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

package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	artifactv1 "github.com/openfluxcd/artifact/api/v1alpha1"
	"github.com/openfluxcd/controller-manager/storage"
	"github.com/openfluxcd/http-source-controller/api/v1alpha1"
	"github.com/openfluxcd/http-source-controller/internal/fetcher"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestHttpReconciler_Reconcile(t *testing.T) {
	tmp, err := os.MkdirTemp("", "test-reconcile")
	require.NoError(t, err)
	defer os.RemoveAll(tmp)

	type fields struct {
		Content       string
		Client        func(url string) client.Client
		Scheme        *runtime.Scheme
		Fetcher       func(client *http.Client) *fetcher.Fetcher
		Storage       *storage.Storage
		AssertErr     func(t *testing.T, err error)
		AssertObjects func(t *testing.T, client client.Client)
	}
	type args struct {
		ctx context.Context
		req controllerruntime.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "should return no error if the request is valid",
			fields: fields{
				Client: func(url string) client.Client {
					return env.FakeKubeClient(
						WithObjects(&v1alpha1.Http{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-http",
								Namespace: "default",
							},
							Spec: v1alpha1.HttpSpec{
								URL: url,
							},
						}))
				},
				Content: "test-content",
				Scheme:  env.scheme,
				Fetcher: func(client *http.Client) *fetcher.Fetcher { return fetcher.NewFetcher(client) },
				Storage: &storage.Storage{
					BasePath: tmp,
					Hostname: "hostname",
				},
				AssertErr: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
				AssertObjects: func(t *testing.T, client client.Client) {
					artifact := &artifactv1.Artifact{}
					err = client.Get(context.TODO(), types.NamespacedName{Name: "test-http", Namespace: "default"}, artifact)
					require.NoError(t, err)
					// <kind>/<namespace>/name>/<filename>
					// The base name must not be there because the file server already adds that.
					assert.Equal(t, "http://hostname/http/default/test-http/0a3666a0710c08aa6d0de92ce72beeb5b93124cce1bf3701c9d6cdeb543cb73e.tar.gz", artifact.Spec.URL)
					assert.Equal(t, "0a3666a0710c08aa6d0de92ce72beeb5b93124cce1bf3701c9d6cdeb543cb73e", artifact.Spec.Revision)
					assert.Equal(t, int64(128), *artifact.Spec.Size)
				},
			},
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-http",
						Namespace: "default",
					},
				},
			},
		},
		{
			name: "should update existing artifact",
			fields: fields{
				Content: "test-content-2",
				Client: func(url string) client.Client {
					return env.FakeKubeClient(
						WithObjects(&v1alpha1.Http{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-http-2",
								Namespace: "default",
							},
							Spec: v1alpha1.HttpSpec{
								URL: url,
							},
						}, &artifactv1.Artifact{
							ObjectMeta: metav1.ObjectMeta{
								Name:      "test-http-2",
								Namespace: "default",
							},
							Spec: artifactv1.ArtifactSpec{
								URL:            "http://hostname/http/default/test-http-2/0a3666a0710c08aa6d0de92ce72beeb5b93124cce1bf3701c9d6cdeb543cb73e.tar.gz",
								Revision:       "0a3666a0710c08aa6d0de92ce72beeb5b93124cce1bf3701c9d6cdeb543cb73e",
								Digest:         "0a3666a0710c08aa6d0de92ce72beeb5b93124cce1bf3701c9d6cdeb543cb73e",
								LastUpdateTime: metav1.Now(),
								Size:           ptr.To[int64](128),
							},
						}))
				},
				Scheme:  env.scheme,
				Fetcher: func(client *http.Client) *fetcher.Fetcher { return fetcher.NewFetcher(client) },
				Storage: &storage.Storage{
					BasePath: tmp,
					Hostname: "hostname",
				},
				AssertErr: func(t *testing.T, err error) {
					require.NoError(t, err)
				},
				AssertObjects: func(t *testing.T, client client.Client) {
					artifact := &artifactv1.Artifact{}
					err = client.Get(context.TODO(), types.NamespacedName{Name: "test-http-2", Namespace: "default"}, artifact)
					require.NoError(t, err)
					// <kind>/<namespace>/name>/<filename>
					// The base name must not be there because the file server already adds that.
					assert.Equal(t, "http://hostname/http/default/test-http-2/d1194ac44b3f0ec54854a5611f0b49d715f133f1bfa454d6de4dc1aa83b67e89.tar.gz", artifact.Spec.URL)
					assert.Equal(t, "d1194ac44b3f0ec54854a5611f0b49d715f133f1bfa454d6de4dc1aa83b67e89", artifact.Spec.Revision)
				},
			},
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{
					NamespacedName: types.NamespacedName{
						Name:      "test-http-2",
						Namespace: "default",
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testserver := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(tt.fields.Content))
			}))
			defer testserver.Close()

			c := tt.fields.Client(testserver.URL)
			r := &HttpReconciler{
				Client:  c,
				Scheme:  tt.fields.Scheme,
				Fetcher: tt.fields.Fetcher(testserver.Client()),
				Storage: tt.fields.Storage,
			}
			_, err := r.Reconcile(tt.args.ctx, tt.args.req)
			tt.fields.AssertErr(t, err)
			tt.fields.AssertObjects(t, c)
		})
	}
}
