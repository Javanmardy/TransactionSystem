package main

import (
	"TransactionSystem/internal/transaction"
	"log"
	"net/http"
)

func main() {
	repo := transaction.NewMockRepo()
	svc := transaction.NewService(repo)
	handler := transaction.NewHandler(svc)

	http.HandleFunc("/transactions", handler.ListUserTransactions)

	log.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
