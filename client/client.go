package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

	if err := unmarshalBody(resp, &authRes); err != nil {
		return authRes, err
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

	if err := unmarshalBody(resp, output); err != nil {
		return err
	}

	return nil
}
