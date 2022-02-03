package webrelay

import (
	"fmt"
	"net/http"
)

type Provider interface {
	Get(url string) (*http.Response, error)
	Do(req *http.Request) (*http.Response, error)
}

type Model interface {
	Name() string
	GetState() ([]bool, error)
	SetState(enable, value []bool) error
	SetRelay(relay int, value bool) error
}

type Client struct {
	Provider        Provider
	UserAgent       string
	Host            string
	UserName        string
	ControlPassword string
	model           Model
}

// Construct a new client
func New(host, username, password string) (*Client, error) {
	httpClient := &http.Client{}
	client := &Client{
		Provider:        httpClient,
		Host:            host,
		UserAgent:       "go-webrelay/" + version,
		UserName:        username,
		ControlPassword: password,
	}
	return client, client.setup()
}

func (c *Client) SetProvider(provider Provider) {
	c.Provider = provider
}

func (c *Client) SetUserAgent(agent string) {
	c.UserAgent = agent
}

func (c *Client) probe() (Model, error) {
	if c.model != nil {
		return c.model, nil
	}
	ok, err := IsWebRelayQuad(c)
	if err != nil {
		return nil, err
	}
	if ok {
		c.model = WebRelayQuadModel(c)
		return c.model, nil
	}
	return nil, fmt.Errorf("Unknown model device")
}

func (c *Client) setup() error {
	if c.model != nil {
		return nil
	}
	m, err := c.probe()
	if err != nil {
		return err
	}
	c.model = m
	return nil
}

func (c *Client) GetState() ([]bool, error) {
	if err := c.setup(); err != nil {
		return nil, err
	}
	return c.model.GetState()
}

func (c *Client) SetState(enable, value []bool) error {
	if err := c.setup(); err != nil {
		return err
	}
	return c.model.SetState(enable, value)
}

func (c *Client) SetRelay(relay int, value bool) error {
	if err := c.setup(); err != nil {
		return err
	}
	return c.model.SetRelay(relay, value)
}

func (c *Client) ModelName() string {
	if c.model == nil {
		return ""
	}
	return c.model.Name()
}
