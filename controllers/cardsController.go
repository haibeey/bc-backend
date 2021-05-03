package controllers

import (
	"bcchallenge/graph/model"
	"bcchallenge/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hasura/go-graphql-client"
)

func AddCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		model.NewResponse(w, 400, fmt.Errorf("Only post method allow"), struct{}{})
		return
	}

	username := checkToken(w, r)
	if len(username) <= 0 {
		return
	}

	cvv := utils.SpaceStringsBuilder(r.FormValue("cvv"))
	cardNo := utils.SpaceStringsBuilder(r.FormValue("cardno"))
	expiryDate := utils.SpaceStringsBuilder(r.FormValue("expirydate"))

	additionalData := make(map[string]string)
	additionalData["name"] = r.FormValue("name")
	additionalData["city"] = r.FormValue("city")
	additionalData["country"] = r.FormValue("country")
	additionalData["line1"] = r.FormValue("line1")
	additionalData["postalcode"] = r.FormValue("postalcode")
	additionalData["email"] = r.FormValue("email")
	additionalData["ip"] = r.RemoteAddr

	for k, v := range additionalData {
		if len(v) <= 0 {
			model.NewResponse(w, 400, fmt.Errorf("Invalid form data(%s not sent along)", k), struct{}{})
			return
		}
	}

	if len(cvv) != 3 || len(cardNo) < 15 || len(expiryDate) != 7 {
		model.NewResponse(w, 400, fmt.Errorf("Invalid form data"), struct{}{})
		return
	}

	if !utils.ContainsOnlyDigit(cvv) || !utils.ContainsOnlyDigit(cardNo) {
		model.NewResponse(w, 400, fmt.Errorf("Invalid card details"), struct{}{})
		return
	}
	appid, err := createCardOnCircle(model.Card{
		ID:         username,
		CardNo:     cardNo,
		ExpiryDate: expiryDate,
		Cvv:        cvv,
	}, additionalData)

	if err != nil {
		model.NewResponse(w, 500, fmt.Errorf("Error creating card on our partner"), err.Error())
		return
	}

	var m struct {
		Card struct {
			ID    graphql.String
			ApiID graphql.String `graphql:"apiID"`
		} `graphql:"createCard(input: $name)"`
	}

	var variables = map[string]interface{}{
		"name": model.NewCard{Title: "card",
			Owner:      username,
			CardNo:     cardNo,
			ExpiryDate: expiryDate,
			Cvv:        cvv,
			APIID:      appid,
		},
	}

	client := graphql.NewClient("http://localhost:8080/query", nil)
	err = client.Mutate(context.Background(), &m, variables)
	if err != nil {
		model.NewResponse(w, 500, fmt.Errorf("Error creating card"), err.Error())
		return
	}

	model.NewResponse(w, 200, fmt.Errorf("Success"), m)
}

func GetUserCards(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		model.NewResponse(w, 400, fmt.Errorf("Only GET method allow"), struct{}{})
		return
	}

	username := checkToken(w, r)
	if len(username) <= 0 {
		return
	}

	client := graphql.NewClient("http://localhost:8080/query", nil)

	var q struct {
		Cards []struct {
			ID graphql.String
		} `graphql:"userCards(userID: $name)"`
	}
	variables := map[string]interface{}{
		"name": graphql.String(username),
	}
	err := client.Query(context.Background(), &q, variables)
	if err != nil && err.Error() != "mongo: no documents in result" {
		model.NewResponse(w, 500, fmt.Errorf("Something went wrong while fetching data"), err.Error())
		return
	}

	model.NewResponse(w, 200, fmt.Errorf("Success"), q)
}

func createCardOnCircle(card model.Card, additionalData map[string]string) (string, error) {
	monthYear := strings.Split(card.ExpiryDate, "/")
	month, err := strconv.Atoi(monthYear[0])
	if err != nil {
		return "", err
	}
	year, err := strconv.Atoi(monthYear[1])
	if err != nil {
		return "", err
	}

	var data struct {
		IdempotencyKey string   `json:"idempotencyKey"`
		KeyId          string   `json:"keyId"`
		EncryptedData  string   `json:"encryptedData"`
		ExpMonth       int      `json:"expMonth"`
		ExpYear        int      `json:"expYear"`
		Metadata       meta     `json:"metadata"`
		BillingDetails bDetails `json:"billingDetails"`
	}

	data.ExpMonth = month
	data.ExpYear = year
	var cardNoCvv struct {
		Number string `json:"number"`
		Cvv    string `json:"cvv"`
	}

	cardNoCvv.Number = card.CardNo
	cardNoCvv.Cvv = card.Cvv
	cardNoCvvbyte, err := json.Marshal(cardNoCvv)
	if err != nil {
		return "", err
	}
	keyMap, err := GetPublicKey()
	if err != nil {
		return "", err
	}

	data.IdempotencyKey = utils.GenerateIdempotentKey()
	data.KeyId = keyMap["keyID"]

	encData, err := utils.PGPEncrypt1(string(cardNoCvvbyte), keyMap["publicKey"])
	if err != nil {
		return "", err
	}
	data.EncryptedData = encData

	data.Metadata = meta{
		Email:     additionalData["email"],
		SessionId: fmt.Sprintf("%d", time.Now().Unix()),
		IpAddress: "102.67.18.203",
	}

	data.BillingDetails = bDetails{
		Name:       additionalData["name"],
		City:       additionalData["city"],
		Country:    additionalData["country"],
		Line1:      additionalData["line1"],
		PostalCode: additionalData["postalcode"],
	}

	res, err := json.MarshalIndent(data, "", "	")
	if err != nil {
		return "", err
	}
	response, err := makePostRequest("https://api-sandbox.circle.com/v1/cards", res)
	if err != nil {
		return "", err
	}

	code, ok := response["code"].(int)
	if !ok {
		return "", fmt.Errorf("Invalid Response format ")
	}

	if code != 201 {
		return "", fmt.Errorf("Error creating card got " + response["raw"].(string))
	}

	appid, ok := response["data"].(map[string]interface{})["id"].(string)
	if !ok {
		return "", fmt.Errorf("Invalid Response format ")
	}
	return appid, nil
}
