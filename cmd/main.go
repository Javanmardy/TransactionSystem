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
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading .env:", err)
	}
}

/********** SIM CONFIG **********/
const (
	simBaseURL = "http://localhost:8080"

	// تعداد کاربرهای شبیه‌سازی
	simUsers = 50

	// نرخ پویا: بین 1 تا 10 تراکنش در ثانیه (کل سیستم)
	tpsMin = 1
	tpsMax = 10

	// هم‌زمانی کارگرها
	simWorkers = 12

	// seeding
	seedBalances = true
	seedMin      = 400
	seedMax      = 1500

	// درصد Fail عمدی
	failRate = 0.10

	// ادمین
	adminUser = "admin"
	adminPass = "admin"
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

	if err := resetDB(db); err != nil {
		log.Fatalf("reset db failed: %v", err)
	}

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
		_ = json.NewEncoder(w).Encode(tx)
	})))

	go runSimulation(db)

	log.Println("Server running at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

/********** DB RESET & ADMIN **********/
func resetDB(db *sql.DB) error {
	log.Println("[reset] truncating tables...")
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS=0")
	if _, err := db.Exec("TRUNCATE TABLE transactions"); err != nil {
		return err
	}
	if _, err := db.Exec("TRUNCATE TABLE users"); err != nil {
		return err
	}
	_, _ = db.Exec("SET FOREIGN_KEY_CHECKS=1")
	log.Println("[reset] done.")
	return nil
}

func promoteAdmin(db *sql.DB, username string) {
	_, _ = db.Exec(`UPDATE users SET role='admin' WHERE username=?`, username)
}

/********** SIMULATION **********/
func runSimulation(db *sql.DB) {
	time.Sleep(2 * time.Second)
	rand.Seed(time.Now().UnixNano())

	if _, err := apiRegister(adminUser, "admin@example.com", adminPass); err != nil {
		log.Printf("[admin][register] warn: %v", err)
	}
	promoteAdmin(db, adminUser)
	adminToken, err := apiLogin(adminUser, adminPass)
	if err != nil {
		log.Printf("[admin][login][ERR]: %v", err)
		return
	}
	log.Printf("[admin][login][OK]")

	users := make([]simUser, 0, simUsers)
	used := map[string]struct{}{}

	for len(users) < simUsers {
		username, email := randomUserIdentFA()
		if _, dupe := used[username]; dupe {
			continue
		}
		used[username] = struct{}{}
		password := "P" + strconv.Itoa(100000+rand.Intn(900000))

		id, regErr := apiRegister(username, email, password)
		if regErr != nil {
			log.Printf("[register][ERR] user=%s err=%v", username, regErr)
			continue
		}
		log.Printf("[register][OK] user=%s id=%d", username, id)

		tok, loginErr := apiLogin(username, password)
		if loginErr != nil {
			log.Printf("[login][ERR] user=%s err=%v", username, loginErr)
			continue
		}
		log.Printf("[login][OK] user=%s", username)

		if id == 0 {
			if fetchedID, err := apiFindUserID(tok, username); err == nil {
				id = fetchedID
			}
		}
		if id == 0 {
			log.Printf("[user-id][ERR] user=%s no id", username)
			continue
		}
		users = append(users, simUser{ID: id, Username: username, Token: tok})
	}

	if len(users) < 2 {
		log.Println("not enough users to simulate transfers")
		return
	}

	if seedBalances {
		if err := apiSeedBalancesWithToken(users, adminToken); err != nil {
			log.Printf("[seed][WARN] %v", err)
		} else {
			log.Printf("[seed][OK] %d users", len(users))
		}
	}

	log.Printf("[sim] traffic: users=%d workers=%d tps=%d..%d", len(users), simWorkers, tpsMin, tpsMax)

	type job struct {
		from simUser
		toID int
		amt  float64
	}
	jobs := make(chan job, 4096)

	var wg sync.WaitGroup
	for w := 0; w < simWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if err := apiTransfer(j.from.Token, j.toID, j.amt); err != nil {
					log.Printf("[tx][FAIL] from=%d to=%d amt=%.0f err=%v", j.from.ID, j.toID, j.amt, err)
				} else {
					log.Printf("[tx][OK]   from=%d to=%d amt=%.0f", j.from.ID, j.toID, j.amt)
				}
			}
		}()
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		rand.Seed(time.Now().UnixNano())

		for {
			targetTPS := tpsMin + rand.Intn(tpsMax-tpsMin+1)
			if targetTPS < 1 {
				targetTPS = 1
			}

			for i := 0; i < targetTPS; i++ {
				from := users[rand.Intn(len(users))]
				to := users[rand.Intn(len(users))]
				if from.ID == to.ID {
					continue
				}

				var amount float64
				if rand.Float64() < failRate {
					amount = float64(50_000 + rand.Intn(150_000))
				} else {
					amount = float64(10 + rand.Intn(90))
				}

				delay := time.Duration(rand.Intn(1000)) * time.Millisecond
				go func(f simUser, toID int, amt float64, d time.Duration) {
					time.Sleep(d)
					jobs <- job{from: f, toID: toID, amt: amt}
				}(from, to.ID, amount, delay)
			}

			<-ticker.C
		}
	}()

	select {}
}

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

func apiSeedBalancesWithToken(users []simUser, adminToken string) error {
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
	req.Header.Set("Authorization", "Bearer "+adminToken)
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

var faFirst = []string{
	"amir", "reza", "mohammad", "sara", "maryam", "ali", "zahra", "fatemeh", "hossein", "mahdi",
	"hamed", "niloufar", "saeed", "elham", "parisa", "shaghayegh", "mina", "farzad", "kian", "yasaman",
}
var faLast = []string{
	"mohammadi", "ahmadi", "hosseini", "jafari", "ghasemi", "moradi", "karimi", "heidari", "abbasi", "soleimani",
	"sadeghi", "rahimi", "ghorbani", "hashemi", "amini", "norouzi", "majidi", "kazemi", "mousavi", "ghanbari",
}
var emailDomains = []string{
	"gmail.com", "yahoo.com", "outlook.com", "proton.me", "icloud.com",
}

func randomUserIdentFA() (username, email string) {
	fn := faFirst[rand.Intn(len(faFirst))]
	ln := faLast[rand.Intn(len(faLast))]
	dom := emailDomains[rand.Intn(len(emailDomains))]
	n := 10 + rand.Intn(90) // 10..99

	switch rand.Intn(5) {
	case 0:
		username = fn + "." + ln
	case 1:
		username = fn + "_" + ln
	case 2:
		username = fn + ln + strconv.Itoa(n)
	case 3:
		username = string(fn[0]) + ln + strconv.Itoa(n)
	default:
		username = ln + "." + fn
	}
	email = username + "@" + dom
	return
}
