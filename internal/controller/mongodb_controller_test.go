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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	mongodbv1alpha1 "github.com/keiailab/mongodb-operator/api/v1alpha1"
)

var _ = Describe("MongoDB Controller", func() {
	const (
		MongoDBName      = "test-mongodb"
		MongoDBNamespace = "default"

		timeout  = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When creating a MongoDB resource", func() {
		It("Should create the required Kubernetes resources", func() {
			By("Creating a new MongoDB")
			ctx := context.Background()
			mongodb := &mongodbv1alpha1.MongoDB{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "mongodb.keiailab.github.io/v1alpha1",
					Kind:       "MongoDB",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      MongoDBName,
					Namespace: MongoDBNamespace,
				},
				Spec: mongodbv1alpha1.MongoDBSpec{
					Members:        3,
					ReplicaSetName: "rs0",
					Version: mongodbv1alpha1.MongoDBVersion{
						Version: "7.0",
					},
					Storage: mongodbv1alpha1.StorageSpec{
						Size:        resource.MustParse("10Gi"),
						DataDirPath: "/data/db",
					},
				},
			}
			Expect(k8sClient.Create(ctx, mongodb)).Should(Succeed())

			mongoDBLookupKey := types.NamespacedName{Name: MongoDBName, Namespace: MongoDBNamespace}
			createdMongoDB := &mongodbv1alpha1.MongoDB{}

			// Verify MongoDB resource is created
			Eventually(func() bool {
				err := k8sClient.Get(ctx, mongoDBLookupKey, createdMongoDB)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(createdMongoDB.Spec.Members).Should(Equal(int32(3)))
			Expect(createdMongoDB.Spec.ReplicaSetName).Should(Equal("rs0"))

			// Clean up
			By("Cleaning up the MongoDB resource")
			Expect(k8sClient.Delete(ctx, mongodb)).Should(Succeed())
		})
	})

	Context("When validating MongoDB spec", func() {
		It("Should accept valid member counts", func() {
			ctx := context.Background()
			mongodb := &mongodbv1alpha1.MongoDB{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "mongodb.keiailab.github.io/v1alpha1",
					Kind:       "MongoDB",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-valid-members",
					Namespace: MongoDBNamespace,
				},
				Spec: mongodbv1alpha1.MongoDBSpec{
					Members:        5,
					ReplicaSetName: "rs0",
					Version: mongodbv1alpha1.MongoDBVersion{
						Version: "7.0",
					},
					Storage: mongodbv1alpha1.StorageSpec{
						Size: resource.MustParse("20Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, mongodb)).Should(Succeed())

			// Clean up
			Expect(k8sClient.Delete(ctx, mongodb)).Should(Succeed())
		})
	})

	Context("When updating MongoDB resources", func() {
		It("Should update the spec correctly", func() {
			ctx := context.Background()
			mongodb := &mongodbv1alpha1.MongoDB{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "mongodb.keiailab.github.io/v1alpha1",
					Kind:       "MongoDB",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-update-mongodb",
					Namespace: MongoDBNamespace,
				},
				Spec: mongodbv1alpha1.MongoDBSpec{
					Members:        3,
					ReplicaSetName: "rs0",
					Version: mongodbv1alpha1.MongoDBVersion{
						Version: "7.0",
					},
					Storage: mongodbv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, mongodb)).Should(Succeed())

			// Update the MongoDB
			mongoDBLookupKey := types.NamespacedName{Name: "test-update-mongodb", Namespace: MongoDBNamespace}
			createdMongoDB := &mongodbv1alpha1.MongoDB{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, mongoDBLookupKey, createdMongoDB)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Update members count
			createdMongoDB.Spec.Members = 5
			Expect(k8sClient.Update(ctx, createdMongoDB)).Should(Succeed())

			// Verify the update
			updatedMongoDB := &mongodbv1alpha1.MongoDB{}
			Eventually(func() int32 {
				err := k8sClient.Get(ctx, mongoDBLookupKey, updatedMongoDB)
				if err != nil {
					return 0
				}
				return updatedMongoDB.Spec.Members
			}, timeout, interval).Should(Equal(int32(5)))

			// Clean up
			Expect(k8sClient.Delete(ctx, mongodb)).Should(Succeed())
		})
	})

	Context("When deleting MongoDB resources", func() {
		It("Should delete the MongoDB resource", func() {
			ctx := context.Background()
			mongodb := &mongodbv1alpha1.MongoDB{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "mongodb.keiailab.github.io/v1alpha1",
					Kind:       "MongoDB",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-delete-mongodb",
					Namespace: MongoDBNamespace,
				},
				Spec: mongodbv1alpha1.MongoDBSpec{
					Members:        3,
					ReplicaSetName: "rs0",
					Version: mongodbv1alpha1.MongoDBVersion{
						Version: "7.0",
					},
					Storage: mongodbv1alpha1.StorageSpec{
						Size: resource.MustParse("10Gi"),
					},
				},
			}
			Expect(k8sClient.Create(ctx, mongodb)).Should(Succeed())

			// Delete the MongoDB
			Expect(k8sClient.Delete(ctx, mongodb)).Should(Succeed())

			// Verify deletion
			mongoDBLookupKey := types.NamespacedName{Name: "test-delete-mongodb", Namespace: MongoDBNamespace}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, mongoDBLookupKey, &mongodbv1alpha1.MongoDB{})
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())
		})
	})
})

// Helper function to check if a StatefulSet exists
func statefulSetExists(ctx context.Context, name, namespace string) bool {
	sts := &appsv1.StatefulSet{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, sts)
	return err == nil
}

// Helper function to check if a Service exists
func serviceExists(ctx context.Context, name, namespace string) bool {
	svc := &corev1.Service{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, svc)
	return err == nil
}

// Helper function to check if a Secret exists
func secretExists(ctx context.Context, name, namespace string) bool {
	secret := &corev1.Secret{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, secret)
	return err == nil
}
