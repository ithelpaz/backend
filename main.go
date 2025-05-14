package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"log"
	"os"

	"ithelp/handlers"
	"ithelp/middleware"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	// ENV yüklə
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  .env faylı tapılmadı, davam edilir...")
	}
    app := fiber.New()

    // Allow all origins, methods and headers
    app.Use(cors.New(cors.Config{
        AllowOrigins: "*",
        AllowHeaders: "*",
        AllowMethods: "*",
    }))

	// Auth route-lar
	app.Post("/api/register", handlers.Register)
	app.Post("/api/login", handlers.Login)
	app.Get("/api/me", middleware.JWTMiddleware(), handlers.Me)

	// User routes
	userGroup := app.Group("/api/users", middleware.JWTMiddleware())
	userGroup.Get("/", handlers.ListUsers)        // Admin only
	userGroup.Get("/:id", handlers.GetUser)       // Self or admin
	userGroup.Put("/:id", handlers.UpdateUser)    // Self or admin
	userGroup.Delete("/:id", handlers.DeleteUser) // Admin only
	
	// Support routes
	supportGroup := app.Group("/api/support", middleware.JWTMiddleware())
	supportGroup.Post("/", handlers.CreateSupportRequest)
	supportGroup.Get("/my", handlers.ListMySupportRequests)
	supportGroup.Get("/", handlers.ListAllSupportRequests)
	supportGroup.Put("/:id", handlers.UpdateSupportRequest)
	supportGroup.Get("/assigned", handlers.ListAssignedSupportRequests)

	// Technician notes
	noteGroup := app.Group("/api/notes", middleware.JWTMiddleware())
	noteGroup.Post("/", handlers.AddTechNote)          // only tech
	noteGroup.Get("/:id", handlers.ListNotesByRequest) // public per request_id

	// Plan routes
	planGroup := app.Group("/api/plans", middleware.JWTMiddleware())
	planGroup.Get("/", handlers.ListPlans)
	planGroup.Post("/", handlers.CreatePlan)
	planGroup.Put("/:id", handlers.UpdatePlan)
	planGroup.Delete("/:id", handlers.DeletePlan)



	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Fatal(app.Listen(":" + port))
}