package gocherwell

import (
	"encoding/json"
	"strconv"
	"time"
)

// AuthResponse represents the authorization response from cherwell
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_typoe"`
	ExpiresIn    *Duration `json:"expires_in"`
	RefreshToken string    `json:"refresh_token"`
}

// BusinessObject represents a business object payload to the cherwell
type BusinessObject struct {
	BusObID       string                 `json:"busObId"`
	BusObRecID    string                 `json:"busObRecId"`
	BusObPublicID string                 `json:"busObPublicId"`
	Fields        []BusinessObjectFields `json:"fields"`
}

// BusinessObjectFields represents the fields of a business object payload
type BusinessObjectFields struct {
	FieldID string `json:"fieldId"`
	Name    string `json:"name"`
	Value   string `json:"value"`
	Dirty   bool   `json:"dirty"`
}

// Duration represents the time returned by the authentication request
type Duration struct {
	time.Duration
}

// UnmarshalJSON meets the interface Unmarshaler of json package
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s int

	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	d.Duration = time.Duration(s) * time.Second

	return nil
}

// MarshalJSON meets the interface Marshaler of json package
func (d *Duration) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Itoa(int(d.Duration / time.Second))), nil
}

// Error is returned and a error occurred by calling the cherwell API
type Error struct {
	StatusCode int
	Message    string
}

func (e Error) Error() string {
	return strconv.Itoa(e.StatusCode) + " " + e.Message
}

// SaveBusObjResponse represents the response returned by cherwell
// as a result of calling SaveBusinessObject
type SaveBusObjResponse struct {
	BusObPublicID string `json:"busObPublicId"`
	BusObRecID    string `json:"busObRecId"`
}
