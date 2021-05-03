package test

import (
	"bcchallenge/controllers"
	"bcchallenge/database"
	"bcchallenge/utils"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"testing"

	"github.com/joho/godotenv"
)

func handleInterupt() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		cleanUpDb()
		os.Exit(1)
	}()
}

func TestApp(t *testing.T) {

	os.Setenv("HOST", "127.0.0.1:8080")
	os.Setenv("TESTING", "TESTING")
	os.Setenv("DBNAME", "bcchallenge")
	err := godotenv.Load("../.env")
	if err != nil {
		t.Fatalf("Error loading .env file")
	}

	handleInterupt()
}

func TestUtils(t *testing.T) {
	if utils.SpaceStringsBuilder("abraham akerele") != "abrahamakerele" {
		t.Fatalf("space not removed")
	}

	if utils.ContainsOnlyDigit("123312e12321") {
		t.Fatalf("Alphabet present")
	}
	m, err := controllers.GetPublicKey()
	if err != nil {
		t.Fatalf("Cant't fetch key present")
	}
	testData := `{"number":"4757140000000001","cvv":"123"}`
	res, err := utils.PGPEncrypt1(testData, m["publicKey"])
	if err != nil {
		t.Fatalf(err.Error())
	}

	fmt.Println(res)
	res, err = utils.PGPEncrypt(testData, m["publicKey"])
	if err != nil {
		t.Fatalf(err.Error())
	}

	log.Println(utils.GenerateIdempotentKey())
	log.Println("all good")
}
func cleanUpDb() {
	if os.Getenv("CLEAR") == "CLEAR" {
		client := database.GetMongoDB().GetClient()
		log.Print(client.Database(database.DbName).Drop(context.TODO()))
	}
}
