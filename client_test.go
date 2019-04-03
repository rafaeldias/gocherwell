package gocherwell

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	host         = "http://127.0.0.1"
	clientID     = "{client_id}"
	user         = "test_usr"
	pass         = "test_usr"
	accessToken  = "{access_token}"
	refreshToken = "{refresh_token}"
)

func validPayload() BusinessObject {
	return BusinessObject{
		BusObID: "{busObjID}",
		Fields: []BusinessObjectFields{
			BusinessObjectFields{
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

				if r.Form.Get("username") != user || r.Form.Get("password") != pass || r.Form.Get("client_id") != clientID {
					http.Error(w, `{"error":"invalid_grant","error_description":"BADREQUEST"}`, http.StatusBadRequest)
					return
				}
			case "refresh_token":
				if r.Form.Get("client_id") != clientID || r.Form.Get("refresh_token") != refreshToken {
					http.Error(w, `{"error":"invalid_grant","error_description":"BADREQUEST"}`, http.StatusBadRequest)
					return
				}
			}

			res, err := json.Marshal(AuthResponse{
				AccessToken:  accessToken,
				TokenType:    "bearer",
				ExpiresIn:    &Duration{time.Duration(1199) * time.Second},
				RefreshToken: refreshToken,
			})
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(res)
		case "/api/V1/savebusinessobject":
			var pieces = strings.Fields(r.Header.Get("Authorization"))

			if len(pieces) != 2 || pieces[1] != accessToken {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			res, err := json.Marshal(SaveBusObjResponse{BusObPublicID: "xyz"})
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(res)
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
}

func TestNewClient_WithEmptyHost_ReturnsValidObject(t *testing.T) {
	c := NewClient("")

	if c == nil {
		t.Fatalf("Expected NewClient to not return nil")
	}
}

func TestNewClient_WithValidHost_ReturnsValidObject(t *testing.T) {
	c := NewClient(host)

	if c == nil {
		t.Fatalf("Expected NewClient to not return nil")
	}
}

func TestAuthenticate_WithValidCredentials_ReturnsAccessToken(t *testing.T) {
	s := cherwellV1Server()

	defer s.Close()

	c := NewClient(s.URL)

	if res, _ := c.Authenticate(AuthInternal, clientID, user, pass); res.AccessToken != accessToken {
		t.Fatalf("Expected AccessToken to be %s, but got: %s", accessToken, res.AccessToken)
	}
}

func TestAuthenticate_RefreshTokenWithValidCredentials_ReturnsNewAccessToken(t *testing.T) {
	s := cherwellV1Server()

	defer s.Close()

	c := NewClient(s.URL)

	if res, _ := c.Authenticate(RefreshToken, clientID, refreshToken); res.AccessToken != accessToken {
		t.Fatalf("Expected AccessToken to be %s, but got: %s", accessToken, res.AccessToken)
	}
}

func TestAuthenticate_InternalWithNotEnoughArgs_ReturnsError(t *testing.T) {
	c := NewClient(host)

	if _, err := c.Authenticate(AuthInternal, clientID); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
	}
}

func TestAuthenticate_RefreshTokenWithNotEnoughArgs_ReturnsError(t *testing.T) {
	c := NewClient(host)

	if _, err := c.Authenticate(RefreshToken, clientID); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
	}
}

func TestAuthenticate_WithInvalidAuthType_ReturnsError(t *testing.T) {
	s := cherwellV1Server()

	defer s.Close()

	c := NewClient(s.URL)

	if _, err := c.Authenticate("Invalid_Auth_Type"); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
	}
}

func TestAuthenticate_WithInvalidAuthGrant_ReturnsErrorCode400(t *testing.T) {
	s := cherwellV1Server()

	defer s.Close()

	c := NewClient(s.URL)

	if _, err := c.Authenticate(AuthInternal, clientID, user, "123"); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
	} else if e, ok := err.(Error); ok {
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

	c := NewClient(s.URL)

	c.Authenticate(AuthInternal, clientID, user, pass)

	if res, _ := c.SaveBusinessObject(p); res.BusObPublicID != "xyz" {
		t.Fatalf("Expected BusObPublicID to be `xyz`, but got: %s", res.BusObPublicID)
	}
}

func TestSaveBusinessObject_WithoutAccessToken_ReturnsErrorCode401(t *testing.T) {
	p := validPayload()
	s := cherwellV1Server()

	defer s.Close()

	c := NewClient(s.URL)

	if _, err := c.SaveBusinessObject(p); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
	} else if e, ok := err.(Error); ok {
		if e.StatusCode != http.StatusForbidden {
			t.Fatalf("Expected err status code to be %d, but got: %d", http.StatusForbidden, e.StatusCode)
		}
	} else {
		t.Fatalf("Expected err to be an instance of cherwell.Error, but got: %+v", err)
	}
}
