package models

import (
	"bcchallenge/graph/model"
	"go.mongodb.org/mongo-driver/bson"
)

func ToTransactionFromBson(mongoM bson.M) (*model.Transaction, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	p := &model.Transaction{}
	err = bson.Unmarshal(uB, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ToTransactionsFromBsons(mongoM []bson.M) ([]*model.Transaction, error) {
	p := []*model.Transaction{}
	for _, uM := range mongoM {
		u, err := ToTransactionFromBson(uM)
		if err != nil {
			return nil, err
		}
		p = append(p, u)
	}
	return p, nil
}
