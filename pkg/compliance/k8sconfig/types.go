// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//revive:disable:var-naming

package k8sconfig

import (
	"net"
	"time"
)

// K8sNodeConfig exported type should have comment or be unexported
type K8sNodeConfig struct {
	Version            string               `json:"version"`
	ManagedEnvironment *K8sManagedEnvConfig `json:"managedEnvironment,omitempty"`
	KubeletService     *FileMeta            `json:"kubeletService,omitempty"`
	AdminKubeconfig    *K8sKubeconfigMeta   `json:"adminKubeconfig,omitempty"`
	Components         struct {
		Etcd                  *K8sEtcdConfig                  `json:"etcd,omitempty"`
		KubeApiserver         *K8sKubeApiserverConfig         `json:"kubeApiserver,omitempty"`
		KubeControllerManager *K8sKubeControllerManagerConfig `json:"kubeControllerManager,omitempty"`
		Kubelet               *K8sKubeletConfig               `json:"kubelet,omitempty"`
		KubeProxy             *K8sKubeProxyConfig             `json:"kubeProxy,omitempty"`
		KubeScheduler         *K8sKubeSchedulerConfig         `json:"kubeScheduler,omitempty"`
	} `json:"components"`
	Manifests struct {
		Etcd                 *FileMeta `json:"etcd,omitempty"`
		KubeContollerManager *FileMeta `json:"kubeControllerManager,omitempty"`
		KubeApiserver        *FileMeta `json:"kubeApiserver,omitempty"`
		KubeScheduler        *FileMeta `json:"kubeScheduler,omitempty"`
	} `json:"manifests"`
	Errors []error `json:"errors,omitempty"`
}

// K8sManagedEnvConfig exported type should have comment or be unexported
type K8sManagedEnvConfig struct {
	Name     string      `json:"name"`
	Metadata interface{} `json:"metadata"`
}

// K8sDirMeta exported type should have comment or be unexported
type K8sDirMeta struct {
	Path  string `json:"path"`
	User  string `json:"user"`
	Group string `json:"group"`
	Mode  uint32 `json:"mode"`
}

// FileMeta exported type should have comment or be unexported
type FileMeta struct {
	Path    string      `json:"path"`
	User    string      `json:"user"`
	Group   string      `json:"group"`
	Mode    uint32      `json:"mode"`
	Content interface{} `json:"content" jsonschema:"type=object"`
}

// K8sTokenFileMeta exported type should have comment or be unexported
type K8sTokenFileMeta struct {
	Path  string `json:"path"`
	User  string `json:"user"`
	Group string `json:"group"`
	Mode  uint32 `json:"mode"`
}

type (
	// k8sAdmissionConfigSource https://github.com/kubernetes/kubernetes/blob/6356023cb42d681b7ad0e6d14d1652247d75b797/staging/src/k8s.io/apiserver/pkg/apis/apiserver/types.go#L30
	k8sAdmissionConfigSource struct {
		Plugins []struct {
			Name          string      `yaml:"name"`
			Path          string      `yaml:"path"`
			Configuration interface{} `yaml:"configuration"`
		} `yaml:"plugins"`
	}

	// K8sAdmissionPluginConfigMeta TODO
	K8sAdmissionPluginConfigMeta struct {
		Name          string      `json:"name"`
		Configuration interface{} `json:"configuration,omitempty"`
	}

	//K8sAdmissionConfigFileMeta TODO
	K8sAdmissionConfigFileMeta struct {
		User    string                          `json:"user,omitempty"`
		Group   string                          `json:"group,omitempty"`
		Path    string                          `json:"path,omitempty"`
		Mode    uint32                          `json:"mode,omitempty"`
		Plugins []*K8sAdmissionPluginConfigMeta `json:"plugins"`
	}
)

// K8sKubeconfigMeta exported type should have comment or be unexported
type K8sKubeconfigMeta struct {
	Path       string      `json:"path,omitempty"`
	User       string      `json:"user,omitempty"`
	Group      string      `json:"group,omitempty"`
	Mode       uint32      `json:"mode,omitempty"`
	Kubeconfig interface{} `json:"kubeconfig"`
}

// K8sKeyFileMeta exported type should have comment or be unexported
type K8sKeyFileMeta struct {
	Path  string `json:"path,omitempty"`
	User  string `json:"user,omitempty"`
	Group string `json:"group,omitempty"`
	Mode  uint32 `json:"mode,omitempty"`
}

// K8sCertFileMeta exported type should have comment or be unexported
type K8sCertFileMeta struct {
	Path        string `json:"path,omitempty"`
	User        string `json:"user,omitempty"`
	Group       string `json:"group,omitempty"`
	Mode        uint32 `json:"mode,omitempty"`
	DirUser     string `json:"dirUser,omitempty"`
	DirGroup    string `json:"dirGroup,omitempty"`
	DirMode     uint32 `json:"dirMode,omitempty"`
	Certificate struct {
		Fingerprint  string `json:"fingerprint"`
		SerialNumber string `json:"serialNumber,omitempty"`
		// struct field SubjectKeyId should be SubjectKeyID
		SubjectKeyId string `json:"subjectKeyId,omitempty"`
		// struct field AuthorityKeyId should be AuthorityKeyID
		AuthorityKeyId string    `json:"authorityKeyId,omitempty"`
		CommonName     string    `json:"commonName"`
		Organization   []string  `json:"organization,omitempty"`
		DNSNames       []string  `json:"dnsNames,omitempty"`
		IPAddresses    []net.IP  `json:"ipAddresses,omitempty"`
		NotAfter       time.Time `json:"notAfter"`
		NotBefore      time.Time `json:"notBefore"`
	} `json:"certificate"`
}

type (
	// k8SKubeconfigSource is used to parse the kubeconfig files. It is not
	// exported as-is, and used to build K8sKubeconfig.
	// https://github.com/kubernetes/kubernetes/blob/ad18954259eae3db51bac2274ed4ca7304b923c4/staging/src/k8s.io/client-go/tools/clientcmd/api/types.go#LL31C1-L55C2
	k8SKubeconfigSource struct {
		Kind       string `yaml:"kind,omitempty"`
		APIVersion string `yaml:"apiVersion,omitempty"`

		Clusters []struct {
			Name    string                     `yaml:"name"`
			Cluster k8sKubeconfigClusterSource `yaml:"cluster"`
		} `yaml:"clusters"`

		Users []struct {
			Name string                  `yaml:"name"`
			User k8sKubeconfigUserSource `yaml:"user"`
		} `yaml:"users"`

		Contexts []struct {
			Name    string                     `yaml:"name"`
			Context k8sKubeconfigContextSource `yaml:"context"`
		} `yaml:"contexts"`

		CurrentContext string `yaml:"current-context"`
	}

	k8sKubeconfigClusterSource struct {
		Server                   string `yaml:"server"`
		TLSServerName            string `yaml:"tls-server-name,omitempty"`
		InsecureSkipTLSVerify    bool   `yaml:"insecure-skip-tls-verify,omitempty"`
		CertificateAuthority     string `yaml:"certificate-authority,omitempty"`
		CertificateAuthorityData string `yaml:"certificate-authority-data,omitempty"`
		ProxyURL                 string `yaml:"proxy-url,omitempty"`
		DisableCompression       bool   `yaml:"disable-compression,omitempty"`
	}

	k8sKubeconfigUserSource struct {
		ClientCertificate     string      `yaml:"client-certificate,omitempty"`
		ClientCertificateData string      `yaml:"client-certificate-data,omitempty"`
		ClientKey             string      `yaml:"client-key,omitempty"`
		Token                 string      `yaml:"token,omitempty"`
		TokenFile             string      `yaml:"tokenFile,omitempty"`
		Username              string      `yaml:"username,omitempty"`
		Password              string      `yaml:"password,omitempty"`
		Exec                  interface{} `yaml:"exec,omitempty"`
	}

	k8sKubeconfigContextSource struct {
		Cluster   string `yaml:"cluster"`
		User      string `yaml:"user"`
		Namespace string `yaml:"namespace,omitempty"`
	}

	// K8SKubeconfig exported type should have a comment
	K8SKubeconfig struct {
		Clusters       map[string]*K8sKubeconfigCluster `json:"clusters"`
		Users          map[string]*K8sKubeconfigUser    `json:"users"`
		Contexts       map[string]*K8sKubeconfigContext `json:"contexts"`
		CurrentContext string                           `json:"currentContext"`
	}

	// K8sKubeconfigCluster exported type should have a comment
	K8sKubeconfigCluster struct {
		Server                string           `json:"server"`
		TLSServerName         string           `json:"tlsServerName,omitempty"`
		InsecureSkipTLSVerify bool             `json:"insecureSkipTlsVerify,omitempty"`
		CertificateAuthority  *K8sCertFileMeta `json:"certificateAuthority,omitempty"`
		ProxyURL              string           `json:"proxyUrl,omitempty"`
		DisableCompression    bool             `json:"disableCompression,omitempty"`
	}

	// K8sKubeconfigUser exported type should have a comment
	K8sKubeconfigUser struct {
		UseToken          bool             `json:"useToken"`
		UsePassword       bool             `json:"usePassword"`
		Exec              interface{}      `json:"exec"`
		ClientCertificate *K8sCertFileMeta `json:"clientCertificate,omitempty"`
		ClientKey         *K8sKeyFileMeta  `json:"clientKey,omitempty"`
	}

	// K8sKubeconfigContext exported type should have a comment
	K8sKubeconfigContext struct {
		Cluster   string `json:"cluster"`
		User      string `json:"user"`
		Namespace string `json:"namespace,omitempty"`
	}
)

type (
	// K8sEncryptionProviderConfigFileMeta https://github.com/kubernetes/kubernetes/blob/e1ad9bee5bba8fbe85a6bf6201379ce8b1a611b1/staging/src/k8s.io/apiserver/pkg/apis/config/types.go#L70
	K8sEncryptionProviderConfigFileMeta struct {
		Path      string `json:"path,omitempty"`
		User      string `json:"user,omitempty"`
		Group     string `json:"group,omitempty"`
		Mode      uint32 `json:"mode,omitempty"`
		Resources []struct {
			Resources []string `yaml:"resources" json:"resources"`
			Providers []struct {
				AESGCM    *K8sEncryptionProviderKeysSource `yaml:"aesgcm,omitempty" json:"aesgcm,omitempty"`
				AESCBC    *K8sEncryptionProviderKeysSource `yaml:"aescbc,omitempty" json:"aescbc,omitempty"`
				Secretbox *K8sEncryptionProviderKeysSource `yaml:"secretbox,omitempty" json:"secretbox,omitempty"`
				Identity  *struct{}                        `yaml:"identity,omitempty" json:"identity,omitempty"`
				KMS       *K8sEncryptionProviderKMSSource  `yaml:"kms,omitempty" json:"kms,omitempty"`
			} `yaml:"providers" json:"providers"`
		} `yaml:"resources" json:"resources"`
	}

	// K8sEncryptionProviderKMSSource TODO
	K8sEncryptionProviderKMSSource struct {
		Name      string `yaml:"name" json:"name"`
		Endpoint  string `yaml:"endpoint" json:"endpoint"`
		CacheSize int    `yaml:"cachesize" json:"cachesize"`
		Timeout   string `yaml:"timeout" json:"timeout"`
	}

	// K8sEncryptionProviderKeysSource TODO
	K8sEncryptionProviderKeysSource struct {
		Keys []struct {
			Name string `yaml:"name" json:"name"`
		} `yaml:"keys" json:"keys"`
	}
)
