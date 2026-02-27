package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"github.com/mubarok-ridho/casheer-report-service/internal/handlers"
	"github.com/mubarok-ridho/casheer-report-service/internal/middleware"
	"github.com/mubarok-ridho/casheer-report-service/internal/models"
	"github.com/mubarok-ridho/casheer-report-service/internal/repository"
	"github.com/mubarok-ridho/casheer-report-service/pkg/database"
	"github.com/mubarok-ridho/casheer-report-service/pkg/messaging"
)

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ Warning: .env file not found, using environment variables")
	}

	// Initialize Database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}

	// Auto Migrate
	log.Println("📦 Running database migrations...")
	if err := db.AutoMigrate(
		&models.Expense{},
		&models.Revenue{},
		&models.ReceiptTemplate{},
		&models.DailyReport{},
	); err != nil {
		log.Fatal("❌ Failed to migrate database:", err)
	}
	log.Println("✅ Database migration completed")

	// Initialize RabbitMQ consumer
	rmq, err := messaging.NewRabbitMQ()
	if err != nil {
		log.Println("⚠️ Warning: Failed to connect to RabbitMQ:", err)
	} else {
		defer rmq.Close()

		// Start consuming order completed events
		reportHandler := handlers.NewReportHandler(db)
		go rmq.ConsumeOrderCompleted(reportHandler.HandleOrderCompleted)
	}

	// Initialize repositories
	expenseRepo := repository.NewExpenseRepository(db)
	reportRepo := repository.NewReportRepository(db)
	templateRepo := repository.NewTemplateRepository(db)

	// Initialize handlers
	expenseHandler := handlers.NewExpenseHandler(expenseRepo)
	reportHandler := handlers.NewReportHandler(db)
	templateHandler := handlers.NewTemplateHandler(templateRepo, db)

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: os.Getenv("APP_NAME"),
	})

	app.Use(cors.New())

	// Setup routes
	setupRoutes(app, expenseHandler, reportHandler, templateHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3003"
	}

	log.Printf("🚀 Report Service starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}

func setupRoutes(
	app *fiber.App,
	expenseHandler *handlers.ExpenseHandler,
	reportHandler *handlers.ReportHandler,
	templateHandler *handlers.TemplateHandler,
) {
	// Protected routes (require JWT)
	api := app.Group("/api/v1", middleware.AuthMiddleware())

	// Report routes
	reportRoutes := api.Group("/reports")
	reportRoutes.Get("/daily", reportHandler.GetDailyReport)
	reportRoutes.Get("/monthly", reportHandler.GetMonthlyReport)
	reportRoutes.Get("/yearly", reportHandler.GetYearlyReport)
	reportRoutes.Get("/revenue", reportHandler.GetRevenueSummary)
	reportRoutes.Get("/expenses", reportHandler.GetExpenseSummary)

	// Expense routes
	expenseRoutes := api.Group("/expenses")
	expenseRoutes.Post("/", expenseHandler.Create)
	expenseRoutes.Get("/", expenseHandler.GetAll)
	expenseRoutes.Get("/:id", expenseHandler.GetByID)
	expenseRoutes.Put("/:id", expenseHandler.Update)
	expenseRoutes.Delete("/:id", expenseHandler.Delete)
	expenseRoutes.Get("/categories/summary", expenseHandler.GetByCategory)

	// Template routes
	templateRoutes := api.Group("/templates")
	templateRoutes.Post("/", templateHandler.Create)
	templateRoutes.Get("/", templateHandler.GetAll)
	templateRoutes.Get("/:id", templateHandler.GetByID)
	templateRoutes.Put("/:id", templateHandler.Update)
	templateRoutes.Delete("/:id", templateHandler.Delete)
	templateRoutes.Patch("/:id/default", templateHandler.SetDefault)

	// Print routes
	printRoutes := api.Group("/print")
	printRoutes.Post("/receipt/:orderId", templateHandler.PrintReceipt)
	printRoutes.Post("/test", templateHandler.PrintTest)
}
