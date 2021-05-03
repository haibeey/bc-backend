package controllers

import (
	"bcchallenge/graph/model"
	"bcchallenge/models"
	"bcchallenge/utils"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/hasura/go-graphql-client"
)

type payment struct {
	Metadata meta `json:"metadata"`
	Amount   struct {
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
	} `json:"amount"`

	Source struct {
		Id string `json:"id"`
		Type string `json:"type"`
	} `json:"source"`

	EncryptedData  string `json:"encryptedData"`  
	IdempotencyKey string `json:"idempotencyKey"` 
	KeyId          string `json:"keyId"`          
	Verification   string `json:"verification"`   
	Description    string `json:"description"`
}

func BillCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		model.NewResponse(w, 400, fmt.Errorf("Only post method allow"), struct{}{})
		return
	}
	username := checkToken(w, r)
	if len(username) <= 0 {
		return
	}

	keyMap, err := GetPublicKey()
	if err != nil {
		model.NewResponse(w, 400, fmt.Errorf("Couldn't get data from payment partner"), struct{}{})
		return
	}

	amountStr := utils.SpaceStringsBuilder(r.FormValue("amount"))
	cardid := utils.SpaceStringsBuilder(r.FormValue("cardid"))

	formData := make(map[string]string)
	formData["amount"] = r.FormValue("amount")
	formData["currency"] = r.FormValue("currency")
	formData["email"] = r.FormValue("email")
	formData["description"] = r.FormValue("description")
	formData["cardid"] = cardid
	formData["amount"] = amountStr

	for k, v := range formData {
		if len(v) <= 0 {
			model.NewResponse(w, 400, fmt.Errorf("Invalid form data(%s not sent along)", k), struct{}{})
			return
		}
	}

	p := payment{
		Metadata: meta{
			IpAddress: "102.67.18.203",
			Email:     formData["email"],
			SessionId: fmt.Sprintf("%d", time.Now().Unix()),
		},
		Description: formData["description"],
	}
	p.Amount.Amount = formData["amount"]
	p.Amount.Currency = formData["currency"]

	if len(amountStr) <= 0 || len(cardid) <= 0 {
		model.NewResponse(w, 400, fmt.Errorf("Invalid  form data(s)"), struct{}{})
		return
	}
	if _, err := strconv.ParseFloat(amountStr, 32); err != nil {
		model.NewResponse(w, 400, fmt.Errorf("Invalid  form amount"), err.Error())
		return
	}
	cardM, err := models.FetchDocByCriterion("id", cardid, "card")
	if err != nil {
		model.NewResponse(w, 400, fmt.Errorf("Couldn't get card"), err.Error())
		return
	}
	card, err := models.ToCardFromBson(cardM)
	if err != nil {
		model.NewResponse(w, 400, fmt.Errorf("Couldn't process card"), err.Error())
		return
	}
	p.EncryptedData = card.Cvv
	p.Source.Id = card.APIID
	p.Source.Type = "card"

	appid, err := makePaymentOnCircle(p, keyMap)
	if err != nil {
		model.NewResponse(w, 500, fmt.Errorf("Error making payment from partner"), err.Error())
		return
	}

	var m struct {
		Transaction struct {
			ID    graphql.String
			ApiID graphql.String `graphql:"apiID"`
		} `graphql:"createTransaction(input: $name)"`
	}

	var variables = map[string]interface{}{
		"name": model.NewTransanction{
			DebitedOrCredited: true,
			APIID:             appid,
			Card:              card.ID,
			Time:              time.Now().String(),
			Amount:            p.Amount.Amount,
		},
	}

	client := graphql.NewClient(os.Getenv("GRAPHQLURL")+"/query", nil)
	err = client.Mutate(context.Background(), &m, variables)
	if err != nil {
		model.NewResponse(w, 500, fmt.Errorf("Error finishing up transaction"), err.Error())
		return
	}

	model.NewResponse(w, 200, fmt.Errorf("Success"), m)

}

func makePaymentOnCircle(p payment, keyMap map[string]string) (string, error) {
	p.IdempotencyKey = utils.GenerateIdempotentKey()
	p.Verification = "none"
	p.KeyId = keyMap["keyID"]
	encData, err := utils.PGPEncrypt1(p.EncryptedData, keyMap["publicKey"])
	if err != nil {
		return "", err
	}
	p.EncryptedData = encData

	res, err := json.MarshalIndent(p, "", "	")
	if err != nil {
		return "", err
	}

	response, err := makePostRequest("https://api-sandbox.circle.com/v1/payments", res)
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
