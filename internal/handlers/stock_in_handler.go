package handlers

import (
	"net/http"

	"inventory-backend/internal/models"
	"inventory-backend/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateStockIn membuat draft barang masuk (Status: CREATED)
func CreateStockIn(c *gin.Context) {
	var input struct {
		ProductID uint `json:"product_id" binding:"required"`
		Qty       int  `json:"qty" binding:"required,gt=0"` // Harus lebih dari 0
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cek apakah produk ada
	var product models.Product
	if err := database.DB.First(&product, input.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	stockIn := models.StockIn{
		ProductID: input.ProductID,
		Qty:       input.Qty,
		Status:    "CREATED",
	}

	if err := database.DB.Create(&stockIn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat Stock In"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Stock In berhasil dibuat", "data": stockIn})
}

// UpdateStockInStatus mengubah status barang masuk
func UpdateStockInStatus(c *gin.Context) {
	id := c.Param("id") // Mengambil ID dari URL
	var input struct {
		Status string `json:"status" binding:"required"` // IN_PROGRESS, DONE, CANCELLED
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Cari data Stock In
	var stockIn models.StockIn
	if err := database.DB.First(&stockIn, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Data Stock In tidak ditemukan"})
		return
	}

	// ATURAN: Tidak bisa di-cancel jika sudah DONE
	if stockIn.Status == "DONE" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak bisa mengubah status, proses sudah DONE"})
		return
	}

	// Gunakan Database Transaction agar aman
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// Update status Stock In
		stockIn.Status = input.Status
		if err := tx.Save(&stockIn).Error; err != nil {
			return err
		}

		// Jika status diubah menjadi DONE, jalankan aturan tambah stok & log
		if input.Status == "DONE" {
			// 1. Tambah Physical & Available Stock
			if err := tx.Model(&models.Product{}).Where("id = ?", stockIn.ProductID).
				Updates(map[string]interface{}{
					"physical_stock":  gorm.Expr("physical_stock + ?", stockIn.Qty),
					"available_stock": gorm.Expr("available_stock + ?", stockIn.Qty),
				}).Error; err != nil {
				return err
			}

			// 2. Catat di tabel Log (History)
			logEntry := models.InventoryLog{
				TransactionType: "IN",
				ReferenceID:     stockIn.ID,
				ProductID:       stockIn.ProductID,
				QtyChange:       stockIn.Qty,
			}
			if err := tx.Create(&logEntry).Error; err != nil {
				return err
			}
		}

		return nil // Return nil jika semua berhasil (Commit)
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses transaksi: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status berhasil diupdate menjadi " + input.Status})
}
