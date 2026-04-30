package dashboard

import (
"github.com/beellz/kumbhpos/pkg/db"
)

type LiveStats struct {
TotalGMV       float64     `db:"total_gmv"      json:"total_gmv"`
TotalTxns      int         `db:"total_txns"     json:"total_txns"`
UPITotal       float64     `db:"upi_total"      json:"upi_total"`
CashTotal      float64     `db:"cash_total"     json:"cash_total"`
AvgOrderValue  float64     `db:"avg_order"      json:"avg_order_value"`
}

type ShopSales struct {
ShopID    string  `db:"shop_id"  json:"shop_id"`
ShopName  string  `db:"name"     json:"shop_name"`
GMV       float64 `db:"gmv"      json:"gmv"`
TxnCount  int     `db:"txn_count" json:"txn_count"`
}

func GetLiveStats() (LiveStats, error) {
var stats LiveStats
err := db.DB.Get(&stats, `
SELECT
COALESCE(SUM(total_amount), 0)                             AS total_gmv,
COUNT(*)                                                   AS total_txns,
COALESCE(SUM(CASE WHEN payment_mode='upi'  THEN total_amount ELSE 0 END), 0) AS upi_total,
COALESCE(SUM(CASE WHEN payment_mode='cash' THEN total_amount ELSE 0 END), 0) AS cash_total,
COALESCE(AVG(total_amount), 0)                             AS avg_order
FROM transactions
WHERE payment_status='confirmed'
  AND created_at >= CURRENT_DATE
`)
return stats, err
}

func GetShopLeaderboard() ([]ShopSales, error) {
var list []ShopSales
err := db.DB.Select(&list, `
SELECT
t.shop_id,
s.name,
COALESCE(SUM(t.total_amount), 0) AS gmv,
COUNT(*) AS txn_count
FROM transactions t
JOIN shops s ON s.id = t.shop_id
WHERE t.payment_status='confirmed'
  AND t.created_at >= CURRENT_DATE
GROUP BY t.shop_id, s.name
ORDER BY gmv DESC
`)
return list, err
}

func PushLiveUpdate(eventType string, data interface{}) {
GlobalHub.Broadcast(eventType, data)
}
