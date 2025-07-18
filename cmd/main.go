package main

import (
	"database/sql"
	"log"
	"net/http"

	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/batch"
	"TransactionSystem/internal/report"
	"TransactionSystem/internal/transaction"
	"TransactionSystem/internal/user"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	db, err := sql.Open("mysql", "root:n61224n61224@tcp(localhost:3306)/transaction_db")
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}
	defer db.Close()
	fs := http.FileServer(http.Dir("./ui"))
	http.Handle("/", fs)
	userService := user.NewMySQLService(db)
	txRepo := transaction.NewDBRepo(db)
	txService := transaction.NewService(txRepo)
	txHandler := transaction.NewHandler(txService)

	batchHandler := batch.NewHandler(txService)

	reportSvc := report.NewService(txService)
	reportHandler := report.NewHandler(reportSvc)
	authHandler := auth.NewHandler(userService)

	http.HandleFunc("/login", authHandler.LoginHandler)
	http.HandleFunc("/register", authHandler.RegisterHandler)

	http.Handle("/transactions", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			txHandler.ListUserTransactions(w, r)
		} else if r.Method == http.MethodPost {
			txHandler.AddTransactionHandler(w, r)
		} else {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})))

	http.Handle("/batch", auth.AuthMiddleware(http.HandlerFunc(batchHandler.ProcessBatch)))
	http.Handle("/report/all", auth.AuthMiddleware(http.HandlerFunc(reportHandler.AllReports)))
	http.Handle("/report", auth.AuthMiddleware(http.HandlerFunc(reportHandler.UserReport)))

	log.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
