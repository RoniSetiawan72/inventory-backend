package handlers

import (
	"net/http"

	"inventory-backend/internal/models"
	"inventory-backend/pkg/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CreateStockOut(c *gin.Context) {
	var input struct {
		ProductID    uint   `json:"product_id" binding:"required"`
		Qty          int    `json:"qty" binding:"required,gt=0"`
		CustomerName string `json:"customer_name"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var product models.Product
		if err := tx.First(&product, input.ProductID).Error; err != nil {
			return err
		}

		if product.AvailableStock < input.Qty {
			return gorm.ErrInvalidData
		}

		if err := tx.Model(&product).Update("available_stock", gorm.Expr("available_stock - ?", input.Qty)).Error; err != nil {
			return err
		}

		stockOut := models.StockOut{
			ProductID:    input.ProductID,
			Qty:          input.Qty,
			Status:       "DRAFT",
			CustomerName: input.CustomerName,
		}

		if err := tx.Create(&stockOut).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		if err == gorm.ErrInvalidData {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Stok tersedia (Available Stock) tidak mencukupi untuk alokasi ini."})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat alokasi Stock Out"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Pesanan berhasil dialokasikan (DRAFT)"})
}

func UpdateStockOutStatus(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		var stockOut models.StockOut
		if err := tx.First(&stockOut, id).Error; err != nil {
			return err
		}

		if stockOut.Status == "DONE" || stockOut.Status == "CANCELLED" {
			return gorm.ErrInvalidTransaction
		}

		if input.Status == "CANCELLED" {
			if err := tx.Model(&models.Product{}).Where("id = ?", stockOut.ProductID).
				Update("available_stock", gorm.Expr("available_stock + ?", stockOut.Qty)).Error; err != nil {
				return err
			}
		}

		if input.Status == "DONE" {
			if err := tx.Model(&models.Product{}).Where("id = ?", stockOut.ProductID).
				Update("physical_stock", gorm.Expr("physical_stock - ?", stockOut.Qty)).Error; err != nil {
				return err
			}

			logEntry := models.InventoryLog{
				TransactionType: "OUT",
				ReferenceID:     stockOut.ID,
				ProductID:       stockOut.ProductID,
				QtyChange:       stockOut.Qty,
			}
			if err := tx.Create(&logEntry).Error; err != nil {
				return err
			}
		}

		stockOut.Status = input.Status
		return tx.Save(&stockOut).Error
	})

	if err != nil {
		if err == gorm.ErrInvalidTransaction {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak bisa mengubah status karena pesanan sudah DONE atau CANCELLED"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses perubahan status: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status berhasil diupdate menjadi " + input.Status})
}
