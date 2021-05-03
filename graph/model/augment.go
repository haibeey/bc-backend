package model

import (
	"encoding/json"
	"net/http"
)

func (u *User) GetID() string {
	return u.ID
}

func (u *User) SetID(id string) {
	u.ID = id
}

func (t *Transaction) GetID() string {
	return t.ID
}

func (t *Transaction) SetID(id string) {
	t.ID = id
}

func (c *Card) GetID() string {
	return c.ID
}

func (c *Card) SetID(id string) {
	c.ID = id
}

// NewResponse example
func NewResponse(w http.ResponseWriter, status int, err error, data interface{}) {
	er := HTTPRes{
		Code:    status,
		Message: err.Error(),
		Data:    data,
	}
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(er)
}

// HTTPRes example
type HTTPRes struct {
	Code    int         `json:"code" example:""`
	Message string      `json:"message" example:"status bad request"`
	Data    interface{} `json:"data"`
}
