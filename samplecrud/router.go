package main

import (
	"database/sql"

	"github.com/gin-gonic/gin"
)

// NewRouter builds the HTTP engine and wires every route to a handler,
// using the DB you pass in.
func NewRouter(db *sql.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())

	h := &ProductHandler{DB: db}

	products := r.Group("/products")
	{
		products.POST("", h.Create)       // POST   /products
		products.GET("", h.List)          // GET    /products
		products.GET("/:id", h.Get)       // GET    /products/:id
		products.PUT("/:id", h.Update)    // PUT    /products/:id
		products.DELETE("/:id", h.Delete) // DELETE /products/:id
	}

	return r
}
