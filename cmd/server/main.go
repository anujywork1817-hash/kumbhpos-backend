package main

import (
    "log"
    "os"

    "github.com/beellz/kumbhpos/config"
    "github.com/beellz/kumbhpos/internal/admin"
    "github.com/beellz/kumbhpos/internal/attendance"
    "github.com/beellz/kumbhpos/internal/auth"
    "github.com/beellz/kumbhpos/internal/catalog"
    "github.com/beellz/kumbhpos/internal/dashboard"
    "github.com/beellz/kumbhpos/internal/inventory"
    "github.com/beellz/kumbhpos/internal/middleware"
    "github.com/beellz/kumbhpos/internal/payment"
    "github.com/beellz/kumbhpos/internal/settlement"
    "github.com/beellz/kumbhpos/internal/shop"
    "github.com/beellz/kumbhpos/internal/staff"
    "github.com/beellz/kumbhpos/internal/transaction"
    "github.com/beellz/kumbhpos/pkg/db"
    redisclient "github.com/beellz/kumbhpos/pkg/redis"
    "github.com/gin-contrib/cors"
    "github.com/gin-gonic/gin"
)

func main() {
    config.Load()
    log.Println("Starting KumbhPOS...")
    log.Println("DATABASE_URL set:", os.Getenv("DATABASE_URL") != "")
    log.Println("REDIS_URL set:", os.Getenv("REDIS_URL") != "")

    if err := db.Connect(); err != nil {
        log.Fatal("DB connection failed:", err)
    }
    log.Println("PostgreSQL connected")

    if err := redisclient.Connect(); err != nil {
        log.Println("Redis connection failed (continuing):", err)
    } else {
        log.Println("Redis connected")
    }

    go dashboard.GlobalHub.Run()
    log.Println("WebSocket hub started")

    r := gin.Default()

    r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"*"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: false,
    }))

    // Public
    r.POST("/api/v1/auth/login", auth.LoginHandler)
    r.POST("/api/v1/payments/webhook", payment.WebhookHandler)
    r.GET("/ws", dashboard.WSHandler)

    api := r.Group("/api/v1", middleware.AuthRequired())

    // Shops
    api.POST("/shops", shop.CreateShopHandler)
    api.GET("/shops", shop.ListShopsHandler)

    // Staff
    api.POST("/staff", staff.CreateStaffHandler)
    api.GET("/shops/:shop_id/staff", staff.ListStaffHandler)

    // Catalog
    api.POST("/categories", catalog.CreateCategoryHandler)
    api.GET("/categories", catalog.ListCategoriesHandler)
    api.POST("/items", catalog.CreateItemHandler)
    api.GET("/items", catalog.ListItemsHandler)
    api.POST("/shops/:shop_id/items", catalog.AssignItemHandler)
    api.GET("/shops/:shop_id/catalog", catalog.ShopCatalogHandler)

    // Transactions
    api.POST("/transactions/checkout", transaction.CheckoutHandler)
    api.GET("/transactions", transaction.ListTransactionsHandler)
    api.GET("/transactions/:id", transaction.GetTransactionHandler)
    api.GET("/transactions/:id/items", transaction.GetTransactionItemsHandler)
    api.POST("/transactions/:id/confirm-upi", transaction.ConfirmUPIHandler)

    // Payments
    api.POST("/payments/qr", payment.CreateQRHandler)
    api.GET("/payments/:id/status", payment.CheckStatusHandler)

    // Dashboard
    api.GET("/dashboard/stats", dashboard.StatsHandler)
    api.GET("/dashboard/leaderboard", dashboard.LeaderboardHandler)

    // Settlement
    api.GET("/settlement/eod", settlement.EODReportHandler)
    api.GET("/settlement/eod/pdf", settlement.EODPDFHandler)

    // Inventory
    api.GET("/inventory/stock", inventory.StockLevelsHandler)
    api.GET("/inventory/low-stock", inventory.LowStockHandler)
    api.POST("/inventory/adjust", inventory.AdjustStockHandler)
    api.POST("/inventory/transfer", inventory.TransferStockHandler)
    api.POST("/inventory/restock", inventory.CreateRestockHandler)
    api.GET("/inventory/restock", inventory.RestockListHandler)
    api.POST("/inventory/restock/:id/approve", inventory.ApproveRestockHandler)

    // Admin
    api.GET("/admin/stats", admin.GlobalStatsHandler)
    api.GET("/admin/shops", admin.ListShopsHandler)
    api.PATCH("/admin/shops/:id", admin.ToggleShopHandler)
    api.GET("/admin/staff", admin.ListStaffHandler)
    api.PATCH("/admin/staff/:id/pin", admin.ResetPINHandler)
    api.PATCH("/admin/staff/:id/active", admin.ToggleStaffHandler)
    api.PATCH("/admin/staff/:id/role", admin.ChangeRoleHandler)
    api.GET("/admin/items", admin.ListItemsHandler)
    api.PATCH("/admin/items/:id/price", admin.UpdatePriceHandler)
    api.PATCH("/admin/items/:id/active", admin.ToggleItemHandler)

    // Attendance
    api.POST("/attendance/clock-in", attendance.ClockInHandler)
    api.POST("/attendance/clock-out", attendance.ClockOutHandler)
    api.GET("/attendance", attendance.ListAttendanceHandler)
    api.GET("/attendance/active", attendance.ActiveShiftsHandler)
    api.GET("/attendance/status/:staff_id", attendance.StaffStatusHandler)

    // Render uses PORT, fallback to SERVER_PORT for local
    port := os.Getenv("PORT")
    if port == "" {
        port = os.Getenv("SERVER_PORT")
    }
    if port == "" {
        port = "8080"
    }
    log.Printf("KumbhPOS Hub running on :%s\n", port)
    r.Run(":" + port)
}
