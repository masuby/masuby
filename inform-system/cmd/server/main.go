package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"inform-system/config"
	"inform-system/internal/database"
	"inform-system/internal/handlers"
	"inform-system/internal/middleware"
	"inform-system/internal/websocket"
)

func main() {
	// Load environment variables
	godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Initialize JWT secret
	middleware.SetJWTSecret(cfg.JWTSecret)

	// Connect to database
	if err := database.Connect(cfg.DatabaseURL); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize database schema
	if err := database.InitSchema(); err != nil {
		log.Fatalf("Failed to initialize schema: %v", err)
	}

	// Seed initial data
	database.SeedIndicators()
	database.SeedAdminUser()
	database.SeedTanzaniaCommittees()

	// Initialize WebSocket hub
	hub := websocket.NewHub()
	go hub.Run()
	handlers.SetWebSocketHub(hub)

	// Initialize Gin router
	router := gin.Default()

	// CORS middleware for frontend integration
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		// Allow requests from frontend dev server and production domain
		allowedOrigins := []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:3000",
			"https://inform.co.tz",
			"https://www.inform.co.tz",
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Load templates
	router.SetFuncMap(template.FuncMap{
		"safe": func(s string) template.HTML { return template.HTML(s) },
	})

	// Get the working directory for templates
	workDir, _ := os.Getwd()
	templatePath := filepath.Join(workDir, "web", "templates", "*")
	staticPath := filepath.Join(workDir, "web", "static")

	router.LoadHTMLGlob(templatePath)
	router.Static("/static", staticPath)

	// Public routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{"title": "INFORM Risk Management System"})
	})

	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", gin.H{"title": "Login"})
	})

	router.GET("/register", func(c *gin.Context) {
		c.HTML(http.StatusOK, "register.html", gin.H{"title": "Register"})
	})

	router.GET("/transparency", func(c *gin.Context) {
		c.HTML(http.StatusOK, "transparency.html", gin.H{"title": "Transparency - Formulas & Data Flow"})
	})

	// API v1 routes
	api := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := api.Group("/auth")
		{
			auth.POST("/login", handlers.Login)
			auth.POST("/register", handlers.Register)
			auth.POST("/logout", handlers.Logout)
		}

		// Transparency routes (public)
		transparency := api.Group("/transparency")
		{
			transparency.GET("/formulas", handlers.GetFormulas)
			transparency.GET("/dataflow", handlers.GetDataFlow)
			transparency.GET("/linkages", handlers.GetSheetLinkages)
			transparency.GET("/api", handlers.GetAPIDocumentation)
		}

		// Protected routes
		protected := api.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			protected.GET("/users/me", handlers.GetCurrentUser)

			// Committee routes
			protected.GET("/committees", handlers.ListCommittees)
			protected.POST("/committees", middleware.RoleMiddleware("admin"), handlers.CreateCommittee)

			// Indicator routes
			protected.GET("/indicators", handlers.ListIndicators)
			protected.GET("/indicators/:id", handlers.GetIndicator)

			// Data entry routes
			protected.POST("/data", handlers.CreateDataEntry)
			protected.GET("/data", handlers.ListDataEntries)
			protected.PUT("/data/:id/verify", middleware.RoleMiddleware("admin", "regional_committee"), handlers.VerifyDataEntry)

			// Risk calculation routes
			protected.GET("/risk/calculate", handlers.CalculateRiskScores)
			protected.GET("/risk/scores", handlers.GetRiskScores)

			// Early Warning System routes (Module 03)
			protected.POST("/hazard-forecasts", handlers.CreateHazardForecast)
			protected.GET("/hazard-forecasts", handlers.ListHazardForecasts)
			protected.POST("/warnings", handlers.CreateWarning)
			protected.GET("/warnings", handlers.ListWarnings)
			protected.GET("/warnings/active", handlers.GetActiveWarnings)
			protected.PUT("/warnings/:id/validate", middleware.RoleMiddleware("admin", "regional_committee"), handlers.ValidateWarning)
		}

		// WebSocket route
		api.GET("/ws", func(c *gin.Context) {
			// Extract user info from token if present
			var userID int64 = 0
			var userRole, adm1Code string = "viewer", ""

			if authHeader := c.GetHeader("Authorization"); authHeader != "" {
				// Parse token for user info
			}

			hub.HandleWebSocket(c.Writer, c.Request, userID, userRole, adm1Code)
		})
	}

	// Protected page routes
	pages := router.Group("/app")
	pages.Use(authPageMiddleware())
	{
		pages.GET("/dashboard", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{"title": "Dashboard"})
		})

		pages.GET("/data-entry", func(c *gin.Context) {
			c.HTML(http.StatusOK, "data_entry.html", gin.H{"title": "Data Entry"})
		})

		pages.GET("/risk-scores", func(c *gin.Context) {
			c.HTML(http.StatusOK, "risk_scores.html", gin.H{"title": "Risk Scores"})
		})

		pages.GET("/committees", func(c *gin.Context) {
			c.HTML(http.StatusOK, "committees.html", gin.H{"title": "Committees"})
		})
	}

	// Start server
	log.Printf("INFORM Risk Management System starting on port %s", cfg.ServerPort)
	log.Printf("Access the application at: http://localhost:%s", cfg.ServerPort)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// authPageMiddleware checks for auth cookie on page requests
func authPageMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := c.Cookie("auth_token")
		if err != nil {
			c.Redirect(http.StatusFound, "/login?redirect="+c.Request.URL.Path)
			c.Abort()
			return
		}
		c.Next()
	}
}
