package handlers

import (
	"net/http"

	"inventory-backend/internal/models"
	"inventory-backend/pkg/database"

	"github.com/gin-gonic/gin"
)

func CreateProduct(c *gin.Context) {
	var input struct {
		SKU  string `json:"sku" binding:"required"`
		Name string `json:"name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		SKU:  input.SKU,
		Name: input.Name,
	}

	// Simpan ke database
	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan produk. Pastikan SKU unik."})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Produk berhasil ditambahkan",
		"data":    product,
	})
}

// GetProducts digunakan untuk mengambil daftar barang (dengan fitur pencarian)
func GetProducts(c *gin.Context) {
	var products []models.Product
	search := c.Query("search")
	query := database.DB.Model(&models.Product{})

	if search != "" {
		query = query.Where("name ILIKE ? OR sku ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data produk"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
	})
}

// AdjustStock digunakan untuk mengubah stok barang secara manual (Stock Adjustment)
func AdjustStock(c *gin.Context) {
	id := c.Param("id")
	var input struct {
		PhysicalStock  int `json:"physical_stock" binding:"required,min=0"`
		AvailableStock int `json:"available_stock" binding:"required,min=0"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah atau stok tidak boleh minus"})
		return
	}

	var product models.Product
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan"})
		return
	}

	product.PhysicalStock = input.PhysicalStock
	product.AvailableStock = input.AvailableStock

	if err := database.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyesuaikan stok"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stock adjustment berhasil",
		"data":    product,
	})
}
