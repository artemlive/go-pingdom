package pingdom

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const (
	defaultBaseURL = "https://api.pingdom.com/"
)

type Client struct {
	User     string
	Password string
	APIKey   string
	BaseURL  *url.URL
	client   *http.Client
}

type Check struct {
	ID                       int    `json:"id"`
	Name                     string `json:"name"`
	Resolution               int    `json:"resolution,omitempty"`
	SendToEmail              bool   `json:"sendtoemail,omitempty"`
	SendToTwitter            bool   `json:"sendtotwitter,omitempty"`
	SendToIPhone             bool   `json:"sendtoiphone,omitempty"`
	SendNotificationWhenDown int    `json:"sendnotificationwhendown,omitempty"`
	NotifyAgainEvery         int    `json:"notifyagainevery,omitempty"`
	NotifyWhenBackup         bool   `json:"notifywhenbackup,omitempty"`
	Created                  int64  `json:"created,omitempty"`
	Hostname                 string `json:"hostname,omitempty"`
	Status                   string `json:"status,omitempty"`
	LastErrorTime            int64  `json:"lasterrortime,omitempty"`
	LastTestTime             int64  `json:"lasttesttime,omitempty"`
	LastResponseTime         int64  `json:"lastresponsetime,omitempty"`
}

type CheckResponse struct {
	Check Check `json:"check"`
}

type ListChecksResponse struct {
	Checks []Check `json:"checks"`
}

type PingdomResponse struct {
	Message string `json:"message"`
}

type PingdomErrorResponse struct {
	Error PingdomError `json:"error"`
}

type PingdomError struct {
	StatusCode int    `json:"statuscode"`
	StatusDesc string `json:"statusdesc"`
	Message    string `json:"errormessage"`
}

func (r *PingdomError) Error() string {
	return fmt.Sprintf("%d %v: %v", r.StatusCode, r.StatusDesc, r.Message)
}

func NewClient(user string, password string, key string) *Client {
	baseURL, _ := url.Parse(defaultBaseURL)
	c := &Client{user, password, key, baseURL, http.DefaultClient}
	return c
}

func (pc *Client) NewRequest(method string, rsc string, params map[string]string) (*http.Request, error) {
	baseUrl, err := url.Parse(pc.BaseURL.String() + rsc)
	if err != nil {
		return nil, err
	}

	if params != nil {
		ps := url.Values{}
		for k, v := range params {
			ps.Set(k, v)
		}
		baseUrl.RawQuery = ps.Encode()
	}

	req, err := http.NewRequest(method, baseUrl.String(), nil)
	req.SetBasicAuth(pc.User, pc.Password)
	req.Header.Add("App-Key", pc.APIKey)
	return req, err
}

func ValidateResponse(r *http.Response) error {
	if c := r.StatusCode; 200 <= c && c <= 299 {
		return nil
	}

	bodyBytes, _ := ioutil.ReadAll(r.Body)
	bodyString := string(bodyBytes)
	m := &PingdomErrorResponse{}
	err := json.Unmarshal([]byte(bodyString), &m)
	if err != nil {
		return err
	}

	return &m.Error
}

func (ck *Check) Params() map[string]string {
	return map[string]string{
		"name": ck.Name,
		"host": ck.Hostname,
		"type": "http",
	}
}

func (pc *Client) ListChecks() ([]Check, error) {
	req, err := pc.NewRequest("GET", "/api/2.0/checks", nil)
	if err != nil {
		return nil, err
	}

	resp, err := pc.client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if err := ValidateResponse(resp); err != nil {
		return nil, err
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	m := &ListChecksResponse{}
	err = json.Unmarshal([]byte(bodyString), &m)
	return m.Checks, err
}

func (pc *Client) CreateCheck(check Check) (*Check, error) {
	req, err := pc.NewRequest("POST", "/api/2.0/checks", check.Params())
	if err != nil {
		return nil, err
	}

	resp, err := pc.client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if err := ValidateResponse(resp); err != nil {
		return nil, err
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	m := &CheckResponse{}
	err = json.Unmarshal([]byte(bodyString), &m)
	return &m.Check, err

}

func (pc *Client) ReadCheck(id int) (*Check, error) {
	req, err := pc.NewRequest("GET", "/api/2.0/checks/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}

	resp, err := pc.client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if err := ValidateResponse(resp); err != nil {
		return nil, err
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	m := &CheckResponse{}
	err = json.Unmarshal([]byte(bodyString), &m)
	return &m.Check, err
}

func (pc *Client) UpdateCheck(id int, check Check) (*PingdomResponse, error) {
	params := check.Params()
	delete(params, "type")
	req, err := pc.NewRequest("PUT", "/api/2.0/checks/"+strconv.Itoa(id), params)
	if err != nil {
		return nil, err
	}

	resp, err := pc.client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if err := ValidateResponse(resp); err != nil {
		return nil, err
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	m := &PingdomResponse{}
	err = json.Unmarshal([]byte(bodyString), &m)
	return m, err
}

func (pc *Client) DeleteCheck(id int) (*PingdomResponse, error) {
	req, err := pc.NewRequest("DELETE", "/api/2.0/checks/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}

	resp, err := pc.client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	if err := ValidateResponse(resp); err != nil {
		return nil, err
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)

	m := &PingdomResponse{}
	err = json.Unmarshal([]byte(bodyString), &m)
	return m, err
}
