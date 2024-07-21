package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"log"

	protos "github.com/AmitSuresh/playground/playservices/v14/currency/protos/currency"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/data"
	"github.com/AmitSuresh/playground/playservices/v14/product-api/handlers"
	"github.com/go-openapi/runtime/middleware"
	gohandlers "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

const (
	shutdownTime   = 6 * time.Second
	httpServerAddr = "0.0.0.0:9090"
)

var (
	serverAddr  = flag.String("addr", "0.0.0.0:9092", "The server address in the format of host:port")
	mongoClient *mongo.Client

	mdb_username string
	mdb_password string
	mdb_cluster  string
	mdb_appname  string
)

func setupHTTPServer(l *zap.Logger, v *data.Validation, cc protos.CurrencyClient, db *data.ProductsDB) *http.Server {

	ph := handlers.NewProducts(l, v, cc, db)

	sm := mux.NewRouter()

	// Handlers for API endpoints
	getR := sm.Methods(http.MethodGet).Subrouter()
	getR.HandleFunc("/products", ph.ListAll).Queries("currency", "{[A-Z{3}]}")
	getR.HandleFunc("/products", ph.ListAll)
	getR.HandleFunc("/products", ph.ListSingleProduct)
	getR.HandleFunc("/products", ph.ListSingleProduct).Queries("id", "{id:[0-9a-fA-F]{24}}", "currency", "{currency:[A-Z]{3}}")
	getR.HandleFunc("/migrate", ph.MigrateDocs).Queries("currency", "{currency:[A-Z]{3}}")

	putR := sm.Methods(http.MethodPut).Subrouter()
	putR.HandleFunc("/products", ph.Update).Queries("id", "{id:[0-9a-fA-F]{24}}", "currency", "{currency:[A-Z]{3}}")
	putR.Use(ph.MiddlewareValidateProduct)

	postR := sm.Methods(http.MethodPost).Subrouter()
	postR.HandleFunc("/products", ph.Create)
	postR.Use(ph.MiddlewareValidateProduct)

	deleteR := sm.Methods(http.MethodDelete).Subrouter()
	deleteR.HandleFunc("/products", ph.Delete).Queries("id", "{id:[0-9a-fA-F]{24}}")

	// Handler for documentation
	opts := middleware.RedocOpts{SpecURL: "/swagger.yaml"}
	sh := middleware.Redoc(opts, nil)
	getR.Handle("/docs", sh)
	getR.Handle("/swagger.yaml", http.FileServer(http.Dir("./")))

	return &http.Server{
		Addr:         httpServerAddr,
		Handler:      gohandlers.CORS(gohandlers.AllowedOrigins([]string{"*"}))(sm),
		IdleTimeout:  120 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 4 * time.Second,
		ErrorLog:     zap.NewStdLog(l),
	}
}

func main() {
	flag.Parse()
	l, _ := zap.NewProduction()
	defer l.Sync()

	err := godotenv.Load()
	if err != nil {
		l.Error("error loading .env file")
	}

	mdb_username = url.QueryEscape(os.Getenv("MDB_USERNAME"))
	mdb_password = url.QueryEscape(os.Getenv("MDB_PASSWORD"))
	mdb_cluster = os.Getenv("MDB_CLUSTER")
	mdb_appname = os.Getenv("MDB_APPNAME")

	if mdb_username == "" || mdb_password == "" || mdb_cluster == "" || mdb_appname == "" {
		log.Fatalf("MongoDB credentials or cluster not properly set")
	}

	mdb_URI := fmt.Sprintf("mongodb+srv://%s:%s@%s/?retryWrites=true&w=majority&appName=%s",
		mdb_username, mdb_password, mdb_cluster, mdb_appname)

	l.Info("[INFO]", zap.Any("serverAddr", *serverAddr))
	grpcConn := data.GetgrpcClient(*serverAddr, l)
	defer grpcConn.Close()

	cc := protos.NewCurrencyClient(grpcConn)

	v := data.NewValidation()
	db := data.GetProductsDB(cc, l, mongoClient)
	mongoClient, err = db.GetMongoClient(mdb_URI)
	if err != nil {
		l.Error("error getting mongo client", zap.Error(err))
	}
	defer db.DisconnectMongoClient()

	err = db.GetMongoCollection("Cluster0", "ecommerce")
	if err != nil {
		l.Error("error retrieving mongo collection", zap.Error(err))
	}

	// Setup HTTP server
	httpServer := setupHTTPServer(l, v, cc, db)

	// Start HTTP server
	go func() {
		l.Info("Starting HTTP server", zap.String("address", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal("error starting HTTP server", zap.Error(err))
		}
	}()

	// Graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	<-signalChan

	l.Info("Shutdown signal received, shutting down servers...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		l.Error("error during HTTP server shutdown", zap.Error(err))
	}

	l.Info("HTTP server shutdown complete.")
}
