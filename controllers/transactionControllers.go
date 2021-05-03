package controllers

import (
	"bcchallenge/graph/model"
	"bcchallenge/utils"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/hasura/go-graphql-client"
)

func AddTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		model.NewResponse(w, 400, fmt.Errorf("Only post method allow"), struct{}{})
		return
	}

	if len(checkToken(w, r)) <= 0 {
		return
	}

	amountStr := utils.SpaceStringsBuilder(r.FormValue("amount"))
	cardid := utils.SpaceStringsBuilder(r.FormValue("cardid"))

	if len(amountStr) <= 0 || len(cardid) <= 0 {
		model.NewResponse(w, 400, fmt.Errorf("Invalid  form data(s)"), struct{}{})
		return
	}

	if _, err := strconv.ParseFloat(amountStr, 32); err != nil {
		model.NewResponse(w, 400, fmt.Errorf("Invalid  form amount"), err.Error())
		return
	}

	var m struct {
		Transanction struct {
			ID     graphql.String
			Amount graphql.String
			Card   graphql.String
		} `graphql:"createTransaction(input: $name)"`
	}

	var variables = map[string]interface{}{
		"name": model.NewTransanction{
			Time:   fmt.Sprintf("%d", time.Now().Unix()),
			Card:   cardid,
			Amount: amountStr,
		},
	}

	client := graphql.NewClient("http://localhost:8080/query", nil)
	err := client.Mutate(context.Background(), &m, variables)
	if err != nil {
		model.NewResponse(w, 500, fmt.Errorf("Error creating card"), err.Error())
		return
	}

	model.NewResponse(w, 200, fmt.Errorf("Success"), m)
}

func GetCardTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		model.NewResponse(w, 400, fmt.Errorf("Only GET method allow"), struct{}{})
		return
	}

	if len(checkToken(w, r)) <= 0 {
		return
	}

	cardid := utils.SpaceStringsBuilder(r.FormValue("cardid"))

	if len(cardid) <= 0 {
		model.NewResponse(w, 400, fmt.Errorf("Invalid  form data(s)"), struct{}{})
		return
	}

	client := graphql.NewClient("http://localhost:8080/query", nil)

	var q struct {
		Transactions []struct {
			ID     graphql.String
			Card   graphql.String
			Amount graphql.String
		} `graphql:"cardTransaction(cardID: $name)"`
	}
	variables := map[string]interface{}{
		"name": graphql.String(cardid),
	}
	err := client.Query(context.Background(), &q, variables)
	if err != nil && err.Error() != "mongo: no documents in result" {
		model.NewResponse(w, 500, fmt.Errorf("Something went wrong while fetching data"), err.Error())
		return
	}

	model.NewResponse(w, 200, fmt.Errorf("Success"), q)
}
