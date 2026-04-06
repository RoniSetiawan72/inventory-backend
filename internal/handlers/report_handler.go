package handlers

import (
	"net/http"

	"inventory-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

func GetReport(c *gin.Context) {
	var results []struct {
		LogID           uint   `json:"log_id"`
		TransactionType string `json:"transaction_type"`
		ReferenceID     uint   `json:"reference_id"`
		ProductName     string `json:"product_name"`
		QtyChange       int    `json:"qty_change"`
		CreatedAt       string `json:"created_at"`
	}

	query := `
		SELECT l.id as log_id, l.transaction_type, l.reference_id, p.name as product_name, l.qty_change, l.created_at
		FROM inventory_logs l
		JOIN products p ON l.product_id = p.id
		ORDER BY l.created_at DESC
	`

	if err := database.DB.Raw(query).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data laporan"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Laporan transaksi berhasil dimuat",
		"data":    results,
	})
}
