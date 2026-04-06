package main

import (
	"fmt"
	"log"
	"os"

	"inventory-backend/internal/handlers"
	"inventory-backend/internal/models"
	"inventory-backend/pkg/database"

	"github.com/gin-contrib/cors" // Import CORS
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: File .env tidak ditemukan, pastikan sudah dibuat.")
	}

	database.ConnectDB()

	fmt.Println("Memulai migrasi database...")
	errMigrate := database.DB.AutoMigrate(
		&models.Product{},
		&models.StockIn{},
		&models.StockOut{},
		&models.InventoryLog{},
	)
	if errMigrate != nil {
		log.Fatal("Gagal melakukan migrasi database: ", errMigrate)
	}
	fmt.Println("✅ Migrasi database berhasil! Tabel sudah dibuat.")

	r := gin.Default()

	r.Use(cors.Default())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong! Server Inventory Backend berjalan lancar.",
		})
	})

	r.POST("/products", handlers.CreateProduct)
	r.GET("/products", handlers.GetProducts)
	r.PUT("/products/:id/adjust", handlers.AdjustStock)

	r.POST("/stock-in", handlers.CreateStockIn)
	r.PUT("/stock-in/:id/status", handlers.UpdateStockInStatus)

	r.POST("/stock-out", handlers.CreateStockOut)
	r.PUT("/stock-out/:id/status", handlers.UpdateStockOutStatus)

	r.GET("/reports", handlers.GetReport)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Server berjalan di port :%s\n", port)
	r.Run(":" + port)
}
