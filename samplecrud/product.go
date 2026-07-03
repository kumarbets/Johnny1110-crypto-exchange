package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Product is the resource we run CRUD against (matches the DB table).
type Product struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// productInput is the write payload — no ID (the DB assigns it).
type productInput struct {
	Name     string  `json:"name" binding:"required"`
	Price    float64 `json:"price" binding:"gte=0"`
	Quantity int     `json:"quantity" binding:"gte=0"`
}

// ProductHandler holds the DB so every handler can reach it.
type ProductHandler struct {
	DB *sql.DB
}

// Create — POST /products
func (h *ProductHandler) Create(c *gin.Context) {
	var in productInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.DB.Exec(
		`INSERT INTO products (name, price, quantity) VALUES (?, ?, ?)`,
		in.Name, in.Price, in.Quantity,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, _ := res.LastInsertId()
	c.JSON(http.StatusCreated, Product{
		ID: id, Name: in.Name, Price: in.Price, Quantity: in.Quantity,
	})
}

// List — GET /products
func (h *ProductHandler) List(c *gin.Context) {
	rows, err := h.DB.Query(`SELECT id, name, price, quantity FROM products ORDER BY id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	products := make([]Product, 0)
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Quantity); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		products = append(products, p)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, products)
}

// Get — GET /products/:id
func (h *ProductHandler) Get(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var p Product
	err = h.DB.QueryRow(
		`SELECT id, name, price, quantity FROM products WHERE id = ?`, id,
	).Scan(&p.ID, &p.Name, &p.Price, &p.Quantity)

	if errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, p)
}

// Update — PUT /products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var in productInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := h.DB.Exec(
		`UPDATE products SET name = ?, price = ?, quantity = ? WHERE id = ?`,
		in.Name, in.Price, in.Quantity, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.JSON(http.StatusOK, Product{
		ID: id, Name: in.Name, Price: in.Price, Quantity: in.Quantity,
	})
}

// Delete — DELETE /products/:id
func (h *ProductHandler) Delete(c *gin.Context) {
	id, err := parseID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	res, err := h.DB.Exec(`DELETE FROM products WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if n, _ := res.RowsAffected(); n == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func parseID(c *gin.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}
