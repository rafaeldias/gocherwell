package client

import "strconv"

type Error struct {
	StatusCode int
	Message    string
}

func (e Error) Error() string {
	return strconv.Itoa(e.StatusCode) + " " + e.Message
}
