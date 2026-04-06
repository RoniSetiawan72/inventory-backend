# Sistem Manajemen Inventaris (Inventory Management System)

Aplikasi *Full-Stack* untuk manajemen inventaris yang mengimplementasikan fitur *Two-Phase Commitment* pada proses barang keluar (*Stock Out*) dan memisahkan pencatatan *Physical Stock* dan *Available Stock* untuk mencegah *overselling*.

## 🏗️ Arsitektur Sistem

Aplikasi ini dibangun menggunakan arsitektur *Client-Server* dengan pemisahan *backend* dan *frontend* secara jelas:

1. **Backend (Golang & PostgreSQL):** - Menggunakan kerangka kerja **Gin** untuk *routing* API dan **GORM** sebagai ORM.
   - Menerapkan prinsip **SOLID** dan **Clean Architecture** ringan (pemisahan antara `models`, `handlers`, dan `database`).
   - Database PostgreSQL dijalankan melalui **Docker** untuk menjamin konsistensi *environment*.
   - Integritas data dijaga secara ketat menggunakan **Database Transactions** pada GORM, terutama saat melakukan operasi *Two-Phase Commitment* (Alokasi -> Eksekusi/Rollback) untuk mencegah *Race Condition*.

2. **Frontend (React.js):**
   - Dibangun menggunakan **Vite** untuk performa *build* yang cepat.
   - State management menggunakan **Zustand** agar lebih ringan dan mengurangi *boilerplate code* dibandingkan Redux.
   - *Styling* antarmuka menggunakan **Tailwind CSS**.

---

## 🚀 Instruksi Cara Menjalankan Aplikasi

### Persyaratan Sistem (*Prerequisites*)
- [Docker](https://www.docker.com/) & Docker Compose
- [Golang](https://go.dev/) (v1.20+)
- [Node.js](https://nodejs.org/) (v18+)

### Langkah 1: Menjalankan Backend & Database
1. Buka terminal dan masuk ke folder `inventory-backend`.
2. Salin file `.env.example` menjadi `.env` (atau buat file `.env` dengan kredensial database).
3. Jalankan container PostgreSQL:
   ```bash
   docker-compose up -d
4. Unduh semua dependencies Go:
   go mod tidy
5. Jalankan server Golang (Server akan berjalan di http://localhost:8080 dan melakukan auto-migrate tabel):
   go run cmd/main.go

### Langkah 2: Menjalankan Frontend
1. Instal semua dependencies Node:
   npm install
2. Jalankan development server:
   npm run dev

### 🤖 AI Tools yang Digunakan
- **Google Gemini** (Sebagai *pair programmer* utama untuk perancangan arsitektur, *debugging*, dan *code generation*).
- **GitHub Copilot / Cursor** (Untuk *autocomplete* kode secara *real-time* di dalam VS Code).

---

### 💬 Prompt Paling Kompleks yang Digunakan
Prompt di bawah ini digunakan untuk merancang logika inti dari aplikasi ini, yaitu *Two-Phase Commitment* pada fitur *Stock Out*:

> *"Buatkan sebuah handler Golang menggunakan GORM untuk fitur Stock Out (Barang Keluar) yang menerapkan sistem Two-Phase Commitment. 
> Tahap 1 (Allocation): Saat menerima request, cek apakah `available_stock` cukup. Jika ya, kurangi `available_stock` tanpa menyentuh `physical_stock`, lalu simpan statusnya sebagai 'DRAFT'.
> Tahap 2 (Execution/Rollback): Buat endpoint untuk mengupdate status. Jika diubah menjadi 'DONE', kurangi `physical_stock` dan catat ke tabel `inventory_logs`. Jika di-cancel ('CANCELLED'), kembalikan stok yang direservasi tadi kembali ke `available_stock`. Gunakan blok DB Transaction agar aman dari race condition."*

---

### 🛠️ Modifikasi Manual demi Best Practice

Meskipun AI memberikan kode yang secara logika berjalan dengan baik, ada bagian krusial yang saya modifikasi secara manual untuk mematuhi **Best Practice terkait Integritas Database**:

**Konteks Masalah:**
Pada generasi kode awal untuk fitur *Update Stock In/Out Status*, AI memberikan implementasi pembaruan data secara sekuensial (berurutan) menggunakan metode `db.Save()` dan `db.Create()` biasa. 

**Modifikasi yang Dilakukan:**
Saya menyadari bahwa jika terjadi kegagalan (misalnya server mati mendadak) tepat di antara proses pemotongan stok di tabel `products` dan pembuatan log di tabel `inventory_logs`, data akan menjadi tidak inkonsisten (stok berkurang tapi riwayat tidak ada).

Oleh karena itu, saya **membungkus ulang logika tersebut secara manual ke dalam blok `database.DB.Transaction(func(tx *gorm.DB) error { ... })`**. 

**Contoh Kode yang Dimodifikasi:**
```go
// KODE AWAL DARI AI (Rentan Inkosistensi Data):
// db.Model(&product).Update("physical_stock", newStock)
// db.Create(&logEntry)

// MODIFIKASI MANUAL SAYA (Menerapkan ACID Properties):
err := database.DB.Transaction(func(tx *gorm.DB) error {
    if err := tx.Model(&models.Product{}).Where("id = ?", stockOut.ProductID).
        Update("physical_stock", gorm.Expr("physical_stock - ?", stockOut.Qty)).Error; err != nil {
        return err // Rollback jika gagal memotong stok
    }

    logEntry := models.InventoryLog{...}
    if err := tx.Create(&logEntry).Error; err != nil {
        return err // Rollback jika gagal mencatat log
    }

    return nil // Commit transaksi jika semua berhasil
})