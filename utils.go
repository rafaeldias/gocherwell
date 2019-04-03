package gocherwell

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func unmarshalBody(resp *http.Response, output interface{}) error {
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
