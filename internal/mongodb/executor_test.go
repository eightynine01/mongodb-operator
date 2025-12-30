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

package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetPodFQDN(t *testing.T) {
	tests := []struct {
		name        string
		podName     string
		serviceName string
		namespace   string
		port        int
		expected    string
	}{
		{
			name:        "standard pod FQDN",
			podName:     "my-mongodb-0",
			serviceName: "my-mongodb-headless",
			namespace:   "default",
			port:        27017,
			expected:    "my-mongodb-0.my-mongodb-headless.default.svc.cluster.local:27017",
		},
		{
			name:        "different namespace",
			podName:     "test-pod-1",
			serviceName: "test-svc",
			namespace:   "mongodb",
			port:        27018,
			expected:    "test-pod-1.test-svc.mongodb.svc.cluster.local:27018",
		},
		{
			name:        "config server pod",
			podName:     "sharded-cfg-0",
			serviceName: "sharded-cfg-headless",
			namespace:   "production",
			port:        27019,
			expected:    "sharded-cfg-0.sharded-cfg-headless.production.svc.cluster.local:27019",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPodFQDN(tt.podName, tt.serviceName, tt.namespace, tt.port)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPodsFQDN(t *testing.T) {
	tests := []struct {
		name        string
		baseName    string
		serviceName string
		namespace   string
		replicas    int
		port        int
		expected    []string
	}{
		{
			name:        "three replica set members",
			baseName:    "my-mongodb",
			serviceName: "my-mongodb-headless",
			namespace:   "default",
			replicas:    3,
			port:        27017,
			expected: []string{
				"my-mongodb-0.my-mongodb-headless.default.svc.cluster.local:27017",
				"my-mongodb-1.my-mongodb-headless.default.svc.cluster.local:27017",
				"my-mongodb-2.my-mongodb-headless.default.svc.cluster.local:27017",
			},
		},
		{
			name:        "single member",
			baseName:    "test-mongo",
			serviceName: "test-mongo-headless",
			namespace:   "test",
			replicas:    1,
			port:        27017,
			expected: []string{
				"test-mongo-0.test-mongo-headless.test.svc.cluster.local:27017",
			},
		},
		{
			name:        "five members config server",
			baseName:    "cfg",
			serviceName: "cfg-headless",
			namespace:   "mongo",
			replicas:    5,
			port:        27019,
			expected: []string{
				"cfg-0.cfg-headless.mongo.svc.cluster.local:27019",
				"cfg-1.cfg-headless.mongo.svc.cluster.local:27019",
				"cfg-2.cfg-headless.mongo.svc.cluster.local:27019",
				"cfg-3.cfg-headless.mongo.svc.cluster.local:27019",
				"cfg-4.cfg-headless.mongo.svc.cluster.local:27019",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetPodsFQDN(tt.baseName, tt.serviceName, tt.namespace, tt.replicas, tt.port)
			assert.Equal(t, tt.expected, result)
			assert.Len(t, result, tt.replicas)
		})
	}
}

func TestExecResult(t *testing.T) {
	result := &ExecResult{
		Stdout:   "test output",
		Stderr:   "test error",
		ExitCode: 0,
	}

	assert.Equal(t, "test output", result.Stdout)
	assert.Equal(t, "test error", result.Stderr)
	assert.Equal(t, 0, result.ExitCode)
}
