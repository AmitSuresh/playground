package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"goworkspace/github.com/AmitSuresh/learn-go-aws/learn-go-lambda-sns-sqs/pkg/logger"

	"go.uber.org/zap"
)

// var ctx context.Context
var log *zap.Logger

func init() {
	var err error
	log, err = zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger", zap.Error(err))
	}
	//ctx = context.Background()
	//ctx = logger.Inject(ctx, log)
	defer log.Sync()
}

type Account struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}

type AccountHandler struct {
	AccountNotifier AccountNotifier
}

type AccountNotifier interface {
	NotifyAccountCreated(context.Context, Account) error
}

type SimpleAccountNotifier struct{}

func (n SimpleAccountNotifier) NotifyAccountCreated(ctx context.Context, account Account) error {
	log.Info("new account created SimpleAccountNotifier", zap.Any("username", account.Username), zap.Any("email", account.Email))
	return nil
}

type BetterAccountNotifier struct{}

func (n BetterAccountNotifier) NotifyAccountCreated(ctx context.Context, account Account) error {
	log.Info("new account created from BetterAccountNotifier", zap.Any("username", account.Username), zap.Any("email", account.Email))
	return nil
}

func (n *AccountHandler) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	var account Account

	ctx := r.Context()
	logger.Inject(ctx, log)
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		log.Fatal("Failed to decode request body", zap.Error(err))
		return
	}
	if err := n.AccountNotifier.NotifyAccountCreated(r.Context(), account); err != nil {
		log.Error("failed to notify Account Created", zap.Error(err))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(account)
}

/* func NotifyAccountCreated(ctx context.Context, account Account) error {
	time.Sleep(time.Millisecond * 500)
	log.Info("new account created", zap.Any("username", account.Username), zap.Any("email", account.Email))
	return nil
} */

func main() {
	defer log.Sync()
	mux := http.NewServeMux()

	accountHandler := &AccountHandler{
		AccountNotifier: BetterAccountNotifier{},
	}
	mux.HandleFunc("/account", accountHandler.handleCreateAccount)
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal("failed to start server", zap.Error(err))
	}
	fmt.Println()
}
