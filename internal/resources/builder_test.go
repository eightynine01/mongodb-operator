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

package resources

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	mongodbv1alpha1 "github.com/keiailab/mongodb-operator/api/v1alpha1"
)

func TestBuildKeyfileSecret(t *testing.T) {
	mdb := &mongodbv1alpha1.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mongodb",
			Namespace: "default",
		},
	}

	secret := BuildKeyfileSecret(mdb)

	assert.Equal(t, "test-mongodb-keyfile", secret.Name)
	assert.Equal(t, "default", secret.Namespace)
	assert.Equal(t, corev1.SecretTypeOpaque, secret.Type)
	assert.Contains(t, secret.Data, "keyfile")
	assert.NotEmpty(t, secret.Data["keyfile"])
}

func TestBuildHeadlessService(t *testing.T) {
	mdb := &mongodbv1alpha1.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mongodb",
			Namespace: "default",
		},
	}

	svc := BuildHeadlessService(mdb)

	assert.Equal(t, "test-mongodb-headless", svc.Name)
	assert.Equal(t, "default", svc.Namespace)
	assert.Equal(t, corev1.ClusterIPNone, svc.Spec.ClusterIP)
	assert.True(t, svc.Spec.PublishNotReadyAddresses)
	assert.Len(t, svc.Spec.Ports, 1)
	assert.Equal(t, int32(27017), svc.Spec.Ports[0].Port)
}

func TestBuildClientService(t *testing.T) {
	mdb := &mongodbv1alpha1.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mongodb",
			Namespace: "default",
		},
	}

	svc := BuildClientService(mdb)

	assert.Equal(t, "test-mongodb", svc.Name)
	assert.Equal(t, "default", svc.Namespace)
	assert.Equal(t, corev1.ServiceTypeClusterIP, svc.Spec.Type)
	assert.Len(t, svc.Spec.Ports, 2)
}

func TestBuildReplicaSetStatefulSet(t *testing.T) {
	mdb := &mongodbv1alpha1.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mongodb",
			Namespace: "default",
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

	sts := BuildReplicaSetStatefulSet(mdb)

	assert.Equal(t, "test-mongodb", sts.Name)
	assert.Equal(t, "default", sts.Namespace)
	assert.Equal(t, int32(3), *sts.Spec.Replicas)
	assert.Equal(t, "test-mongodb-headless", sts.Spec.ServiceName)
	assert.Len(t, sts.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, "mongodb", sts.Spec.Template.Spec.Containers[0].Name)
}

func TestBuildReplicaSetStatefulSetWithStorageClass(t *testing.T) {
	storageClass := "fast-storage"
	mdb := &mongodbv1alpha1.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mongodb",
			Namespace: "default",
		},
		Spec: mongodbv1alpha1.MongoDBSpec{
			Members:        3,
			ReplicaSetName: "rs0",
			Version: mongodbv1alpha1.MongoDBVersion{
				Version: "7.0",
			},
			Storage: mongodbv1alpha1.StorageSpec{
				StorageClassName: storageClass,
				Size:             resource.MustParse("20Gi"),
				DataDirPath:      "/data/db",
			},
		},
	}

	sts := BuildReplicaSetStatefulSet(mdb)

	require.Len(t, sts.Spec.VolumeClaimTemplates, 1)
	require.NotNil(t, sts.Spec.VolumeClaimTemplates[0].Spec.StorageClassName)
	assert.Equal(t, storageClass, *sts.Spec.VolumeClaimTemplates[0].Spec.StorageClassName)
}

func TestBuildReplicaSetStatefulSetWithoutStorageClass(t *testing.T) {
	mdb := &mongodbv1alpha1.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mongodb",
			Namespace: "default",
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

	sts := BuildReplicaSetStatefulSet(mdb)

	require.Len(t, sts.Spec.VolumeClaimTemplates, 1)
	assert.Nil(t, sts.Spec.VolumeClaimTemplates[0].Spec.StorageClassName)
}

func TestBuildReplicaSetStatefulSetWithMonitoring(t *testing.T) {
	mdb := &mongodbv1alpha1.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mongodb",
			Namespace: "default",
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
			Monitoring: &mongodbv1alpha1.MonitoringSpec{
				Enabled: true,
			},
		},
	}

	sts := BuildReplicaSetStatefulSet(mdb)

	// Should have 2 containers: mongodb and exporter
	assert.Len(t, sts.Spec.Template.Spec.Containers, 2)
	assert.Equal(t, "mongodb", sts.Spec.Template.Spec.Containers[0].Name)
	assert.Equal(t, "exporter", sts.Spec.Template.Spec.Containers[1].Name)
}

func TestBuildLabels(t *testing.T) {
	labels := buildLabels("my-instance", "replicaset")

	assert.Equal(t, "mongodb", labels["app.kubernetes.io/name"])
	assert.Equal(t, "my-instance", labels["app.kubernetes.io/instance"])
	assert.Equal(t, "replicaset", labels["app.kubernetes.io/component"])
	assert.Equal(t, "mongodb-operator", labels["app.kubernetes.io/managed-by"])
}

func TestGetMongoDBImage(t *testing.T) {
	tests := []struct {
		name     string
		version  mongodbv1alpha1.MongoDBVersion
		expected string
	}{
		{
			name: "version only",
			version: mongodbv1alpha1.MongoDBVersion{
				Version: "7.0",
			},
			expected: "mongo:7.0",
		},
		{
			name: "custom image",
			version: mongodbv1alpha1.MongoDBVersion{
				Version: "7.0",
				Image:   "myregistry/mongo:7.0-custom",
			},
			expected: "myregistry/mongo:7.0-custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getMongoDBImage(tt.version)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildMongoDBConfigMap(t *testing.T) {
	mdb := &mongodbv1alpha1.MongoDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-mongodb",
			Namespace: "default",
		},
	}

	cm := BuildMongoDBConfigMap(mdb)

	assert.Equal(t, "test-mongodb-scripts", cm.Name)
	assert.Equal(t, "default", cm.Namespace)
	assert.Contains(t, cm.Data, "readiness-probe.sh")
}

func TestBuildConfigServerStatefulSet(t *testing.T) {
	mdbsh := &mongodbv1alpha1.MongoDBSharded{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sharded",
			Namespace: "default",
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
		},
	}

	sts := BuildConfigServerStatefulSet(mdbsh)

	assert.Equal(t, "test-sharded-cfg", sts.Name)
	assert.Equal(t, int32(3), *sts.Spec.Replicas)
	assert.Contains(t, sts.Spec.Template.Spec.Containers[0].Args, "--configsvr")
}

func TestBuildShardStatefulSet(t *testing.T) {
	mdbsh := &mongodbv1alpha1.MongoDBSharded{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sharded",
			Namespace: "default",
		},
		Spec: mongodbv1alpha1.MongoDBShardedSpec{
			Version: mongodbv1alpha1.MongoDBVersion{
				Version: "7.0",
			},
			Shards: mongodbv1alpha1.ShardSpec{
				Count:           2,
				MembersPerShard: 3,
				Storage: mongodbv1alpha1.StorageSpec{
					Size: resource.MustParse("50Gi"),
				},
			},
		},
	}

	sts := BuildShardStatefulSet(mdbsh, 0)

	assert.Equal(t, "test-sharded-shard-0", sts.Name)
	assert.Equal(t, int32(3), *sts.Spec.Replicas)
	assert.Contains(t, sts.Spec.Template.Spec.Containers[0].Args, "--shardsvr")
}

func TestBuildMongosDeployment(t *testing.T) {
	mdbsh := &mongodbv1alpha1.MongoDBSharded{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sharded",
			Namespace: "default",
		},
		Spec: mongodbv1alpha1.MongoDBShardedSpec{
			Version: mongodbv1alpha1.MongoDBVersion{
				Version: "7.0",
			},
			ConfigServer: mongodbv1alpha1.ConfigServerSpec{
				Members: 3,
			},
			Mongos: mongodbv1alpha1.MongosSpec{
				Replicas: 2,
			},
		},
	}

	deploy := BuildMongosDeployment(mdbsh)

	assert.Equal(t, "test-sharded-mongos", deploy.Name)
	assert.Equal(t, int32(2), *deploy.Spec.Replicas)
	assert.Equal(t, "mongos", deploy.Spec.Template.Spec.Containers[0].Command[0])
}
