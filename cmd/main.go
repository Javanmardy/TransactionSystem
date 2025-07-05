package main

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/report"
	"TransactionSystem/internal/transaction"
	"log"
	"net/http"
)

func main() {
	repo := transaction.NewMockRepo()
	svc := transaction.NewService(repo)
	handler := transaction.NewHandler(svc)

	reportSvc := report.NewService(svc)
	reportHandler := report.NewHandler(reportSvc)

	http.Handle("/transactions", auth.AuthMiddleware(http.HandlerFunc(handler.ListUserTransactions)))
	http.Handle("/report", auth.AuthMiddleware(http.HandlerFunc(reportHandler.UserReport)))
	http.HandleFunc("/login", auth.LoginHandler)

	log.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
