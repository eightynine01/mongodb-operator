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
	"bytes"
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Executor handles executing commands in MongoDB pods
type Executor struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// NewExecutor creates a new MongoDB command executor
func NewExecutor() (*Executor, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &Executor{
		clientset: clientset,
		config:    cfg,
	}, nil
}

// ExecResult contains the result of a command execution
type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// ExecuteCommand executes a command in a pod container
func (e *Executor) ExecuteCommand(ctx context.Context, podName, namespace, container string, command []string) (*ExecResult, error) {
	req := e.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     false,
			Stdout:    true,
			Stderr:    true,
			TTY:       false,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(e.config, "POST", req.URL())
	if err != nil {
		return nil, fmt.Errorf("failed to create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})

	result := &ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
	}

	if err != nil {
		result.ExitCode = 1
		// Don't return error for non-zero exit codes, just set the exit code
		if !strings.Contains(err.Error(), "command terminated with exit code") {
			return result, fmt.Errorf("failed to execute command: %w", err)
		}
	}

	return result, nil
}

// ExecuteMongosh executes a mongosh command in the MongoDB container
func (e *Executor) ExecuteMongosh(ctx context.Context, podName, namespace, command string) (*ExecResult, error) {
	return e.ExecuteMongoshWithPort(ctx, podName, namespace, command, 27017)
}

// ExecuteMongoshWithPort executes a mongosh command with specified port
func (e *Executor) ExecuteMongoshWithPort(ctx context.Context, podName, namespace, command string, port int) (*ExecResult, error) {
	return e.ExecuteMongoshInContainer(ctx, podName, namespace, "mongodb", command, port)
}

// ExecuteMongoshInContainer executes a mongosh command in a specified container
func (e *Executor) ExecuteMongoshInContainer(ctx context.Context, podName, namespace, container, command string, port int) (*ExecResult, error) {
	return e.ExecuteCommand(ctx, podName, namespace, container, []string{
		"mongosh",
		"--quiet",
		"--port", fmt.Sprintf("%d", port),
		"--eval",
		command,
	})
}

// ExecuteMongoshWithAuth executes a mongosh command with authentication
func (e *Executor) ExecuteMongoshWithAuth(ctx context.Context, podName, namespace, username, password, authDB, command string) (*ExecResult, error) {
	return e.ExecuteMongoshWithAuthAndPort(ctx, podName, namespace, username, password, authDB, command, 27017)
}

// ExecuteMongoshWithAuthAndPort executes a mongosh command with authentication and specified port
func (e *Executor) ExecuteMongoshWithAuthAndPort(ctx context.Context, podName, namespace, username, password, authDB, command string, port int) (*ExecResult, error) {
	return e.ExecuteMongoshWithAuthInContainer(ctx, podName, namespace, "mongodb", username, password, authDB, command, port)
}

// ExecuteMongoshWithAuthInContainer executes a mongosh command with authentication in a specified container
func (e *Executor) ExecuteMongoshWithAuthInContainer(ctx context.Context, podName, namespace, container, username, password, authDB, command string, port int) (*ExecResult, error) {
	return e.ExecuteCommand(ctx, podName, namespace, container, []string{
		"mongosh",
		"--quiet",
		"--port", fmt.Sprintf("%d", port),
		"-u", username,
		"-p", password,
		"--authenticationDatabase", authDB,
		"--eval",
		command,
	})
}

// ExecuteMongoshJSON executes a mongosh command and expects JSON output
func (e *Executor) ExecuteMongoshJSON(ctx context.Context, podName, namespace, command string) (*ExecResult, error) {
	return e.ExecuteMongoshJSONWithPort(ctx, podName, namespace, command, 27017)
}

// ExecuteMongoshJSONWithPort executes a mongosh command with specified port and expects JSON output
func (e *Executor) ExecuteMongoshJSONWithPort(ctx context.Context, podName, namespace, command string, port int) (*ExecResult, error) {
	// Wrap command to output JSON
	jsonCommand := fmt.Sprintf("JSON.stringify(%s)", command)
	return e.ExecuteMongoshWithPort(ctx, podName, namespace, jsonCommand, port)
}

// ExecuteMongoshOnPrimary executes a command on the primary member
func (e *Executor) ExecuteMongoshOnPrimary(ctx context.Context, podName, namespace, command string) (*ExecResult, error) {
	// First check if this pod is primary
	checkPrimary := `
		const status = rs.status();
		const myState = status.myState;
		if (myState !== 1) {
			throw new Error("Not primary");
		}
	`
	result, err := e.ExecuteMongosh(ctx, podName, namespace, checkPrimary)
	if err != nil {
		return nil, fmt.Errorf("failed to check primary status: %w", err)
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("pod is not primary: %s", result.Stderr)
	}

	// Execute the actual command
	return e.ExecuteMongosh(ctx, podName, namespace, command)
}

// Ping checks if MongoDB is responding
func (e *Executor) Ping(ctx context.Context, podName, namespace string) error {
	result, err := e.ExecuteMongosh(ctx, podName, namespace, "db.adminCommand('ping')")
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("ping failed: %s", result.Stderr)
	}
	return nil
}

// GetPodFQDN returns the fully qualified domain name for a pod
func GetPodFQDN(podName, serviceName, namespace string, port int) string {
	return fmt.Sprintf("%s.%s.%s.svc.cluster.local:%d", podName, serviceName, namespace, port)
}

// GetPodsFQDN returns FQDNs for multiple pods
func GetPodsFQDN(baseName, serviceName, namespace string, replicas int, port int) []string {
	fqdns := make([]string, replicas)
	for i := 0; i < replicas; i++ {
		podName := fmt.Sprintf("%s-%d", baseName, i)
		fqdns[i] = GetPodFQDN(podName, serviceName, namespace, port)
	}
	return fqdns
}
