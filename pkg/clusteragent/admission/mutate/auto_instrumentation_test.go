// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubeapiserver
// +build kubeapiserver

package mutate

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInjectAutoInstruConfig(t *testing.T) {
	tests := []struct {
		name           string
		pod            *corev1.Pod
		libsToInject   []libInfo
		expectedEnvKey string
		expectedEnvVal string
		wantErr        bool
	}{
		{
			name: "nominal case: java",
			pod:  fakePod("java-pod"),
			libsToInject: []libInfo{
				{
					lang:  "java",
					image: "gcr.io/datadoghq/dd-lib-java-init:v1",
				},
			},
			expectedEnvKey: "JAVA_TOOL_OPTIONS",
			expectedEnvVal: " -javaagent:/datadog-lib/dd-java-agent.jar",
			wantErr:        false,
		},
		{
			name: "JAVA_TOOL_OPTIONS not empty",
			pod:  fakePodWithEnvValue("java-pod", "JAVA_TOOL_OPTIONS", "predefined"),
			libsToInject: []libInfo{
				{
					lang:  "java",
					image: "gcr.io/datadoghq/dd-lib-java-init:v1",
				},
			},
			expectedEnvKey: "JAVA_TOOL_OPTIONS",
			expectedEnvVal: "predefined -javaagent:/datadog-lib/dd-java-agent.jar",
			wantErr:        false,
		},
		{
			name: "JAVA_TOOL_OPTIONS set via ValueFrom",
			pod:  fakePodWithEnvFieldRefValue("java-pod", "JAVA_TOOL_OPTIONS", "path"),
			libsToInject: []libInfo{
				{
					lang:  "java",
					image: "gcr.io/datadoghq/dd-lib-java-init:v1",
				},
			},
			wantErr: true,
		},
		{
			name: "nominal case: js",
			pod:  fakePod("js-pod"),
			libsToInject: []libInfo{
				{
					lang:  "js",
					image: "gcr.io/datadoghq/dd-lib-js-init:v1",
				},
			},
			expectedEnvKey: "NODE_OPTIONS",
			expectedEnvVal: " --require=/datadog-lib/node_modules/dd-trace/init",
			wantErr:        false,
		},
		{
			name: "NODE_OPTIONS not empty",
			pod:  fakePodWithEnvValue("js-pod", "NODE_OPTIONS", "predefined"),
			libsToInject: []libInfo{
				{
					lang:  "js",
					image: "gcr.io/datadoghq/dd-lib-js-init:v1",
				},
			},
			expectedEnvKey: "NODE_OPTIONS",
			expectedEnvVal: "predefined --require=/datadog-lib/node_modules/dd-trace/init",
			wantErr:        false,
		},
		{
			name: "NODE_OPTIONS set via ValueFrom",
			pod:  fakePodWithEnvFieldRefValue("js-pod", "NODE_OPTIONS", "path"),
			libsToInject: []libInfo{
				{
					lang:  "js",
					image: "gcr.io/datadoghq/dd-lib-js-init:v1",
				},
			},
			wantErr: true,
		},
		{
			name: "nominal case: python",
			pod:  fakePod("python-pod"),
			libsToInject: []libInfo{
				{
					lang:  "python",
					image: "gcr.io/datadoghq/dd-lib-python-init:v1",
				},
			},
			expectedEnvKey: "PYTHONPATH",
			expectedEnvVal: "/datadog-lib/",
			wantErr:        false,
		},
		{
			name: "PYTHONPATH not empty",
			pod:  fakePodWithEnvValue("python-pod", "PYTHONPATH", "predefined"),
			libsToInject: []libInfo{
				{
					lang:  "python",
					image: "gcr.io/datadoghq/dd-lib-python-init:v1",
				},
			},
			expectedEnvKey: "PYTHONPATH",
			expectedEnvVal: "/datadog-lib/:predefined",
			wantErr:        false,
		},
		{
			name: "PYTHONPATH set via ValueFrom",
			pod:  fakePodWithEnvFieldRefValue("python-pod", "PYTHONPATH", "path"),
			libsToInject: []libInfo{
				{
					lang:  "python",
					image: "gcr.io/datadoghq/dd-lib-python-init:v1",
				},
			},
			wantErr: true,
		},
		{
			name: "Unknown language",
			pod:  fakePod("unknown-pod"),
			libsToInject: []libInfo{
				{
					lang:  "unknown",
					image: "gcr.io/datadoghq/dd-lib-unknown-init:v1",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := injectAutoInstruConfig(tt.pod, tt.libsToInject)
			require.False(t, (err != nil) != tt.wantErr)
			if err != nil {
				return
			}
			assertLibReq(t, tt.pod, tt.libsToInject[0].lang, tt.libsToInject[0].image, tt.expectedEnvKey, tt.expectedEnvVal)
		})
	}
}

func assertLibReq(t *testing.T, pod *corev1.Pod, lang language, image, envKey, envVal string) {
	// Empty dir volume
	volumeFound := false
	for _, volume := range pod.Spec.Volumes {
		if volume.Name == "datadog-auto-instrumentation" {
			require.NotNil(t, volume.VolumeSource.EmptyDir)
			volumeFound = true
			break
		}
	}
	require.True(t, volumeFound)

	// Init container
	initContainerFound := false
	for _, container := range pod.Spec.InitContainers {
		if container.Name == fmt.Sprintf("datadog-lib-%s-init", lang) {
			require.Equal(t, image, container.Image)
			require.Equal(t, []string{"sh", "copy-lib.sh", "/datadog-lib"}, container.Command)
			require.Equal(t, "datadog-auto-instrumentation", container.VolumeMounts[0].Name)
			require.Equal(t, "/datadog-lib", container.VolumeMounts[0].MountPath)
			initContainerFound = true
			break
		}
	}
	require.True(t, initContainerFound)

	// App container
	container := pod.Spec.Containers[0]
	require.Equal(t, "datadog-auto-instrumentation", container.VolumeMounts[0].Name)
	require.Equal(t, "/datadog-lib", container.VolumeMounts[0].MountPath)
	envFound := false
	for _, env := range container.Env {
		if env.Name == envKey {
			require.Equal(t, envVal, env.Value)
			envFound = true
			break
		}
	}
	require.True(t, envFound)
}

func TestExtractLibInfo(t *testing.T) {
	tests := []struct {
		name                 string
		pod                  *corev1.Pod
		containerRegistry    string
		expectedLibsToInject []libInfo
	}{
		{
			name:              "java",
			pod:               fakePodWithAnnotation("admission.datadoghq.com/java-lib.version", "v1"),
			containerRegistry: "registry",
			expectedLibsToInject: []libInfo{
				{
					lang:  "java",
					image: "registry/dd-lib-java-init:v1",
				},
			},
		},
		{
			name:              "js",
			pod:               fakePodWithAnnotation("admission.datadoghq.com/js-lib.version", "v1"),
			containerRegistry: "registry",
			expectedLibsToInject: []libInfo{
				{
					lang:  "js",
					image: "registry/dd-lib-js-init:v1",
				},
			},
		},
		{
			name:              "python",
			pod:               fakePodWithAnnotation("admission.datadoghq.com/python-lib.version", "v1"),
			containerRegistry: "registry",
			expectedLibsToInject: []libInfo{
				{
					lang:  "python",
					image: "registry/dd-lib-python-init:v1",
				},
			},
		},
		{
			name:              "custom",
			pod:               fakePodWithAnnotation("admission.datadoghq.com/java-lib.custom-image", "custom/image"),
			containerRegistry: "registry",
			expectedLibsToInject: []libInfo{
				{
					lang:  "java",
					image: "custom/image",
				},
			},
		},
		{
			name:                 "unknown",
			pod:                  fakePodWithAnnotation("admission.datadoghq.com/unknown-lib.version", "v1"),
			containerRegistry:    "registry",
			expectedLibsToInject: []libInfo{},
		},
		{
			name: "java and js",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"admission.datadoghq.com/java-lib.version": "v1",
						"admission.datadoghq.com/js-lib.version":   "v1",
					},
				},
			},
			containerRegistry: "registry",
			expectedLibsToInject: []libInfo{
				{
					lang:  "java",
					image: "registry/dd-lib-java-init:v1",
				},
				{
					lang:  "js",
					image: "registry/dd-lib-js-init:v1",
				},
			},
		},
		{
			name: "java and js on specific containers",
			pod: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"admission.datadoghq.com/java-app.java-lib.version": "v1",
						"admission.datadoghq.com/node-app.js-lib.version":   "v1",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "java-app",
						},
						{
							Name: "node-app",
						},
					},
				},
			},
			containerRegistry: "registry",
			expectedLibsToInject: []libInfo{
				{
					ctrName: "java-app",
					lang:    "java",
					image:   "registry/dd-lib-java-init:v1",
				},
				{
					ctrName: "node-app",
					lang:    "js",
					image:   "registry/dd-lib-js-init:v1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			libsToInject := extractLibInfo(tt.pod, tt.containerRegistry)
			require.Equal(t, tt.expectedLibsToInject, libsToInject)
		})
	}
}

func TestInjectLibConfig(t *testing.T) {
	tests := []struct {
		name         string
		pod          *corev1.Pod
		lang         language
		wantErr      bool
		expectedEnvs []corev1.EnvVar
	}{
		{
			name:    "nominal case",
			pod:     fakePodWithAnnotation("admission.datadoghq.com/java-lib.config.v1", `{"version":1,"service_language":"java","runtime_metrics_enabled":true,"tracing_rate_limit":50}`),
			lang:    java,
			wantErr: false,
			expectedEnvs: []corev1.EnvVar{
				{
					Name:  "DD_RUNTIME_METRICS_ENABLED",
					Value: "true",
				},
				{
					Name:  "DD_TRACE_RATE_LIMIT",
					Value: "50",
				},
			},
		},
		{
			name:         "invalid json",
			pod:          fakePodWithAnnotation("admission.datadoghq.com/java-lib.config.v1", "invalid"),
			lang:         java,
			wantErr:      true,
			expectedEnvs: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := injectLibConfig(tt.pod, tt.lang)
			require.False(t, (err != nil) != tt.wantErr)
			if err != nil {
				return
			}
			container := tt.pod.Spec.Containers[0]
			envCount := 0
			for _, expectEnv := range tt.expectedEnvs {
				for _, contEnv := range container.Env {
					if expectEnv.Name == contEnv.Name {
						require.Equal(t, expectEnv.Value, contEnv.Value)
						envCount++
						break
					}
				}
			}
			require.Equal(t, len(tt.expectedEnvs), envCount)
		})
	}
}
