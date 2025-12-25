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

	"github.com/kayotklimenko/gozon-go/orders-service/internal/application/commands"
	"github.com/kayotklimenko/gozon-go/orders-service/internal/application/queries"
	"github.com/kayotklimenko/gozon-go/orders-service/internal/infrastructure/messaging/rabbitmq"
	pgRepo "github.com/kayotklimenko/gozon-go/orders-service/internal/infrastructure/persistence/postgres"
	httpInterface "github.com/kayotklimenko/gozon-go/orders-service/internal/interfaces/http"
	"github.com/kayotklimenko/gozon-go/orders-service/pkg/config"
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

	orderRepo := pgRepo.NewOrderRepository(db)
	outboxRepo := pgRepo.NewOutboxRepository(db)
	unitOfWork := pgRepo.NewUnitOfWork(db)

	createOrderHandler := commands.NewCreateOrderHandler(orderRepo, outboxRepo, unitOfWork)
	updateStatusHandler := commands.NewUpdateOrderStatusHandler(orderRepo)

	getOrderHandler := queries.NewGetOrderHandler(orderRepo)
	getOrdersHandler := queries.NewGetOrdersHandler(orderRepo)

	handlers := httpInterface.NewHandlers(createOrderHandler, getOrderHandler, getOrdersHandler)
	router := httpInterface.NewRouter(handlers)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	outboxProcessor, err := rabbitmq.NewOutboxProcessor(outboxRepo, rabbitConn)
	if err != nil {
		log.Fatalf("Failed to create outbox processor: %v", err)
	}
	defer outboxProcessor.Close()
	go outboxProcessor.Start(ctx)

	paymentResultConsumer, err := rabbitmq.NewPaymentResultConsumer(rabbitConn, updateStatusHandler)
	if err != nil {
		log.Fatalf("Failed to create payment result consumer: %v", err)
	}
	defer paymentResultConsumer.Close()
	go paymentResultConsumer.Start(ctx)

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		log.Printf("Orders Service started on port %s", cfg.Server.Port)
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
		&pgRepo.OrderModel{},
		&pgRepo.OutboxModel{},
	)
	log.Println("Database migrated")
}
