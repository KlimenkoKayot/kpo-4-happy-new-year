package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/kayotklimenko/gozon-go/payments-service/internal/application/commands"
	"github.com/kayotklimenko/gozon-go/payments-service/internal/application/queries"
	"github.com/kayotklimenko/gozon-go/payments-service/internal/infrastructure/messaging/rabbitmq"
	pgRepo "github.com/kayotklimenko/gozon-go/payments-service/internal/infrastructure/persistence/postgres"
	httpInterface "github.com/kayotklimenko/gozon-go/payments-service/internal/interfaces/http"
	"github.com/kayotklimenko/gozon-go/payments-service/pkg/config"
)

func main() {
	cfg := config.Load()

	db := connectDB(cfg)
	migrateDB(db)

	rabbitConn, err := rabbitmq.Connect(cfg.GetRabbitMQURL())
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer rabbitConn.Close()

	if err := rabbitmq.SetupExchangeAndQueues(rabbitConn); err != nil {
		log.Fatalf("Failed to setup RabbitMQ: %v", err)
	}

	accountRepo := pgRepo.NewAccountRepository(db)
	transactionRepo := pgRepo.NewTransactionRepository(db)
	inboxRepo := pgRepo.NewInboxRepository(db)
	outboxRepo := pgRepo.NewOutboxRepository(db)
	unitOfWork := pgRepo.NewUnitOfWork(db)

	createAccountHandler := commands.NewCreateAccountHandler(accountRepo)
	depositHandler := commands.NewDepositHandler(accountRepo, transactionRepo, unitOfWork)
	processPaymentHandler := commands.NewProcessPaymentHandler(
		accountRepo,
		transactionRepo,
		inboxRepo,
		outboxRepo,
		unitOfWork,
	)

	getBalanceHandler := queries.NewGetBalanceHandler(accountRepo)

	handlers := httpInterface.NewHandlers(createAccountHandler, depositHandler, getBalanceHandler)
	router := httpInterface.NewRouter(handlers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	paymentRequestConsumer, err := rabbitmq.NewPaymentRequestConsumer(rabbitConn, processPaymentHandler)
	if err != nil {
		log.Fatalf("Failed to create payment request consumer: %v", err)
	}
	defer paymentRequestConsumer.Close()
	go paymentRequestConsumer.Start(ctx)

	outboxProcessor, err := rabbitmq.NewOutboxProcessor(outboxRepo, rabbitConn)
	if err != nil {
		log.Fatalf("Failed to create outbox processor: %v", err)
	}
	defer outboxProcessor.Close()
	go outboxProcessor.Start(ctx)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Payments Service started on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	time.Sleep(2 * time.Second)
	log.Println("Server stopped")
}

func connectDB(cfg *config.Config) *gorm.DB {
	var db *gorm.DB
	var err error

	for i := 0; i < 30; i++ {
		db, err = gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Warn),
		})
		if err == nil {
			log.Println("Connected to database")
			return db
		}
		log.Printf("Waiting for database... attempt %d", i+1)
		time.Sleep(2 * time.Second)
	}

	log.Fatalf("Failed to connect to database: %v", err)
	return nil
}

func migrateDB(db *gorm.DB) {
	db.AutoMigrate(
		&pgRepo.AccountModel{},
		&pgRepo.TransactionModel{},
		&pgRepo.InboxModel{},
		&pgRepo.OutboxModel{},
	)
	log.Println("Database migrated")
}
