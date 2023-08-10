// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build kubelet

// Package kubelet TODO comment
package kubelet

import (
	"github.com/DataDog/datadog-agent/pkg/aggregator/sender"
	"github.com/DataDog/datadog-agent/pkg/autodiscovery/integration"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	core "github.com/DataDog/datadog-agent/pkg/collector/corechecks"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/containers/kubelet/common"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/containers/kubelet/provider/node"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/containers/kubelet/provider/pod"
	"github.com/DataDog/datadog-agent/pkg/util/containers"
	"github.com/DataDog/datadog-agent/pkg/util/kubernetes/kubelet"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

const (
	// KubeletCheckName exported const should have comment (or a comment on this block) or be unexported
	KubeletCheckName = "kubelet_core"
)

// Provider provides the metrics related to a given Kubelet endpoint
type Provider interface {
	Provide(kubelet.KubeUtilInterface, sender.Sender) error
}

// KubeletCheck wraps the config and the metric stores needed to run the check
type KubeletCheck struct {
	core.CheckBase
	instance  *common.KubeletConfig
	filter    *containers.Filter
	providers []Provider
}

// NewKubeletCheck returns a new KubeletCheck
func NewKubeletCheck(base core.CheckBase, instance *common.KubeletConfig) *KubeletCheck {
	filter, err := containers.GetSharedMetricFilter()
	if err != nil {
		log.Warnf("Can't get container include/exclude filter, no filtering will be applied: %v", err)
	}

	providers := initProviders(filter, instance)

	return &KubeletCheck{
		CheckBase: base,
		instance:  instance,
		filter:    filter,
		providers: providers,
	}
}

func initProviders(filter *containers.Filter, config *common.KubeletConfig) []Provider {
	podProvider := pod.NewProvider(filter, config)
	// nodeProvider collects from the /spec endpoint, which was hidden by default in k8s 1.18 and removed in k8s 1.19.
	// It is here for backwards compatibility.
	nodeProvider := node.NewProvider(config)

	return []Provider{
		podProvider,
		nodeProvider,
	}
}

// KubeletFactory returns a new KubeletCheck
func KubeletFactory() check.Check {
	return NewKubeletCheck(core.NewCheckBase(KubeletCheckName), &common.KubeletConfig{})
}

// Configure exported method should have comment or be unexported
func (k *KubeletCheck) Configure(integrationConfigDigest uint64, config, initConfig integration.Data, source string) error {
	err := k.CommonConfigure(integrationConfigDigest, initConfig, config, source)
	if err != nil {
		return err
	}

	err = k.instance.Parse(config)
	if err != nil {
		return err
	}

	return nil
}

// Run exported method should have comment or be unexported
func (k *KubeletCheck) Run() error {
	sender, err := k.GetSender()
	if err != nil {
		return err
	}
	defer sender.Commit()

	// Get client
	kc, err := kubelet.GetKubeUtil()
	if err != nil {
		_ = k.Warnf("Error initialising check: %s", err)
		return err
	}

	for _, provider := range k.providers {
		err = provider.Provide(kc, sender)
		if err != nil {
			_ = k.Warnf("Error reporting metrics: %s", err)
		}
	}

	return nil
}

func init() {
	core.RegisterCheck(KubeletCheckName, KubeletFactory)
}
