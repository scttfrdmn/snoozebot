package notification

import (
	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/notification/providers/slack"
)

// ProviderFactory is a function that creates a new notification provider
type ProviderFactory func(logger hclog.Logger) NotificationProvider

// providerFactories is a map of provider names to their factories
var providerFactories = map[string]ProviderFactory{
	slack.ProviderName: func(logger hclog.Logger) NotificationProvider {
		return slack.New(logger)
	},
}

// RegisterProviderFactory registers a new provider factory
func RegisterProviderFactory(name string, factory ProviderFactory) {
	providerFactories[name] = factory
}

// GetProviderFactory returns the factory for a provider
func GetProviderFactory(name string) (ProviderFactory, bool) {
	factory, ok := providerFactories[name]
	return factory, ok
}