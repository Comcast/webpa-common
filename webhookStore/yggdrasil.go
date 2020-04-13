package webhookStore

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/xmidt-org/bascule/acquire"
	"github.com/xmidt-org/webpa-common/logging"
	"github.com/xmidt-org/webpa-common/webhook"
	"io/ioutil"
	"net/http"
	"time"
)

type YggdrasilConfig struct {
	Client       *http.Client
	Prefix       string
	PullInterval time.Duration
	Address      string
	Auth         acquire.Acquirer
}

type YggdrasilClient struct {
	client  *http.Client
	options *storeConfig
	config  YggdrasilConfig
	ticker  *time.Ticker
}

func CreateYggdrasilStore(config YggdrasilConfig, options ...Option) *YggdrasilClient {
	clientStore := &YggdrasilClient{
		client: config.Client,
		options: &storeConfig{
			logger: logging.DefaultLogger(),
		},
		config: config,
	}
	for _, o := range options {
		o(clientStore.options)
	}
	clientStore.ticker = time.NewTicker(config.PullInterval)
	go func() {
		for range clientStore.ticker.C {
			if clientStore.options.listener != nil {
				hooks, err := clientStore.GetWebhook()
				if err == nil {
					clientStore.options.listener.Update(hooks)
				} else {
					logging.Error(clientStore.options.logger).Log(logging.MessageKey(), "failed to get webhooks ", logging.ErrorKey(), err)
				}
			}
		}
	}()
	return clientStore
}

func (c *YggdrasilClient) GetWebhook() ([]webhook.W, error) {
	hooks := []webhook.W{}
	request, err := http.NewRequest("GET", fmt.Sprint("%s/store/%s", c.config.Address, c.config.Prefix), nil)
	if err != nil {
		return []webhook.W{}, err
	}
	auth, err := c.config.Auth.Acquire()
	if err != nil {
		return []webhook.W{}, err
	}
	if auth != "" {
		request.Header.Add("Authorization", auth)
	}
	response, err := c.client.Do(request)
	if err != nil {
		return []webhook.W{}, err
	}
	if response.StatusCode != 200 {
		return []webhook.W{}, errors.New("failed to get webhooks, non 200 statuscode")
	}
	data, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return []webhook.W{}, err
	}
	request.Body.Close()

	body := map[string]map[string]interface{}{}
	err = json.Unmarshal(data, &body)
	if err != nil {
		return []webhook.W{}, err
	}

	for _, value := range body {
		data, err := json.Marshal(&value)
		if err != nil {
			continue
		}
		var hook webhook.W
		err = json.Unmarshal(data, &hook)
		if err != nil {
			continue
		}
		hooks = append(hooks, hook)
	}

	return hooks, nil
}

func (c *YggdrasilClient) Push(w webhook.W) error {
	id := base64.RawURLEncoding.EncodeToString([]byte(w.ID()))
	data, err := json.Marshal(&w)
	if err != nil {
		return err
	}
	request, err := http.NewRequest("POST", fmt.Sprint("%s/store/%s/%s", c.config.Address, c.config.Prefix, id), bytes.NewReader(data))
	if err != nil {
		return err
	}
	auth, err := c.config.Auth.Acquire()
	if err != nil {
		return err
	}
	if auth != "" {
		request.Header.Add("Authorization", auth)
	}
	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New("failed to push webhook, non 200 statuscode")
	}
	return nil
}

func (c *YggdrasilClient) Remove(id string) error {
	request, err := http.NewRequest("DELETE", fmt.Sprint("%s/store/%s/%s", c.config.Address, c.config.Prefix, id), nil)
	if err != nil {
		return err
	}
	auth, err := c.config.Auth.Acquire()
	if err != nil {
		return err
	}
	if auth != "" {
		request.Header.Add("Authorization", auth)
	}
	response, err := c.client.Do(request)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New("failed to delete webhook, non 200 statuscode")
	}
	return nil
}

func (c *YggdrasilClient) Stop(context context.Context) {
	c.ticker.Stop()
}

func (c *YggdrasilClient) SetListener(listener Listener) error {
	c.options.listener = listener
	return nil
}