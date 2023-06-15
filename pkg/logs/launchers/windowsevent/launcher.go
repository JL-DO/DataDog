// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

//go:build windows
// +build windows

package windowsevent

import (
	"github.com/DataDog/datadog-agent/pkg/util/log"

	"github.com/DataDog/datadog-agent/comp/logs/agent/config"
	"github.com/DataDog/datadog-agent/pkg/logs/auditor"
	"github.com/DataDog/datadog-agent/pkg/logs/launchers"
	"github.com/DataDog/datadog-agent/pkg/logs/pipeline"
	"github.com/DataDog/datadog-agent/pkg/logs/sources"
	"github.com/DataDog/datadog-agent/pkg/logs/tailers"
	tailer "github.com/DataDog/datadog-agent/pkg/logs/tailers/windowsevent"
	newtailer "github.com/DataDog/datadog-agent/pkg/logs/tailers/windowsevent-new"
	"github.com/DataDog/datadog-agent/pkg/util/startstop"
	"github.com/DataDog/datadog-agent/pkg/util/winutil/eventlog/api"
)

type Tailer interface {
	Start(bookmark string)
	startstop.Stoppable
	Identifier() string
}

// Launcher is in charge of starting and stopping windows event logs tailers
type Launcher struct {
	evtapi           evtapi.API
	sources          chan *sources.LogSource
	pipelineProvider pipeline.Provider
	registry         auditor.Registry
	tailers          map[string]Tailer
	stop             chan struct{}
}

// NewLauncher returns a new Launcher.
func NewLauncher() *Launcher {
	return &Launcher{
		tailers: make(map[string]Tailer),
		stop:    make(chan struct{}),
	}
}

// Start starts the launcher.
func (l *Launcher) Start(sourceProvider launchers.SourceProvider, pipelineProvider pipeline.Provider, registry auditor.Registry, tracker *tailers.TailerTracker) {
	l.pipelineProvider = pipelineProvider
	l.sources = sourceProvider.GetAddedForType(config.WindowsEventType)
	availableChannels, err := EnumerateChannels()
	if err != nil {
		log.Debug("Could not list windows event log channels: ", err)
	} else {
		log.Debug("Found available windows event log channels: ", availableChannels)
	}
	go l.run()
}

// run starts new tailers.
func (l *Launcher) run() {
	for {
		select {
		case source := <-l.sources:
			identifier := tailer.Identifier(source.Config.ChannelPath, source.Config.Query)
			if _, exists := l.tailers[identifier]; exists {
				// tailer already setup
				continue
			}
			tailer, err := l.setupTailer(source)
			if err != nil {
				log.Info("Could not set up windows event log tailer: ", err)
			} else {
				l.tailers[identifier] = tailer
			}
		case <-l.stop:
			return
		}
	}
}

// Stop stops all active tailers
func (l *Launcher) Stop() {
	l.stop <- struct{}{}
	stopper := startstop.NewParallelStopper()
	for _, tailer := range l.tailers {
		stopper.Add(tailer)
		delete(l.tailers, tailer.Identifier())
	}
	stopper.Stop()
}

// sanitizedConfig sets default values for the config
func (l *Launcher) sanitizedConfig(sourceConfig *config.LogsConfig) *tailer.Config {
	config := &tailer.Config{
		ChannelPath: sourceConfig.ChannelPath,
		Query:       sourceConfig.Query,
	}
	if config.Query == "" {
		config.Query = "*"
	}
	return config
}

// setupTailer configures and starts a new tailer
func (l *Launcher) setupTailer(source *sources.LogSource) (Tailer, error) {
	sanitizedConfig := l.sanitizedConfig(source.Config)
	var t Tailer
	if source.Config.Type == "windows_event_new" {
		config := &newtailer.Config{
			ChannelPath: sanitizedConfig.ChannelPath,
			Query:       sanitizedConfig.Query,
		}
		t = newtailer.NewTailer(l.evtapi, source, config, l.pipelineProvider.NextPipelineChan())
	} else {
		config := &tailer.Config{
			ChannelPath: sanitizedConfig.ChannelPath,
			Query:       sanitizedConfig.Query,
		}
		t = tailer.NewTailer(source, config, l.pipelineProvider.NextPipelineChan())
	}
	bookmark := l.registry.GetOffset(t.Identifier())
	t.Start(bookmark)
	return t, nil
}
