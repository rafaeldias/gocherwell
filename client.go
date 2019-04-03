package gocherwell

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	// AuthInternal is the internal type of authentication
	AuthInternal = "internal"
	// RefreshToken is the refresh_token type of authentication
	RefreshToken = "refresh_token"

	argsLenError = "Invalid number of arguments. Expected %d, but got %d"
)

// CherwellClient represents a struct used for requestion the cherwell API
type CherwellClient struct {
	host  string
	token string
}

// NewClient returns a new CherwellClient
func NewClient(host string) *CherwellClient {
	return &CherwellClient{host: host}
}

// Authenticate authenticates the user and returns an AuthResponse or error
func (c *CherwellClient) Authenticate(authType string, args ...string) (AuthResponse, error) {
	var (
		vals    url.Values
		authRes AuthResponse
	)

	switch authType {
	case AuthInternal:
		if l := len(args); l < 3 {
			return authRes, fmt.Errorf(argsLenError, 4, l+1)
		}

		vals = url.Values{
			"grant_type": {"password"},
			"client_id":  {args[0]},
			"username":   {args[1]},
			"password":   {args[2]},
			"auth_mode":  {AuthInternal},
		}
	case RefreshToken:
		if l := len(args); l < 2 {
			return authRes, fmt.Errorf(argsLenError, 3, l+1)
		}

		vals = url.Values{
			"grant_type":    {RefreshToken},
			"client_id":     {args[0]},
			"refresh_token": {args[1]},
		}
	default:
		return authRes, errors.New("invalid Authentication Type")
	}

	resp, err := http.PostForm(c.host+"/token", vals)
	if err != nil {
		return authRes, err
	}

	if err := parseRequest(resp, &authRes); err != nil {
		return authRes, err
	}

	c.token = authRes.AccessToken

	return authRes, nil
}

// SaveBusinessObject returns a SaveBusObjResponse or an error
func (c *CherwellClient) SaveBusinessObject(bo BusinessObject) (SaveBusObjResponse, error) {
	var saveBusRes SaveBusObjResponse

	if err := c.requestAPI("/api/V1/savebusinessobject", http.MethodPost, bo, &saveBusRes); err != nil {
		return saveBusRes, err
	}

	return saveBusRes, nil
}

func (c *CherwellClient) requestAPI(endpoint, method string, payload, output interface{}) error {

	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, c.host+endpoint, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	return parseRequest(resp, output)
}
