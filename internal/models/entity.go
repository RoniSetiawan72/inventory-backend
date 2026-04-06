package models

import (
	"time"
)

// Product merepresentasikan tabel inventory
type Product struct {
	ID             uint   `gorm:"primaryKey"`
	SKU            string `gorm:"uniqueIndex;not null"`
	Name           string `gorm:"not null"`
	PhysicalStock  int    `gorm:"default:0"` // Total fisik di gudang
	AvailableStock int    `gorm:"default:0"` // Fisik dikurangi yang sedang dialokasikan
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// StockIn merepresentasikan tabel barang masuk
type StockIn struct {
	ID        uint    `gorm:"primaryKey"`
	ProductID uint    `gorm:"not null"`
	Product   Product `gorm:"foreignKey:ProductID"`
	Qty       int     `gorm:"not null"`
	Status    string  `gorm:"type:varchar(20);not null;default:'CREATED'"` // CREATED, IN_PROGRESS, DONE, CANCELLED
	CreatedAt time.Time
	UpdatedAt time.Time
}

// StockOut merepresentasikan tabel barang keluar
type StockOut struct {
	ID           uint    `gorm:"primaryKey"`
	ProductID    uint    `gorm:"not null"`
	Product      Product `gorm:"foreignKey:ProductID"`
	Qty          int     `gorm:"not null"`
	Status       string  `gorm:"type:varchar(20);not null;default:'DRAFT'"` // DRAFT, IN_PROGRESS, DONE, CANCELLED
	CustomerName string  `gorm:"type:varchar(100)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type InventoryLog struct {
	ID              uint   `gorm:"primaryKey"`
	TransactionType string `gorm:"type:varchar(10);not null"` // IN / OUT
	ReferenceID     uint   `gorm:"not null"`                  // ID dari StockIn atau StockOut
	ProductID       uint   `gorm:"not null"`
	QtyChange       int    `gorm:"not null"` // Jumlah yang berubah (absolut)
	CreatedAt       time.Time
}
