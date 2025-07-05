package main

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/transaction"
	"log"
	"net/http"
)

func main() {
	repo := transaction.NewMockRepo()
	svc := transaction.NewService(repo)
	handler := transaction.NewHandler(svc)

	http.Handle("/transactions", auth.AuthMiddleware(http.HandlerFunc(handler.ListUserTransactions)))
	http.HandleFunc("/login", auth.LoginHandler)

	log.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
