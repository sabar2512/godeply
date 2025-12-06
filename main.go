package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Struktur model bioskop
type Bioskop struct {
	ID     int     `db:"id" json:"id"`
	Nama   string  `db:"nama" json:"nama"`
	Lokasi string  `db:"lokasi" json:"lokasi"`
	Rating float32 `db:"rating" json:"rating"`
}

// Global DB
var db *sqlx.DB

func main() {
	var err error
	// Koneksi ke PostgreSQL
	dsn := "host=localhost port=5432 user=sabar password=postgres dbname=bioskopdb sslmode=disable"
	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatal("Gagal koneksi database:", err)
	}
	defer db.Close()

	// Test koneksi
	if err = db.Ping(); err != nil {
		log.Fatal("Database tidak dapat diakses:", err)
	}
	log.Println("✓ Koneksi database berhasil")

	// Init Gin
	r := gin.Default()

	// Endpoints CRUD
	r.POST("/bioskop", createBioskop)       // Create
	r.GET("/bioskop", getAllBioskop)        // Read All
	r.GET("/bioskop/:id", getBioskopByID)   // Read By ID
	r.PUT("/bioskop/:id", updateBioskop)    // Update
	r.DELETE("/bioskop/:id", deleteBioskop) // Delete

	// Jalankan server
	log.Println("Server berjalan di http://localhost:8080")
	r.Run(":8080")
}

// Handler POST /bioskop - Create
func createBioskop(c *gin.Context) {
	var input Bioskop

	// Ambil JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Format JSON tidak valid",
			"detail": err.Error(),
		})
		return
	}

	// Validasi
	if input.Nama == "" || input.Lokasi == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama dan Lokasi tidak boleh kosong",
		})
		return
	}

	if input.Rating < 0 || input.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rating harus antara 0 dan 5",
		})
		return
	}

	// Query insert
	query := `INSERT INTO bioskop (nama, lokasi, rating) VALUES ($1, $2, $3) RETURNING id`
	var newID int
	err := db.QueryRow(query, input.Nama, input.Lokasi, input.Rating).Scan(&newID)
	if err != nil {
		log.Println("Error insert:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menyimpan data ke database",
		})
		return
	}

	input.ID = newID
	c.JSON(http.StatusCreated, gin.H{
		"message": "Bioskop berhasil ditambahkan",
		"data":    input,
	})
}

// Handler GET /bioskop - Read All
func getAllBioskop(c *gin.Context) {
	var bioskopList []Bioskop

	// Query untuk mendapatkan semua data
	query := `SELECT id, nama, lokasi, rating FROM bioskop ORDER BY id`
	err := db.Select(&bioskopList, query)
	if err != nil {
		log.Println("Error get all:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data dari database",
		})
		return
	}

	// Jika tidak ada data
	if len(bioskopList) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "Belum ada data bioskop",
			"data":    []Bioskop{},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil diambil",
		"total":   len(bioskopList),
		"data":    bioskopList,
	})
}

// Handler GET /bioskop/:id - Read By ID
func getBioskopByID(c *gin.Context) {
	// Ambil ID dari parameter URL
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID harus berupa angka",
		})
		return
	}

	var bioskop Bioskop

	// Query untuk mendapatkan data berdasarkan ID
	query := `SELECT id, nama, lokasi, rating FROM bioskop WHERE id = $1`
	err = db.Get(&bioskop, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Bioskop dengan ID tersebut tidak ditemukan",
			})
			return
		}
		log.Println("Error get by ID:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data dari database",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Data berhasil diambil",
		"data":    bioskop,
	})
}

// Handler PUT /bioskop/:id - Update
func updateBioskop(c *gin.Context) {
	// Ambil ID dari parameter URL
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID harus berupa angka",
		})
		return
	}

	var input Bioskop

	// Ambil JSON
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "Format JSON tidak valid",
			"detail": err.Error(),
		})
		return
	}

	// Validasi input
	if input.Nama == "" || input.Lokasi == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama dan Lokasi tidak boleh kosong",
		})
		return
	}

	if input.Rating < 0 || input.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rating harus antara 0 dan 5",
		})
		return
	}

	// Cek apakah data dengan ID tersebut ada
	var exists bool
	checkQuery := `SELECT EXISTS(SELECT 1 FROM bioskop WHERE id = $1)`
	err = db.Get(&exists, checkQuery, id)
	if err != nil {
		log.Println("Error checking existence:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memeriksa data",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Bioskop dengan ID tersebut tidak ditemukan",
		})
		return
	}

	// Query update
	query := `UPDATE bioskop SET nama = $1, lokasi = $2, rating = $3 WHERE id = $4`
	result, err := db.Exec(query, input.Nama, input.Lokasi, input.Rating, id)
	if err != nil {
		log.Println("Error update:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memperbarui data",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tidak ada data yang diperbarui",
		})
		return
	}

	input.ID = id
	c.JSON(http.StatusOK, gin.H{
		"message": "Data bioskop berhasil diperbarui",
		"data":    input,
	})
}

// Handler DELETE /bioskop/:id - Delete
func deleteBioskop(c *gin.Context) {
	// Ambil ID dari parameter URL
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID harus berupa angka",
		})
		return
	}

	// Cek apakah data dengan ID tersebut ada
	var bioskop Bioskop
	checkQuery := `SELECT id, nama, lokasi, rating FROM bioskop WHERE id = $1`
	err = db.Get(&bioskop, checkQuery, id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Bioskop dengan ID tersebut tidak ditemukan",
			})
			return
		}
		log.Println("Error checking existence:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memeriksa data",
		})
		return
	}

	// Query delete
	query := `DELETE FROM bioskop WHERE id = $1`
	result, err := db.Exec(query, id)
	if err != nil {
		log.Println("Error delete:", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghapus data",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Tidak ada data yang dihapus",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bioskop berhasil dihapus",
		"data":    bioskop,
	})
}
