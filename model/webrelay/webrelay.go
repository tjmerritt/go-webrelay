package model_webrelay

// UNTESTED

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/tjmerritt/go-webrelay"
)

var model = "WebRelay"

func init() {
	webrelay.RegisterModel(model, probe)
}

func probe(c *webrelay.Client) (bool, webrelay.Model) {
	found, err := IsWebRelay(c); 
	if err != nil || !found {
		return false, nil
	}
	return true, WebRelayModel(c)
}

type webRelay struct {
	c *webrelay.Client
}

type param struct {
	key   string
	value string
}

type webRelayRequest struct {
	params []param
}

type webRelayResponse struct {
	RelayState   int `xml:"relay1state"`
	InputState   int `xml:"inputstate"`
	RebootStatee int `xml:"rebootstate"`
	TotalReboots int `xml:"totalreboots"`
}

func IsWebRelay(c *webrelay.Client) (bool, error) {
	url := &url.URL{
		Scheme: "http",
		Host:   c.Host,
		Path:   "menu.html",
	}
	resp, err := c.Provider.Get(url.String())
	if err != nil {
		return false, err
	}
	body := resp.Body
	defer body.Close()
	buf, err := ioutil.ReadAll(body)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(buf), "<title>WebRelay</title>"), nil
}

func WebRelayModel(c *webrelay.Client) webrelay.Model {
	return &webRelay{
		c: c,
	}
}

func (wrq *webRelay) Name() string {
	return "WebRelay"
}

func (wrq *webRelay) GetState() ([]bool, error) {
	req := webRelayRequest{}
	resp, err := wrq.stateFull(&req)
	if err != nil {
		return nil, err
	}

	values := make([]bool, 1)
	if resp.RelayState > 0 {
		values[0] = true
	}
	return values, nil
}

func (wrq *webRelay) SetState(enable, values []bool) error {
	if len(enable) > 1 || len(values) > 1 {
		return fmt.Errorf("Relay arrays too large")
	}
	req := webRelayRequest{}
	req.params = []param{}
	if enable[0] {
		v := false
		if len(values) > 0 {
			v = values[0]
		}
		vStr := "0"
		if v {
			vStr = "1"
		}
		req.params = append(req.params, param{
			key:   "relayState",
			value: vStr,
		})
	}
	_, err := wrq.stateFull(&req)
	return err
}

func (wrq *webRelay) SetRelay(relay int, value bool) error {
	if relay < 1 || relay > 1 {
		return fmt.Errorf("Relay number is out of range")
	}
	enable := []bool{ true }
	values := []bool{ value }
	return wrq.SetState(enable, values)
}

// Send the stateFull request and parse the resulting response
func (wrq webRelay) stateFull(req *webRelayRequest) (*webRelayResponse, error) {
	values := make(url.Values)
	for _, p := range req.params {
		a := values[p.key]
		if a == nil {
			a = []string{}
		}
		a = append(a, p.value)
		values[p.key] = a
	}
	url := &url.URL{
		Scheme:   "http",
		Host:     wrq.c.Host,
		Path:     "/stateFull.xml",
		RawQuery: values.Encode(),
	}
	headers := http.Header{
		"Accept": []string{"*/*"},
	}
	if wrq.c.UserName != "" || wrq.c.ControlPassword != "" {
		passwdStr := wrq.c.UserName + ":" + wrq.c.ControlPassword
		encodedPasswd := base64.StdEncoding.EncodeToString([]byte(passwdStr))
		headers["Authorization"] = []string{"Basic " + encodedPasswd}
	}
	httpReq := &http.Request{
		Method: "GET",
		URL:    url,
		Header: headers,
	}
	resp, err := wrq.c.Provider.Do(httpReq)
	if err != nil {
		return nil, err
	}
	body := resp.Body
	defer body.Close()
	xmlDecoder := xml.NewDecoder(body)
	response := &webRelayResponse{}
	err = xmlDecoder.Decode(response)
	if err != nil {
		return nil, err
	}
	return response, nil
}
