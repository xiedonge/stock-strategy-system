package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/xiedonge/stock-strategy-system/backend/internal/services"
)

// NewRouter wires the HTTP routes to services.
func NewRouter(stockService *services.StockService, strategyService *services.StrategyService, analysisService *services.AnalysisService) *gin.Engine {
	router := gin.Default()
	router.Use(cors.Default())

	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		api.POST("/demo/seed", func(c *gin.Context) {
			if err := stockService.SeedDemoData(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "demo data seeded"})
		})

		api.GET("/stocks", func(c *gin.Context) {
			stocks, err := stockService.ListStocks()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, stocks)
		})

		api.GET("/stocks/:code/klines", func(c *gin.Context) {
			code := c.Param("code")
			interval := c.Query("interval")
			limit, _ := strconv.Atoi(c.Query("limit"))
			klines, err := stockService.GetKLines(code, interval, limit)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, klines)
		})

		api.GET("/strategies", func(c *gin.Context) {
			strategies, err := strategyService.List()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, strategies)
		})

		api.POST("/strategies", func(c *gin.Context) {
			var req StrategyRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			strategy := req.toModel()
			if err := strategyService.Create(strategy); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, strategy)
		})

		api.GET("/strategies/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			strategy, err := strategyService.Get(uint(id))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, strategy)
		})

		api.PUT("/strategies/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			strategy, err := strategyService.Get(uint(id))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}

			var req StrategyRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			req.apply(strategy)

			if err := strategyService.Update(strategy); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, strategy)
		})

		api.DELETE("/strategies/:id", func(c *gin.Context) {
			id, _ := strconv.Atoi(c.Param("id"))
			if err := strategyService.Delete(uint(id)); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "deleted"})
		})

		api.POST("/screen", func(c *gin.Context) {
			var req ScreenRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			results, err := analysisService.Screen(req.StrategyID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, results)
		})

		api.POST("/backtest", func(c *gin.Context) {
			var req BacktestRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			result, err := analysisService.RunBacktest(req.StrategyID, req.StockCode, req.InitialCapital)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, result)
		})
	}

	return router
}
