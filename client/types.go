package client

import (
	"encoding/json"
	"strconv"
	"time"
)

type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_typoe"`
	ExpiresIn    *Duration `json:"expires_in"`
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

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s int

	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	d.Duration = time.Duration(s) * time.Second

	return nil
}

func (d *Duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(d.Duration / time.Second))), nil
}
