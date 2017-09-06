package text

import (
	"encoding/json"
	_ "fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	cherwell "github.com/rafaeldias/gocherwell/client"
)

const (
	HOST          = "http://127.0.0.1"
	CLIENT_ID     = "{client_id}"
	USER          = "test_usr"
	PASS          = "test_usr"
	ACCESS_TOKEN  = "{access_token}"
	REFRESH_TOKEN = "{refresh_token}"
)

func validPayload() cherwell.BusinessObject {
	return cherwell.BusinessObject{
		BusObID: "{busObjID}",
		Fields: []cherwell.BusinessObjectFields{
			cherwell.BusinessObjectFields{
				FieldID: "{field_id}",
				Name:    "Name",
				Value:   "Value",
				Dirty:   true,
			},
		},
	}
}

func cherwellV1Server() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.RequestURI {
		case "/token":
			r.ParseForm()

			switch r.Form.Get("grant_type") {
			case "password":

				if r.Form.Get("username") != USER || r.Form.Get("password") != PASS || r.Form.Get("client_id") != CLIENT_ID {
					http.Error(w, `{"error":"invalid_grant","error_description":"BADREQUEST"}`, http.StatusBadRequest)
					return
				}
			case "refresh_token":
				if r.Form.Get("client_id") != CLIENT_ID || r.Form.Get("refresh_token") != REFRESH_TOKEN {
					http.Error(w, `{"error":"invalid_grant","error_description":"BADREQUEST"}`, http.StatusBadRequest)
					return
				}
			}

			authRes := cherwell.AuthResponse{
				AccessToken:  ACCESS_TOKEN,
				TokenType:    "bearer",
				ExpiresIn:    &cherwell.Duration{time.Duration(1199) * time.Second},
				RefreshToken: REFRESH_TOKEN,
			}

			if res, err := json.Marshal(authRes); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write(res)
			}
		case "/api/V1/savebusinessobject":
			var pieces = strings.Fields(r.Header.Get("Authorization"))

			if len(pieces) != 2 || pieces[1] != ACCESS_TOKEN {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			busObjRes := cherwell.SaveBusObjResponse{BusObPublicID: "xyz"}

			if res, err := json.Marshal(busObjRes); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write(res)
			}
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
}

func TestNewClient_WithEmptyHost_ReturnsValidObject(t *testing.T) {
	c := cherwell.NewClient("")

	if c == nil {
		t.Fatalf("Expected NewClient to not return nil")
	}
}

func TestNewClient_WithValidHost_ReturnsValidObject(t *testing.T) {
	c := cherwell.NewClient(HOST)

	if c == nil {
		t.Fatalf("Expected NewClient to not return nil")
	}
}

func TestAuthenticate_WithValidCredentials_ReturnsAccessToken(t *testing.T) {
	s := cherwellV1Server()

	defer s.Close()

	c := cherwell.NewClient(s.URL)

	if res, _ := c.Authenticate(cherwell.AUTH_INTERNAL, CLIENT_ID, USER, PASS); res.AccessToken != ACCESS_TOKEN {
		t.Fatalf("Expected AccessToken to be %s, but got: %s", ACCESS_TOKEN, res.AccessToken)
	}
}

func TestAuthenticate_RefreshTokenWithValidCredentials_ReturnsNewAccessToken(t *testing.T) {
	s := cherwellV1Server()

	defer s.Close()

	c := cherwell.NewClient(s.URL)

	if res, _ := c.Authenticate(cherwell.REFRESH_TOKEN, CLIENT_ID, REFRESH_TOKEN); res.AccessToken != ACCESS_TOKEN {
		t.Fatalf("Expected AccessToken to be %s, but got: %s", ACCESS_TOKEN, res.AccessToken)
	}
}

func TestAuthenticate_InternalWithNotEnoughArgs_ReturnsError(t *testing.T) {
	c := cherwell.NewClient(HOST)

	if _, err := c.Authenticate(cherwell.AUTH_INTERNAL, CLIENT_ID); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
	}
}

func TestAuthenticate_RefreshTokenWithNotEnoughArgs_ReturnsError(t *testing.T) {
	c := cherwell.NewClient(HOST)

	if _, err := c.Authenticate(cherwell.REFRESH_TOKEN, CLIENT_ID); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
	}
}

func TestAuthenticate_WithInvalidAuthType_ReturnsError(t *testing.T) {
	s := cherwellV1Server()

	defer s.Close()

	c := cherwell.NewClient(s.URL)

	if _, err := c.Authenticate("Invalid_Auth_Type"); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
	}
}

func TestAuthenticate_WithInvalidAuthGrant_ReturnsErrorCode400(t *testing.T) {
	s := cherwellV1Server()

	defer s.Close()

	c := cherwell.NewClient(s.URL)

	if _, err := c.Authenticate(cherwell.AUTH_INTERNAL, CLIENT_ID, USER, "123"); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
	} else if e, ok := err.(cherwell.Error); ok {
		if e.StatusCode != http.StatusBadRequest {
			t.Fatalf("Expected err status code to be %d, but got: %d", http.StatusBadRequest, e.StatusCode)
		}
	} else {
		t.Fatalf("Expected err to be an instance of cherwell.Error, but got: %+v", err)
	}
}

func TestSaveBusinessObject_WithValidPayloadAndToken_ReturnsSaveBusObjResponse(t *testing.T) {
	p := validPayload()
	s := cherwellV1Server()

	defer s.Close()

	c := cherwell.NewClient(s.URL)

	c.Authenticate(cherwell.AUTH_INTERNAL, CLIENT_ID, USER, PASS)

	if res, _ := c.SaveBusinessObject(p); res.BusObPublicID != "xyz" {
		t.Fatalf("Expected BusObPublicID to be `xyz`, but got: %s", res.BusObPublicID)
	}
}

func TestSaveBusinessObject_WithoutAccessToken_ReturnsErrorCode401(t *testing.T) {
	p := validPayload()
	s := cherwellV1Server()

	defer s.Close()

	c := cherwell.NewClient(s.URL)

	if _, err := c.SaveBusinessObject(p); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
	} else if e, ok := err.(cherwell.Error); ok {
		if e.StatusCode != http.StatusForbidden {
			t.Fatalf("Expected err status code to be %d, but got: %d", http.StatusForbidden, e.StatusCode)
		}
	} else {
		t.Fatalf("Expected err to be an instance of cherwell.Error, but got: %+v", err)
	}
}
