package gocherwell

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func parseRequest(resp *http.Response, output interface{}) error {
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return Error{resp.StatusCode, string(body)}
	}

	return json.Unmarshal(body, output)
}
