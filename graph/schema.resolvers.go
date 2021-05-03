package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"bcchallenge/graph/generated"
	"bcchallenge/graph/model"
	bcmodel "bcchallenge/models"
	"bcchallenge/utils"
	"context"
	"fmt"
)

func (r *mutationResolver) CreateUser(ctx context.Context, input model.NewUser) (string, error) {
	h, err := bcmodel.HashPassword(input.Password)
	if err != nil {
		return h, err
	}
	u := model.User{Name: input.Username, Password: h}
	err = bcmodel.Insert(&u, "user")
	if err != nil {
		return "", err
	}
	return utils.GenerateToken(u.Name)
}

func (r *mutationResolver) Login(ctx context.Context, input model.Login) (string, error) {
	correct := bcmodel.Authenticate(&model.User{Name: input.Username, Password: input.Password})
	if !correct {
		return "", fmt.Errorf("Password does not match")
	}
	token, err := utils.GenerateToken(input.Username)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (r *mutationResolver) RefreshToken(ctx context.Context, input model.RefreshTokenInput) (string, error) {
	username, err := utils.ParseToken(input.Token)
	if err != nil {
		return "", fmt.Errorf("access denied")
	}
	token, err := utils.GenerateToken(username)
	if err != nil {
		return "", err
	}
	return token, nil
}

func (r *mutationResolver) CreateCard(ctx context.Context, input model.NewCard) (*model.Card, error) {
	c := model.Card{
		CardNo:     input.CardNo,
		Cvv:        input.Cvv,
		ExpiryDate: input.ExpiryDate,
		Owner:      input.Owner,
		APIID:      input.APIID,
	}
	err := bcmodel.Insert(&c, "card")
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *mutationResolver) CreateTransaction(ctx context.Context, input model.NewTransanction) (*model.Transaction, error) {
	t := model.Transaction{
		Card:              input.Card,
		Time:              input.Time,
		Amount:            input.Amount,
		DebitedOrCredited: input.DebitedOrCredited,
		APIID:             input.APIID,
	}
	err := bcmodel.Insert(&t, "transaction")
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *queryResolver) Users(ctx context.Context) ([]*model.User, error) {
	allUser, err := bcmodel.FetchAll("user")
	if err != nil {
		return nil, err
	}
	return bcmodel.ToUsersFromBsons(allUser)
}

func (r *queryResolver) FindUser(ctx context.Context, name string) (*model.User, error) {
	uM, err := bcmodel.FetchDocByCriterion("name", name, "user")
	if err != nil {
		return nil, err
	}
	return bcmodel.ToUserFromBson(uM)
}

func (r *queryResolver) Cards(ctx context.Context) ([]*model.Card, error) {
	cards, err := bcmodel.FetchAll("card")
	if err != nil {
		return nil, err
	}
	return bcmodel.ToCardsFromBsons(cards)
}

func (r *queryResolver) Transactions(ctx context.Context) ([]*model.Transaction, error) {
	transactions, err := bcmodel.FetchAll("transaction")
	if err != nil {
		return nil, err
	}
	return bcmodel.ToTransactionsFromBsons(transactions)
}

func (r *queryResolver) UserCards(ctx context.Context, userID string) ([]*model.Card, error) {
	cards, err := bcmodel.FetchDocByCriterionMultipleRes(
		map[string]string{"owner": userID},
		"card",
	)
	if err != nil {
		return nil, err
	}
	return bcmodel.ToCardsFromBsons(cards)
}

func (r *queryResolver) UserTransaction(ctx context.Context, userID string) ([]*model.Transaction, error) {
	transactions, err := bcmodel.FetchDocByCriterionMultipleRes(
		map[string]string{"owner": userID},
		"transaction",
	)
	if err != nil {
		return nil, err
	}
	return bcmodel.ToTransactionsFromBsons(transactions)
}

func (r *queryResolver) CardTransaction(ctx context.Context, cardID string) ([]*model.Transaction, error) {
	transactions, err := bcmodel.FetchDocByCriterionMultipleRes(
		map[string]string{"card": cardID},
		"transaction",
	)
	if err != nil {
		return nil, err
	}
	return bcmodel.ToTransactionsFromBsons(transactions)
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
