package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	AUTH_INTERNAL = "internal"
	REFRESH_TOKEN = "refresh_token"

	ARGS_LEN_ERROR = "Invalid number of arguments. Expected %d, but got %d"
)

type CherwellClient struct {
	host  string
	token string
}

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_typoe"`
	ExpiresIn    time.Time `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
}

type BusinessObject struct {
	BusObID       string                 `json:"busObId"`
	BusObRecID    string                 `json:"busObRecId"`
	BusObPublicID string                 `json:"busObPublicId"`
	Fields        []BusinessObjectFields `json:"fields"`
}

type BusinessObjectFields struct {
	FieldID string `json:"fieldId"`
	Name    string `json:"name"`
	Value   string `json:"value"`
	Dirty   bool   `json:"dirty"`
}

type SaveBusObjResponse struct {
	BusObPublicID string `json:"busObPublicId"`
	BusObRecID    string `json:"busObRecId"`
}

func NewClient(host string) *CherwellClient {
	return &CherwellClient{host: host}
}

func (c *CherwellClient) Authenticate(tp string, args ...string) (AuthResponse, error) {
	var (
		vals    url.Values
		authRes AuthResponse
	)

	switch tp {
	case AUTH_INTERNAL:
		if l := len(args); l < 3 {
			return authRes, fmt.Errorf(ARGS_LEN_ERROR, 4, l+1)
		}

		vals = url.Values{
			"grant_type": {"password"},
			"client_id":  {args[0]},
			"username":   {args[1]},
			"password":   {args[2]},
			"auth_mode":  {AUTH_INTERNAL},
		}
	case REFRESH_TOKEN:
		if l := len(args); l < 2 {
			return authRes, fmt.Errorf(ARGS_LEN_ERROR, 3, l+1)
		}

		vals = url.Values{
			"grant_type":    {REFRESH_TOKEN},
			"client_id":     {args[0]},
			"refresh_token": {args[1]},
		}
	default:
		return authRes, errors.New("Invalid Authentication Type.")
	}

	resp, err := http.PostForm(c.host+"/token", vals)
	if err != nil {
		return authRes, err
	}
	defer resp.Body.Close()

	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return authRes, err
	} else if resp.StatusCode == http.StatusOK {
		json.Unmarshal(body, &authRes)
	} else {
		return authRes, Error{resp.StatusCode, string(body)}
	}

	c.token = authRes.AccessToken

	return authRes, nil
}

func (c *CherwellClient) SaveBusinessObject(bo BusinessObject) (SaveBusObjResponse, error) {
	var (
		saveBusRes SaveBusObjResponse
	)

	if err := c.requestAPI("/api/V1/savebusinessobject", http.MethodPost, bo, &saveBusRes); err != nil {
		return saveBusRes, err
	}

	return saveBusRes, nil
}

func (c *CherwellClient) requestAPI(endpoint, method string, payload, output interface{}) error {
	var (
		err  error
		p    []byte
		resp *http.Response
	)

	p, err = json.Marshal(payload)
	if err != nil {
		return err
	}

	req, _ := http.NewRequest(method, c.host+endpoint, bytes.NewBuffer(p))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	client := &http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if body, err := ioutil.ReadAll(resp.Body); err != nil {
		return err
	} else if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(body, output); err != nil {
			return err
		}
	} else {
		return Error{resp.StatusCode, string(body)}
	}

	return nil
}
