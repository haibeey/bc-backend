package main

import (
	"bcchallenge/controllers"
	"bcchallenge/graph"
	"bcchallenge/graph/generated"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/go-chi/chi"
	"github.com/joho/godotenv"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	router := chi.NewRouter()

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)
	router.HandleFunc("/signup", controllers.SignUp)
	router.HandleFunc("/login", controllers.Login)
	router.HandleFunc("/add-card", controllers.AddCard)
	router.HandleFunc("/user-cards", controllers.GetUserCards)
	router.HandleFunc("/add-transaction", controllers.AddTransaction)
	router.HandleFunc("/card-transactions", controllers.GetCardTransactions)
	router.HandleFunc("/bill-card", controllers.BillCard)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
