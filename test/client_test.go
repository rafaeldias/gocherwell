package text

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	cherwell "github.com/rafaeldias/cherwell/client"
)

const (
	HOST          = "http://127.0.0.1"
	CLIENT_ID     = "0aa6e1b5-2280-4764-8d0c-d1922fc34180"
	USER          = "test_usr"
	PASS          = "test_usr"
	ACCESS_TOKEN  = "T2W0540_k99H78p1ibmX7eHaxLALX38DXWtCMf1brFypCvbc77KyOfKBPoVykw5F2CGVbuPk-F0QvtBwzLx8z-dcaRqoaM0L5qqBgH_NaMzL7DLL3ILmnEdlwfm4RRXNjnbwrgdz06vSMbUZ2ODcufOgipWhZpFCpaOCK4lQL11Oyj-2vCQhUNkTAdtBGReyxRdICFWDXiX4NVA08hfOd7VMGrWVOhtzFXhBRGFAWF4WwW4kLsWo_pK5b7sX_BqLGyLIm4w5Se2maY_eFhAaZZq39RivOhAYN8uLA6UzAE-LkOvHAjOWbP_W4gJLpnEdv42BXj0_jCZBCRmpXuHxmHSHlG3UhI3ZsgZa9ZrtTVCbUpNFe88PgZYYq0XzZ8nt2XxhzRCVhQE5bwZw5QM-LU86M4S6Pr99QZt2-64irmMq6lfJxcj13rbJH1xxfabMHRE1xLkBZAkNMpJlCnMw3sR57f9tJzHUKymbPL9WhLmu0FUzMPFjOg-SZhlQpmU9Ojmtym3btf0yOkfxLiR6gaRLYFHX1eDzrimy7NlCt2E"
	REFRESH_TOKEN = "423b345af13945afbe331169db8d0cd2"
)

func validPayload() cherwell.BusinessObject {
	return cherwell.BusinessObject{
		BusObID: "939ede4e7c0b06d3f7dbd248fc9edb20330dfc397c",
		Fields: []cherwell.BusinessObjectFields{
			cherwell.BusinessObjectFields{
				FieldID: "BO:939ede4e7c0b06d3f7dbd248fc9edb20330dfc397c,FI:939ede4f6a8e2735a0242e4c1aae97c9b295ba30b9",
				Name:    "Message",
				Value:   "Coloque aqui o e-mail",
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

			if r.Form.Get("username") != USER || r.Form.Get("password") != PASS || r.Form.Get("client_id") != CLIENT_ID {
				http.Error(w, `{"error":"invalid_grant","error_description":"BADREQUEST"}`, http.StatusBadRequest)
				return
			}
			w.Write([]byte(fmt.Sprintf(`{"access_token":"%s","token_type":"bearer","expires_in":1199,"refresh_token":"%s","as:client_id":"%s","username":"%s",".issued":"Mon, 04 Sep 2017 22:52:36 GMT",".expires":"Mon, 04 Sep 2017 23:12:36 GMT"}`, ACCESS_TOKEN, REFRESH_TOKEN, CLIENT_ID, USER)))
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

func TestAuthenticate_InternalWithNotEnoughArgs_ReturnsError(t *testing.T) {
	c := cherwell.NewClient(HOST)

	if _, err := c.Authenticate(cherwell.AUTH_INTERNAL, CLIENT_ID); err == nil {
		t.Fatalf("Expected err not to be nil, but got: %+v", err)
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
