package controllers

import (
	"bcchallenge/graph/model"
	"context"
	"fmt"
	"net/http"

	"github.com/hasura/go-graphql-client"
)

func SignUp(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		model.NewResponse(w, 400, fmt.Errorf("Only post method allow"), struct{}{})
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	if len(username) < 0 || len(password) <= 0 {
		model.NewResponse(w, 400, fmt.Errorf("No username or password in request"), struct{}{})
		return
	}

	client := graphql.NewClient("http://localhost:8080/query", nil)
	var q struct {
		FindUser struct {
			Name graphql.String
		} `graphql:"findUser(name: $name)"`
	}
	variables := map[string]interface{}{
		"name": graphql.String(username),
	}
	err := client.Query(context.Background(), &q, variables)
	if err != nil && err.Error() != "mongo: no documents in result" {
		model.NewResponse(w, 500, fmt.Errorf("Something went wrong while fetching data"), err.Error())
		return
	}

	if q.FindUser.Name == graphql.String(username) {
		model.NewResponse(w, 400, fmt.Errorf("Username taken"), struct{}{})
		return
	}

	var m struct {
		CreateUser string `graphql:"createUser(input: $name)"`
	}

	variables = map[string]interface{}{
		"name": model.NewUser{Username: username, Password: password},
	}

	err = client.Mutate(context.Background(), &m, variables)
	if err != nil {
		model.NewResponse(w, 500, fmt.Errorf("Error creating user"), err.Error())
		return
	}

	model.NewResponse(w, 200, fmt.Errorf("Success"), m)

}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		model.NewResponse(w, 400, fmt.Errorf("Only post method allow"), struct{}{})
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	if len(username) < 0 || len(password) <= 0 {
		model.NewResponse(w, 400, fmt.Errorf("No username or password in request"), struct{}{})
		return
	}

	client := graphql.NewClient("http://localhost:8080/query", nil)

	var m struct {
		CreateUser string `graphql:"login(input: $name)"`
	}

	var variables = map[string]interface{}{
		"name": model.Login{Username: username, Password: password},
	}

	err := client.Mutate(context.Background(), &m, variables)
	if err != nil {
		model.NewResponse(w, 400, fmt.Errorf("Error loggin in"), err.Error())
		return
	}

	model.NewResponse(w, 200, fmt.Errorf("Success"), m)
}
