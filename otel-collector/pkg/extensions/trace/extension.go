// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package trace

import (
	"context"

	"github.com/DataDog/datadog-agent/pkg/trace/agent"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/extension"
)

type traceAgent struct {
	config *Config
	agent  *agent.Agent
}

var _ extension.Extension = (*traceAgent)(nil)

func (a *traceAgent) Start(_ context.Context, host component.Host) error {
	go a.agent.Run()
	return nil
}

func (a *traceAgent) Shutdown(_ context.Context) error {
	return nil
}