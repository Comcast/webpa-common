/**
 * Copyright 2021 Comcast Cable Communications Management, LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package xwebhook

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/xmidt-org/argus/chrysom"
	"github.com/xmidt-org/argus/model"
	"github.com/xmidt-org/argus/store"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/webpa-common/xmetrics"
)

// Service describes the core operations around webhook subscriptions.
// Initialize() provides a service ready to use and the controls around watching for updates.
type Service interface {
	// Add adds the given owned webhook to the current list of webhooks. If the operation
	// succeeds, a non-nil error is returned.
	Add(owner string, w *Webhook) error

	// AllWebhooks lists all the current webhooks for the given owner.
	// If an owner is not provided, all webhooks are returned.
	AllWebhooks(owner string) ([]Webhook, error)
}
type loggerGroup struct {
	Error log.Logger
	Debug log.Logger
}

type service struct {
	argus   *chrysom.Client
	loggers *loggerGroup
}

func (s *service) Add(owner string, w *Webhook) error {
	item, err := webhookToItem(w)
	if err != nil {
		return err
	}
	ctx := logging.WithLogger(context.Background(), s.loggers.Debug)
	result, err := s.argus.PushItem(ctx, owner, *item)
	if err != nil {
		return err
	}

	if result == chrysom.CreatedPushResult || result == chrysom.UpdatedPushResult {
		return nil
	}
	return errors.New("operation to add webhook to db failed")
}

func (s *service) AllWebhooks(owner string) ([]Webhook, error) {
	s.loggers.Debug.Log("msg", "AllWebhooks called", "owner", owner)
	ctx := logging.WithLogger(context.Background(), s.loggers.Debug)
	items, err := s.argus.GetItems(ctx, owner)
	if err != nil {
		return nil, err
	}
	webhooks := []Webhook{}
	for i := range items {
		item := items[i]
		webhook, err := itemToWebhook(&item)
		if err != nil {
			continue
		}
		webhooks = append(webhooks, *webhook)
	}

	return webhooks, nil
}

func webhookToItem(w *Webhook) (*model.Item, error) {
	encodedWebhook, err := json.Marshal(w)
	if err != nil {
		return nil, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(encodedWebhook, &data)
	if err != nil {
		return nil, err
	}

	TTLSeconds := int64(w.Duration.Seconds())

	return &model.Item{
		Data: data,
		ID:   store.Sha256HexDigest(w.Config.URL),
		TTL:  &TTLSeconds,
	}, nil
}

func itemToWebhook(i *model.Item) (*Webhook, error) {
	w := new(Webhook)
	encodedWebhook, err := json.Marshal(i.Data)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(encodedWebhook, w)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func newLoggerGroup(root log.Logger) *loggerGroup {
	if root == nil {
		root = log.NewNopLogger()
	}

	return &loggerGroup{
		Debug: log.WithPrefix(root, level.Key(), level.DebugValue()),
		Error: log.WithPrefix(root, level.Key(), level.ErrorValue()),
	}

}

// Initialize builds the webhook service from the given configuration. It allows adding watchers for the internal subscription state. Call the returned
// function when you are done watching for updates.
func Initialize(cfg *Config, provider xmetrics.Registry, watches ...Watch) (Service, func(), error) {
	validateConfig(cfg)

	watches = append(watches, webhookListSizeWatch(provider.NewGauge(WebhookListSizeGauge)))

	cfg.Argus.Listen.Listener = createArgusListener(watches...)

	m := &chrysom.Measures{
		Polls: provider.NewCounterVec(chrysom.PollCounter),
	}

	argus, err := chrysom.NewClient(cfg.Argus, m, nil, nil)
	if err != nil {
		return nil, nil, err
	}

	svc := &service{
		loggers: newLoggerGroup(cfg.Argus.Logger),
		argus:   argus,
	}

	argus.Start(context.Background())

	return svc, func() { argus.Stop(context.Background()) }, nil
}

func createArgusListener(watches ...Watch) chrysom.Listener {
	if len(watches) < 1 {
		return nil
	}
	return chrysom.ListenerFunc(func(items chrysom.Items) {
		webhooks := itemsToWebhooks(items)
		for _, watch := range watches {
			watch.Update(webhooks)
		}
	})
}

func itemsToWebhooks(items []model.Item) []Webhook {
	webhooks := []Webhook{}
	for _, item := range items {
		webhook, err := itemToWebhook(&item)
		if err != nil {
			continue
		}
		webhooks = append(webhooks, *webhook)
	}
	return webhooks
}

func validateConfig(cfg *Config) {
	if cfg.WatchUpdateInterval == 0 {
		cfg.WatchUpdateInterval = time.Second * 5
	}
}
