package notification

import (
	"github.com/hashicorp/go-hclog"
	"github.com/scttfrdmn/snoozebot/pkg/notification/providers/email"
	"github.com/scttfrdmn/snoozebot/pkg/notification/providers/slack"
	"github.com/scttfrdmn/snoozebot/pkg/notification/types"
)

// ProviderFactory is a function that creates a new notification provider
type ProviderFactory func(logger hclog.Logger) types.NotificationProvider

// providerFactories is a map of provider names to their factories
var providerFactories = map[string]ProviderFactory{
	slack.ProviderName: func(logger hclog.Logger) types.NotificationProvider {
		return slack.New(logger)
	},
	email.ProviderName: func(logger hclog.Logger) types.NotificationProvider {
		return email.New(logger)
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