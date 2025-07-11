package main

import (
	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/batch"
	"TransactionSystem/internal/report"
	"TransactionSystem/internal/transaction"
	"log"
	"net/http"
)

func main() {
	db, err := transaction.InitDB("root", "n61224n61224", "localhost:3306", "transaction_db")
	if err != nil {
		log.Fatal("Failed to connect to DB: ", err)
	}

	repo := transaction.NewDBRepo(db)
	svc := transaction.NewService(repo)
	handler := transaction.NewHandler(svc)
	batchHandler := batch.NewHandler(svc)

	reportSvc := report.NewService(svc)
	reportHandler := report.NewHandler(reportSvc)

	http.Handle("/transactions", auth.AuthMiddleware(http.HandlerFunc(handler.ListUserTransactions)))
	http.Handle("/report", auth.AuthMiddleware(http.HandlerFunc(reportHandler.UserReport)))
	http.HandleFunc("/login", auth.LoginHandler)
	http.Handle("/batch", auth.AuthMiddleware(http.HandlerFunc(batchHandler.ProcessBatch)))
	http.Handle("/transactions", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handler.ListUserTransactions(w, r)
		} else if r.Method == http.MethodPost {
			handler.AddTransactionHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

	log.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
