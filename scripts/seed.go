package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/lenon/portfolios/internal/database"
	"github.com/lenon/portfolios/internal/models"
	"github.com/lenon/portfolios/internal/utils"
)

// TestUser represents a test user for seeding
type TestUser struct {
	Email    string
	Password string
}

var testUsers = []TestUser{
	{
		Email:    "test1@example.com",
		Password: "Test1234",
	},
	{
		Email:    "test2@example.com",
		Password: "Test5678",
	},
	{
		Email:    "admin@example.com",
		Password: "Admin123",
	},
	{
		Email:    "demo@example.com",
		Password: "Demo1234",
	},
	{
		Email:    "user@example.com",
		Password: "User1234",
	},
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Connect to database
	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("Connected to database successfully")
	log.Println("Starting database seeding...")

	// Check if users already exist
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		log.Printf("Warning: Database already contains %d users. Do you want to continue? (y/N): ", count)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			log.Println("Seeding cancelled")
			return
		}
	}

	// Create test users
	successCount := 0
	skipCount := 0

	for _, testUser := range testUsers {
		// Check if user already exists
		var existingUser models.User
		result := db.Where("email = ?", testUser.Email).First(&existingUser)
		if result.Error == nil {
			log.Printf("User %s already exists, skipping", testUser.Email)
			skipCount++
			continue
		}

		// Hash password
		hashedPassword, err := utils.HashPassword(testUser.Password)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", testUser.Email, err)
			continue
		}

		// Create user
		user := models.User{
			Email:        testUser.Email,
			PasswordHash: hashedPassword,
		}

		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user %s: %v", testUser.Email, err)
			continue
		}

		log.Printf("Created user: %s (password: %s)", testUser.Email, testUser.Password)
		successCount++
	}

	log.Println("\n=== Seeding Summary ===")
	log.Printf("Successfully created: %d users", successCount)
	log.Printf("Skipped (already exist): %d users", skipCount)
	log.Printf("Total users in database: %d", successCount+skipCount)
	log.Println("\n=== Test User Credentials ===")
	log.Println("You can use these credentials to test the application:")
	for _, testUser := range testUsers {
		log.Printf("  Email: %s | Password: %s", testUser.Email, testUser.Password)
	}
	log.Println("\nSeeding completed successfully!")
}
