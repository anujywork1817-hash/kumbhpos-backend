package jobs

import (
"fmt"
"log"
"time"

"github.com/beellz/kumbhpos/internal/dashboard"
"github.com/beellz/kumbhpos/internal/inventory"
"github.com/beellz/kumbhpos/internal/settlement"
"github.com/beellz/kumbhpos/pkg/db"
"github.com/robfig/cron/v3"
)

var Scheduler *cron.Cron

func StartScheduler() {
Scheduler = cron.New(cron.WithSeconds())

// EOD Settlement — runs every day at 11:59 PM
Scheduler.AddFunc("0 59 23 * * *", func() {
log.Println("[CRON] Running EOD settlement...")
runEODSettlement()
})

// Low stock check — runs every hour
Scheduler.AddFunc("0 0 * * * *", func() {
log.Println("[CRON] Checking low stock levels...")
checkLowStock()
})

// Terminal stale check — runs every 5 minutes
Scheduler.AddFunc("0 */5 * * * *", func() {
checkStaleTerminals()
})

// Mark offline terminals — runs every 2 minutes
Scheduler.AddFunc("0 */2 * * * *", func() {
markOfflineTerminals()
})

Scheduler.Start()
log.Println("Scheduler started — EOD@23:59, LowStock@hourly, TerminalCheck@5min")
}

func runEODSettlement() {
date := time.Now().Format("2006-01-02")
report, err := settlement.GenerateEODReport(date)
if err != nil {
log.Println("[CRON][EOD] Report generation failed:", err)
return
}

filename, err := settlement.GeneratePDF(report)
if err != nil {
log.Println("[CRON][EOD] PDF generation failed:", err)
return
}

log.Printf("[CRON][EOD] Report saved: %s | GMV: %.2f | Beellz: %.2f\n",
filename, report.TotalGMV, report.BeellzShare)

// Broadcast to dashboard
dashboard.PushLiveUpdate("eod_settlement_done", map[string]interface{}{
"date":          report.Date,
"total_gmv":     report.TotalGMV,
"beellz_share":  report.BeellzShare,
"total_txns":    report.TotalTxns,
"pdf_file":      filename,
})
}

func checkLowStock() {
alerts, err := inventory.GetLowStockAlerts()
if err != nil {
log.Println("[CRON][STOCK] Low stock check failed:", err)
return
}
if len(alerts) == 0 {
return
}

log.Printf("[CRON][STOCK] %d low stock alerts found\n", len(alerts))

for _, a := range alerts {
log.Printf("[CRON][STOCK] ALERT: %s @ %s — qty: %d (threshold: %d)\n",
a.ItemName, a.ShopName, a.StockQty, a.LowStockThreshold)
}

// Broadcast to dashboard
dashboard.PushLiveUpdate("low_stock_alert", map[string]interface{}{
"count":  len(alerts),
"alerts": alerts,
})
}

func checkStaleTerminals() {
rows, err := db.DB.Queryx(`
SELECT shop_id, last_heartbeat
FROM terminal_sync_log
WHERE is_online = true
  AND (NOW() - last_heartbeat) > INTERVAL '5 minutes'
`)
if err != nil {
return
}
defer rows.Close()

for rows.Next() {
var shopID string
var lastHB time.Time
rows.Scan(&shopID, &lastHB)
staleFor := time.Since(lastHB).Round(time.Second)
log.Printf("[CRON][TERMINAL] Stale terminal: shop %s — last seen %s ago\n", shopID, staleFor)
dashboard.PushLiveUpdate("terminal_stale", map[string]interface{}{
"shop_id":    shopID,
"stale_for":  fmt.Sprintf("%s", staleFor),
"last_heartbeat": lastHB.Format(time.RFC3339),
})
}
}

func markOfflineTerminals() {
_, err := db.DB.Exec(`
UPDATE terminal_sync_log
SET is_online = false
WHERE is_online = true
  AND (NOW() - last_heartbeat) > INTERVAL '2 minutes'
`)
if err != nil {
log.Println("[CRON][TERMINAL] Mark offline failed:", err)
}
}

// ManualEOD — can be triggered via API for testing
func ManualEOD(date string) error {
if date == "" {
date = time.Now().Format("2006-01-02")
}
report, err := settlement.GenerateEODReport(date)
if err != nil {
return err
}
_, err = settlement.GeneratePDF(report)
if err != nil {
return err
}
dashboard.PushLiveUpdate("eod_settlement_done", map[string]interface{}{
"date":         report.Date,
"total_gmv":    report.TotalGMV,
"beellz_share": report.BeellzShare,
"total_txns":   report.TotalTxns,
})
log.Printf("[MANUAL EOD] GMV: %.2f | Beellz: %.2f\n", report.TotalGMV, report.BeellzShare)
return nil
}
