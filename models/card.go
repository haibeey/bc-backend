package models

import (
	"bcchallenge/graph/model"
	"go.mongodb.org/mongo-driver/bson"
)

func ToCardFromBson(mongoM bson.M) (*model.Card, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	p := &model.Card{}
	err = bson.Unmarshal(uB, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ToCardsFromBsons(mongoM []bson.M) ([]*model.Card, error) {
	p := []*model.Card{}
	for _, uM := range mongoM {
		u, err := ToCardFromBson(uM)
		if err != nil {
			return nil, err
		}
		p = append(p, u)
	}
	return p, nil
}
