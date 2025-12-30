/*
Copyright 2024 Keiailab.

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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	mongodbv1alpha1 "github.com/keiailab/mongodb-operator/api/v1alpha1"
)

var _ = Describe("MongoDBSharded Controller", func() {
	const (
		ShardedName      = "test-sharded"
		ShardedNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a MongoDBSharded resource", func() {
		It("Should create the sharded cluster resources", func() {
			By("Creating a new MongoDBSharded")
			ctx := context.Background()
			sharded := &mongodbv1alpha1.MongoDBSharded{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "mongodb.keiailab.github.io/v1alpha1",
					Kind:       "MongoDBSharded",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      ShardedName,
					Namespace: ShardedNamespace,
				},
				Spec: mongodbv1alpha1.MongoDBShardedSpec{
					Version: mongodbv1alpha1.MongoDBVersion{
						Version: "7.0",
					},
					ConfigServer: mongodbv1alpha1.ConfigServerSpec{
						Members: 3,
						Storage: mongodbv1alpha1.StorageSpec{
							Size: resource.MustParse("10Gi"),
						},
					},
					Shards: mongodbv1alpha1.ShardSpec{
						Count:           2,
						MembersPerShard: 3,
						Storage: mongodbv1alpha1.StorageSpec{
							Size: resource.MustParse("50Gi"),
						},
					},
					Mongos: mongodbv1alpha1.MongosSpec{
						Replicas: 2,
					},
				},
			}
			Expect(k8sClient.Create(ctx, sharded)).Should(Succeed())

			shardedLookupKey := types.NamespacedName{Name: ShardedName, Namespace: ShardedNamespace}
			createdSharded := &mongodbv1alpha1.MongoDBSharded{}

			// Verify MongoDBSharded resource is created
			Eventually(func() bool {
				err := k8sClient.Get(ctx, shardedLookupKey, createdSharded)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSharded.Spec.ConfigServer.Members).Should(Equal(int32(3)))
			Expect(createdSharded.Spec.Shards.Count).Should(Equal(int32(2)))
			Expect(createdSharded.Spec.Shards.MembersPerShard).Should(Equal(int32(3)))
			Expect(createdSharded.Spec.Mongos.Replicas).Should(Equal(int32(2)))

			// Clean up
			By("Cleaning up the MongoDBSharded resource")
			Expect(k8sClient.Delete(ctx, sharded)).Should(Succeed())
		})
	})

	Context("When validating MongoDBSharded spec", func() {
		It("Should accept valid shard configuration", func() {
			ctx := context.Background()
			sharded := &mongodbv1alpha1.MongoDBSharded{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "mongodb.keiailab.github.io/v1alpha1",
					Kind:       "MongoDBSharded",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-valid-shards",
					Namespace: ShardedNamespace,
				},
				Spec: mongodbv1alpha1.MongoDBShardedSpec{
					Version: mongodbv1alpha1.MongoDBVersion{
						Version: "7.0",
					},
					ConfigServer: mongodbv1alpha1.ConfigServerSpec{
						Members: 3,
						Storage: mongodbv1alpha1.StorageSpec{
							Size: resource.MustParse("10Gi"),
						},
					},
					Shards: mongodbv1alpha1.ShardSpec{
						Count:           4,
						MembersPerShard: 3,
						Storage: mongodbv1alpha1.StorageSpec{
							Size: resource.MustParse("100Gi"),
						},
					},
					Mongos: mongodbv1alpha1.MongosSpec{
						Replicas: 3,
					},
				},
			}
			Expect(k8sClient.Create(ctx, sharded)).Should(Succeed())

			// Clean up
			Expect(k8sClient.Delete(ctx, sharded)).Should(Succeed())
		})

		It("Should accept custom storage class", func() {
			ctx := context.Background()
			sharded := &mongodbv1alpha1.MongoDBSharded{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "mongodb.keiailab.github.io/v1alpha1",
					Kind:       "MongoDBSharded",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-storage-class",
					Namespace: ShardedNamespace,
				},
				Spec: mongodbv1alpha1.MongoDBShardedSpec{
					Version: mongodbv1alpha1.MongoDBVersion{
						Version: "7.0",
					},
					ConfigServer: mongodbv1alpha1.ConfigServerSpec{
						Members: 3,
						Storage: mongodbv1alpha1.StorageSpec{
							Size:             resource.MustParse("10Gi"),
							StorageClassName: "fast-ssd",
						},
					},
					Shards: mongodbv1alpha1.ShardSpec{
						Count:           2,
						MembersPerShard: 3,
						Storage: mongodbv1alpha1.StorageSpec{
							Size:             resource.MustParse("50Gi"),
							StorageClassName: "fast-ssd",
						},
					},
					Mongos: mongodbv1alpha1.MongosSpec{
						Replicas: 2,
					},
				},
			}
			Expect(k8sClient.Create(ctx, sharded)).Should(Succeed())

			// Verify storage class is set
			shardedLookupKey := types.NamespacedName{Name: "test-storage-class", Namespace: ShardedNamespace}
			createdSharded := &mongodbv1alpha1.MongoDBSharded{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, shardedLookupKey, createdSharded)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdSharded.Spec.ConfigServer.Storage.StorageClassName).Should(Equal("fast-ssd"))
			Expect(createdSharded.Spec.Shards.Storage.StorageClassName).Should(Equal("fast-ssd"))

			// Clean up
			Expect(k8sClient.Delete(ctx, sharded)).Should(Succeed())
		})
	})

	Context("When updating MongoDBSharded resources", func() {
		It("Should update the mongos replicas", func() {
			ctx := context.Background()
			sharded := &mongodbv1alpha1.MongoDBSharded{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "mongodb.keiailab.github.io/v1alpha1",
					Kind:       "MongoDBSharded",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-update-sharded",
					Namespace: ShardedNamespace,
				},
				Spec: mongodbv1alpha1.MongoDBShardedSpec{
					Version: mongodbv1alpha1.MongoDBVersion{
						Version: "7.0",
					},
					ConfigServer: mongodbv1alpha1.ConfigServerSpec{
						Members: 3,
						Storage: mongodbv1alpha1.StorageSpec{
							Size: resource.MustParse("10Gi"),
						},
					},
					Shards: mongodbv1alpha1.ShardSpec{
						Count:           2,
						MembersPerShard: 3,
						Storage: mongodbv1alpha1.StorageSpec{
							Size: resource.MustParse("50Gi"),
						},
					},
					Mongos: mongodbv1alpha1.MongosSpec{
						Replicas: 2,
					},
				},
			}
			Expect(k8sClient.Create(ctx, sharded)).Should(Succeed())

			// Update the MongoDBSharded
			shardedLookupKey := types.NamespacedName{Name: "test-update-sharded", Namespace: ShardedNamespace}
			createdSharded := &mongodbv1alpha1.MongoDBSharded{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, shardedLookupKey, createdSharded)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update mongos replicas
			createdSharded.Spec.Mongos.Replicas = 4
			Expect(k8sClient.Update(ctx, createdSharded)).Should(Succeed())

			// Verify the update
			updatedSharded := &mongodbv1alpha1.MongoDBSharded{}
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, shardedLookupKey, updatedSharded)
				if err != nil {
					return 0
				}
				return updatedSharded.Spec.Mongos.Replicas
			}, timeout, interval).Should(Equal(int32(4)))

			// Clean up
			Expect(k8sClient.Delete(ctx, sharded)).Should(Succeed())
		})
	})

	Context("When deleting MongoDBSharded resources", func() {
		It("Should delete the MongoDBSharded resource", func() {
			ctx := context.Background()
			sharded := &mongodbv1alpha1.MongoDBSharded{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "mongodb.keiailab.github.io/v1alpha1",
					Kind:       "MongoDBSharded",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-delete-sharded",
					Namespace: ShardedNamespace,
				},
				Spec: mongodbv1alpha1.MongoDBShardedSpec{
					Version: mongodbv1alpha1.MongoDBVersion{
						Version: "7.0",
					},
					ConfigServer: mongodbv1alpha1.ConfigServerSpec{
						Members: 3,
						Storage: mongodbv1alpha1.StorageSpec{
							Size: resource.MustParse("10Gi"),
						},
					},
					Shards: mongodbv1alpha1.ShardSpec{
						Count:           2,
						MembersPerShard: 3,
						Storage: mongodbv1alpha1.StorageSpec{
							Size: resource.MustParse("50Gi"),
						},
					},
					Mongos: mongodbv1alpha1.MongosSpec{
						Replicas: 2,
					},
				},
			}
			Expect(k8sClient.Create(ctx, sharded)).Should(Succeed())

			// Delete the MongoDBSharded
			Expect(k8sClient.Delete(ctx, sharded)).Should(Succeed())

			// Verify deletion
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{Name: "test-delete-sharded", Namespace: ShardedNamespace}, &mongodbv1alpha1.MongoDBSharded{})
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})
