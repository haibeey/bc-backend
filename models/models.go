package models

import (
	"bcchallenge/database"
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DocModel interface {
	GetID() string
	SetID(id string)
}

//Insert insert into the db
func Insert(doc DocModel, collectionName string) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	result, err := collection.InsertOne(context.TODO(), doc)
	if err != nil {
		return err
	}
	doc.SetID(result.InsertedID.(primitive.ObjectID).Hex())
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "id", Value: doc.GetID()}}}}
	err = Update(doc, update, collectionName)
	return err
}

//Update update a doc in the database
func Update(doc DocModel, update interface{}, collectionName string) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	s, err := primitive.ObjectIDFromHex(doc.GetID())
	if err != nil {
		return err
	}
	filter := bson.M{"_id": s}

	opts := options.Update().SetUpsert(false)

	_, err = collection.UpdateOne(context.TODO(), filter, update, opts)
	return err
}

//Deleter remove a doc from the db
func Delete(doc DocModel, collectionName string) error {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	s, err := primitive.ObjectIDFromHex(doc.GetID())
	if err != nil {
		return err
	}
	filter := bson.M{"_id": s}

	opts := options.Delete().SetCollation(&options.Collation{
		Locale:    "en_US",
		Strength:  1,
		CaseLevel: false,
	})

	_, err = collection.DeleteOne(context.TODO(), filter, opts)
	return err
}

//FetchDocByCriterion returns a  struct that tha matches the particular criteria
// i.e FetchDocByCriterion("username","abraham","user") returns a user struct where username is abraham
func FetchDocByCriterion(criteria, value, collectionName string) (bson.M, error) {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	filter := bson.M{criteria: value}
	doc := bson.M{}

	err := collection.FindOne(context.TODO(), filter).Decode(doc)

	if err != nil {
		return nil, err
	}
	return doc, nil
}

//FetchDocByCriterionMultipleRes
func FetchDocByCriterionMultipleRes(criterias map[string]string, collectionName string) ([]bson.M, error) {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	filter := bson.M{}

	for k, v := range criterias {
		filter[k] = v
	}
	docs := []bson.M{}
	opts := options.Find().SetSort(bson.D{{Key: "CreatedAt", Value: 1}})
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &docs); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return docs, nil
}

//FetchAll
func FetchAll(collectionName string) ([]bson.M, error) {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	filter := bson.M{}
	docs := []bson.M{}

	opts := options.Find().SetSort(bson.D{{Key: "CreatedAt", Value: 1}})
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &docs); err != nil {
		return nil, err
	}
	return docs, nil
}

//FetchDocByCriterionMultiple returns a doc struct that tha matches the particular criteria
// i.e FetchDocByCriterionMultiple("username","abraham") returns a user struct where username is abraham and more
func FetchDocByCriterionMultiple(criteria, collectionName string, values []string) ([]bson.M, error) {
	db := database.GetMongoDB()
	client := db.GetClient()
	defer database.PutDBBack(db)
	collection := client.Database(database.DbName).Collection(collectionName)
	filter := bson.M{criteria: bson.M{"$in": values}}
	docs := []bson.M{}

	opts := options.Find().SetSort(bson.D{{Key: "CreatedAt", Value: 1}})
	cursor, err := collection.Find(context.TODO(), filter, opts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.TODO(), &docs); err != nil {
		return nil, err
	}
	return docs, nil
}
