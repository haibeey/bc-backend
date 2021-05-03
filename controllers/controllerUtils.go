package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"bcchallenge/graph/model"
	"bcchallenge/utils"

	_ "github.com/hasura/go-graphql-client"
)

type meta struct {
	Email     string `json:"email"`
	SessionId string `json:"sessionId"`
	IpAddress string `json:"ipAddress"`
}

type bDetails struct {
	Name       string `json:"name"`
	City       string `json:"city"`
	Country    string `json:"country"`
	Line1      string `json:"line1"`
	PostalCode string `json:"postalCode"`
}

func checkToken(w http.ResponseWriter, r *http.Request) string {
	var header = r.Header.Get("Authorization")
	if len(header) <= 0 {
		model.NewResponse(w, 400, fmt.Errorf("No bearer token sent"), struct{}{})
		return ""
	}

	headersContent := strings.Split(header, " ")
	if len(headersContent) != 2 {
		model.NewResponse(w, 400, fmt.Errorf("Invalid token sent"), struct{}{})
		return ""
	}
	username, err := utils.ParseToken(headersContent[1])
	if err != nil {
		model.NewResponse(w, 400, fmt.Errorf("Invalid token sent(Error parsing token)"), err.Error())
		return ""
	}

	return username
}
func makeGetRequest(url string, val url.Values) (map[string]interface{}, error) {
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("CIRCLEAPIKEY")))
	var data map[string]interface{}
	client := &http.Client{Timeout: time.Second * 10}
	response, err := client.Do(request)
	if err != nil {
		return data, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return data, err
	}
	defer response.Body.Close()
	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func makePostRequest(url string, body []byte) (map[string]interface{}, error) {
	request, err := http.NewRequest("POST", url, nil)
	request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", os.Getenv("CIRCLEAPIKEY")))
	request.Header.Add("Content-Type", "application/json")
	request.Header.Add("Accept", "application/json")
	request.Body = ioutil.NopCloser(bytes.NewReader(body))

	var data map[string]interface{}
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return data, err
	}
	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return data, err
	}
	defer response.Body.Close()
	err = json.Unmarshal(body, &data)
	if err != nil {
		return data, err
	}

	data["code"] = response.StatusCode
	data["raw"] = string(body)
	return data, nil
}

func getMyWalletID() (string, error) {
	data, err := makeGetRequest("https://api-sandbox.circle.com/v1/configuration", url.Values{})
	if err != nil {
		return "", err
	}
	id, ok := data["data"].(map[string]interface{})["payments"].(map[string]interface{})["masterWalletId"]
	if !ok {
		return "", fmt.Errorf("Bad data format from api")
	}
	return id.(string), nil
}

func GetPublicKey() (map[string]string, error) {
	data, err := makeGetRequest("https://api-sandbox.circle.com/v1/encryption/public", url.Values{})
	if err != nil {
		return nil, err
	}
	data, ok := data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("Bad data format from api")
	}

	result := make(map[string]string)
	key, ok := data["keyId"].(string)
	if !ok {
		return nil, fmt.Errorf("Bad data format from api")
	}
	result["keyID"] = key

	key, ok = data["publicKey"].(string)
	if !ok {
		return nil, fmt.Errorf("Bad data format from api")
	}
	result["publicKey"] = key

	return result, nil
}
