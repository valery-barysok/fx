// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package core

import (
	"sync"
	"time"

	"go.uber.org/fx/core/config"
	"go.uber.org/fx/core/ulog"

	"github.com/uber-go/tally"
)

// A ServiceHost represents the hosting environment for a service instance
type ServiceHost interface {
	Name() string
	Description() string
	Roles() []string
	State() ServiceState
	Metrics() tally.Scope
	Observer() Observer
	Config() config.ConfigurationProvider
	Items() map[string]interface{}
	Logger() ulog.Log
}

// A ServiceHostContainer is meant to be embedded in a LifecycleObserver
// if you want access to the underlying ServiceHost
type ServiceHostContainer struct {
	ServiceHost
}

// SetContainer sets the ServiceHost instance on the container.
// NOTE: This is not thread-safe, and should only be called once during startup.
func (s *ServiceHostContainer) SetContainer(sh ServiceHost) {
	s.ServiceHost = sh
}

// SetContainerer is the interface for anything that you can call SetContainer on
type SetContainerer interface {
	SetContainer(ServiceHost)
}

type serviceCore struct {
	standardConfig serviceConfig
	roles          []string
	state          ServiceState
	configProvider config.ConfigurationProvider
	scopeMux       sync.Mutex
	scope          tally.Scope
	observer       Observer
	items          map[string]interface{}
	logConfig      ulog.Configuration
	log            ulog.Log
}

var _ ServiceHost = &serviceCore{}

func (s *serviceCore) Name() string {
	return s.standardConfig.ServiceName
}

func (s *serviceCore) Description() string {
	return s.standardConfig.ServiceDescription
}

// ServiceOwner is a string in config.
// ServiceOwner is also a struct that embeds ServiceHost
// confus?
func (s *serviceCore) Owner() string {
	return s.standardConfig.ServiceOwner
}

func (s *serviceCore) State() ServiceState {
	return s.state
}

func (s *serviceCore) Roles() []string {
	return s.standardConfig.ServiceRoles
}

// What items?
func (s *serviceCore) Items() map[string]interface{} {
	return s.items
}

func (s *serviceCore) Metrics() tally.Scope {
	// TODO(glib): this is really inefficient, since everyone needing to aquire the scope
	// will hit this mutex. It's much better to initialize the scope during service init, which is
	// currently tricky due to no strict enforcement of options order.
	s.scopeMux.Lock()
	defer s.scopeMux.Unlock()

	// If metrics have not been initialize through the setup, provide a null reporter
	if s.scope == nil {
		s.scope = tally.NewRootScope("", nil, tally.NullStatsReporter, time.Second)
	}

	return s.scope
}

func (s *serviceCore) Observer() Observer {
	return s.observer
}

func (s *serviceCore) Config() config.ConfigurationProvider {
	return s.configProvider
}

func (s *serviceCore) Logger() ulog.Log {
	return s.log
}
