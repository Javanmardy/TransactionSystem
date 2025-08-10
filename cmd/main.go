package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"TransactionSystem/internal/auth"
	"TransactionSystem/internal/batch"
	"TransactionSystem/internal/report"
	"TransactionSystem/internal/transaction"
	"TransactionSystem/internal/user"

	_ "github.com/go-sql-driver/mysql"
)


const (
	simBaseURL   = "http://localhost:8080"
	simUsers     = 60   // چند کاربر بسازد
	simTPS       = 1    // تراکنش بر ثانیه هدف
	simWorkers   = 12   // تعداد goroutine ها
	seedBalances = true // شارژ اولیه؟
	seedMin      = 300  // حداقل شارژ
	seedMax      = 1200 // حداکثر شارژ
	adminUser    = "u2" // ادمین موجود
	adminPass    = "p2" // پسورد ادمین
)

type loginResp struct {
	Token string `json:"token"`
}
type registerResp struct {
	ID int `json:"id"`
}

type simUser struct {
	ID       int
	Username string
	Token    string
}

func main() {
	db, err := sql.Open(
		"mysql",
		"root:n61224n61224@tcp(localhost:3306)/transaction_db?parseTime=true&loc=Local",
	)
	if err != nil {
		log.Fatal("Failed to connect to MySQL:", err)
	}
	defer db.Close()

	// ---- Server setup ----
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
	userHandler := user.NewHandler(userService)
	http.Handle("/users", auth.AuthMiddleware(http.HandlerFunc(userHandler.ListUsers)))

	http.Handle("/batch", auth.AuthMiddleware(
		auth.RoleRequired("admin", http.HandlerFunc(batchHandler.ProcessBatch)),
	))

	http.Handle("/report/all", auth.AuthMiddleware(http.HandlerFunc(reportHandler.AllReports)))
	http.Handle("/report", auth.AuthMiddleware(http.HandlerFunc(reportHandler.UserReport)))
	http.Handle("/report/summary", auth.AuthMiddleware(http.HandlerFunc(reportHandler.AdminReport)))
	http.Handle("/tx/", auth.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/tx/")
		id, _ := strconv.Atoi(idStr)
		if id <= 0 {
			http.Error(w, "bad id", http.StatusBadRequest)
			return
		}
		tx := txService.GetTransactionByID(id)
		if tx == nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tx)
	})))

	// ---- Simulation Goroutine ----
	go runSimulation()
	log.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

/********** SIMULATION **********/
func runSimulation() {
	time.Sleep(2 * time.Second)
	rand.Seed(time.Now().UnixNano())

	users := make([]simUser, 0, simUsers)
	for i := 0; i < simUsers; i++ {
		username := fmt.Sprintf("bot%d", rand.Intn(1_000_000))
		password := "pass123"
		email := username + "@mail.com"

		// register
		regID, regErr := apiRegister(username, email, password)
		if regErr != nil {
			log.Printf("[register][ERR] user=%s err=%v", username, regErr)
			continue
		}
		log.Printf("[register][OK] user=%s id=%d", username, regID)

		// login
		token, loginErr := apiLogin(username, password)
		if loginErr != nil {
			log.Printf("[login][ERR] user=%s err=%v", username, loginErr)
			continue
		}
		log.Printf("[login][OK] user=%s", username)

		id := regID
		if id == 0 {
			if fetchedID, err := apiFindUserID(token, username); err == nil {
				id = fetchedID
			}
		}
		if id == 0 {
			log.Printf("[user-id][ERR] user=%s no id", username)
			continue
		}

		users = append(users, simUser{ID: id, Username: username, Token: token})
	}

	if len(users) < 2 {
		log.Println("not enough users to simulate transfers")
		return
	}

	if seedBalances {
		if err := apiSeedBalances(users); err != nil {
			log.Printf("[seed][WARN] %v", err)
		} else {
			log.Printf("[seed][OK] seeded %d users", len(users))
		}
	}

	perWorker := float64(simTPS) / float64(simWorkers)
	log.Printf("[sim] start traffic: users=%d workers=%d tps=%d (≈%.2f per worker)",
		len(users), simWorkers, simTPS, perWorker)

	var wg sync.WaitGroup
	for w := 0; w < simWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			interval := time.Duration(float64(time.Second) / perWorker)
			if interval < time.Millisecond {
				interval = time.Millisecond
			}
			t := time.NewTicker(interval)
			defer t.Stop()

			for range t.C {
				// pick random distinct users
				from := users[rand.Intn(len(users))]
				to := users[rand.Intn(len(users))]
				if from.ID == to.ID {
					continue
				}
				amount := float64(rand.Intn(90) + 10)

				if err := apiTransfer(from.Token, to.ID, amount); err != nil {
					log.Printf("[tx][FAIL] from=%d to=%d amt=%.0f err=%v", from.ID, to.ID, amount, err)
				} else {
					log.Printf("[tx][OK] from=%d to=%d amt=%.0f", from.ID, to.ID, amount)
				}
			}
		}()
	}
	wg.Wait()
}

/********** API helpers **********/
func apiRegister(username, email, password string) (int, error) {
	b, _ := json.Marshal(map[string]any{
		"username": username, "email": email, "password": password,
	})
	resp, err := http.Post(simBaseURL+"/register", "application/json", bytes.NewReader(b))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("status %d", resp.StatusCode)
	}
	var r registerResp
	_ = json.NewDecoder(resp.Body).Decode(&r)
	return r.ID, nil
}

func apiLogin(username, password string) (string, error) {
	b, _ := json.Marshal(map[string]string{"username": username, "password": password})
	resp, err := http.Post(simBaseURL+"/login", "application/json", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	var lr loginResp
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return "", err
	}
	return lr.Token, nil
}

func apiFindUserID(token, username string) (int, error) {
	req, _ := http.NewRequest(http.MethodGet, simBaseURL+"/users", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var list []struct {
		ID       int    `json:"id"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return 0, err
	}
	for _, u := range list {
		if u.Username == username {
			return u.ID, nil
		}
	}
	return 0, fmt.Errorf("not found")
}

func apiSeedBalances(users []simUser) error {
	// admin login
	tok, err := apiLogin(adminUser, adminPass)
	if err != nil {
		return fmt.Errorf("admin login failed: %w", err)
	}

	type batchTx struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
		Status string  `json:"status"`
	}
	var payload struct {
		Transactions []batchTx `json:"transactions"`
	}
	for _, u := range users {
		amt := rand.Intn(seedMax-seedMin+1) + seedMin
		payload.Transactions = append(payload.Transactions, batchTx{
			UserID: u.ID, Amount: float64(amt), Status: "success",
		})
	}
	b, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, simBaseURL+"/batch", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+tok)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("batch status=%d", resp.StatusCode)
	}
	return nil
}

func apiTransfer(token string, toUserID int, amount float64) error {
	body := map[string]any{"to_user_id": toUserID, "amount": amount}
	b, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, simBaseURL+"/transactions", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
