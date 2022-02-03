
package webrelay

import (
	"fmt"
	//"bytes"
	//"io"
	"io/ioutil"
	"strings"
	"net/http"
	"net/url"
	"encoding/base64"
	"encoding/xml"
)

type webRelayQuad struct {
	c *Client
}

type param struct {
	key string
	value string
}

type webRelayQuadRequest struct {
	params []param
}

type webRelayQuadResponse struct {
        Relay1State int         `xml:"relay1state"`
        Relay2State int         `xml:"relay2state"`
        Relay3State int         `xml:"relay3state"`
        Relay4State int         `xml:"relay4state"`
}

func IsWebRelayQuad(c *Client) (bool, error) {
	url := &url.URL{
		Scheme: "http",
		Host: c.Host,
		Path: "menu.html",
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
	return strings.Contains(string(buf), "<title>WebRelay-Quad</title>"), nil
}

func WebRelayQuadModel(c *Client) Model {
	return &webRelayQuad{
		c: c,
	}
}

func (wrq *webRelayQuad) Name() string {
	return "WebRelay-Quad"
}

func (wrq *webRelayQuad) GetState() ([]bool, error) {
	req := webRelayQuadRequest{}
	resp, err := wrq.stateFull(&req)
	if err != nil {
		return nil, err
	}

	values := make([]bool, 4)
	if resp.Relay1State > 0 {
		values[0] = true
	}
	if resp.Relay2State > 0 {
		values[1] = true
	}
	if resp.Relay3State > 0 {
		values[2] = true
	}
	if resp.Relay4State > 0 {
		values[3] = true
	}
	return values, nil
}

func (wrq *webRelayQuad) SetState(enable, values []bool) error {
	if len(enable) > 4 || len(values) > 4 {
		return fmt.Errorf("Relay arrays too large")
	}
	req := webRelayQuadRequest{}
	req.params = []param{}
	for i := range enable {
		if !enable[i] {
			continue
		}
		v := false
		if i < len(values) {
			v = values[i]
		}
		vNum := 0
		if v {
			vNum = 1
		}
		req.params = append(req.params, param{
			key: fmt.Sprintf("relay%dState", i+1),
			value: fmt.Sprintf("%d", vNum),
		})
	}
	_, err := wrq.stateFull(&req)
	return err
}

func (wrq *webRelayQuad) SetRelay(relay int, value bool) error {
	if relay < 1 || relay > 4 {
		return fmt.Errorf("Relay number is out of range")
	}
	enable := make([]bool, relay)
	values := make([]bool, relay)
	enable[relay-1] = true
	values[relay-1] = value
	return wrq.SetState(enable, values)
}

// Send the stateFull request and parse the resulting response
func (wrq webRelayQuad) stateFull(req *webRelayQuadRequest) (*webRelayQuadResponse, error) {
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
		Scheme: "http",
		Host: wrq.c.Host,
		Path: "/stateFull.xml",
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
		URL: url,
		Header: headers,
	}
	//fmt.Printf("URL: \"%s\"\n", httpReq.URL)
	resp, err := wrq.c.Provider.Do(httpReq)
	if err != nil {
		return nil, err
	}
	body := resp.Body
	defer body.Close()
	//buf := &bytes.Buffer{}
	//teeBody := io.TeeReader(body, buf)
	xmlDecoder := xml.NewDecoder(body)
	response := &webRelayQuadResponse{}
	err = xmlDecoder.Decode(response)
	if err != nil {
		return nil, err
	}
	//txt, _ := ioutil.ReadAll(buf)
	//fmt.Printf("Got: %s\n", txt)
	//fmt.Printf("Cvt: %v\n", response)
	return response, nil
}
