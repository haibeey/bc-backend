package models

import (
	"bcchallenge/graph/model"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func ToUserFromBson(mongoM bson.M) (*model.User, error) {
	uB, err := bson.Marshal(mongoM)
	if err != nil {
		return nil, err
	}
	p := &model.User{}
	err = bson.Unmarshal(uB, p)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func ToUsersFromBsons(mongoM []bson.M) ([]*model.User, error) {
	p := []*model.User{}
	for _, uM := range mongoM {
		u, err := ToUserFromBson(uM)
		if err != nil {
			return nil, err
		}
		p = append(p, u)
	}
	return p, nil
}

func Create(u *model.User) error {
	hashedPassword, err := HashPassword(u.Password)
	if err != nil {
		return err
	}
	u.Password = hashedPassword
	return Insert(u, "user")
}

func GetUserIdByUsername(username string) (string, error) {
	uM, err := FetchDocByCriterion("name", username, "user")
	if err != nil {
		return "", err
	}
	u, err := ToUserFromBson(uM)
	if err != nil {
		return "", err
	}
	return u.ID, nil
}

func Authenticate(u *model.User) bool {
	uM, err := FetchDocByCriterion("name", u.Name, "user")
	if err != nil {
		return false
	}
	uu, err := ToUserFromBson(uM)
	if err != nil {
		return false
	}

	return CheckPasswordHash(u.Password, uu.Password)
}

//HashPassword hashes given password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

//CheckPassword hash compares raw password with it's hashed values
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
