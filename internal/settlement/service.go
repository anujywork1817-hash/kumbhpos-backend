package settlement

import (
"fmt"
"time"

"github.com/beellz/kumbhpos/pkg/db"
"github.com/jung-kurt/gofpdf"
)

type ShopSettlement struct {
ShopID    string  `db:"shop_id"   json:"shop_id"`
ShopName  string  `db:"name"      json:"shop_name"`
TotalGMV  float64 `db:"total_gmv" json:"total_gmv"`
UPITotal  float64 `db:"upi_total" json:"upi_total"`
CashTotal float64 `db:"cash_total" json:"cash_total"`
TxnCount  int     `db:"txn_count" json:"txn_count"`
BeellzCut float64 `json:"beellz_cut"`
}

type EODReport struct {
Date        string           `json:"date"`
TotalGMV    float64          `json:"total_gmv"`
TotalUPI    float64          `json:"total_upi"`
TotalCash   float64          `json:"total_cash"`
TotalTxns   int              `json:"total_txns"`
BeellzShare float64          `json:"beellz_share_40pct"`
Shops       []ShopSettlement `json:"shops"`
}

func GenerateEODReport(date string) (EODReport, error) {
if date == "" {
date = time.Now().Format("2006-01-02")
}

var shops []ShopSettlement
err := db.DB.Select(&shops, `
SELECT
t.shop_id,
s.name,
COALESCE(SUM(t.total_amount), 0)                                              AS total_gmv,
COALESCE(SUM(CASE WHEN t.payment_mode='upi'  THEN t.total_amount ELSE 0 END),0) AS upi_total,
COALESCE(SUM(CASE WHEN t.payment_mode='cash' THEN t.total_amount ELSE 0 END),0) AS cash_total,
COUNT(*) AS txn_count
FROM transactions t
JOIN shops s ON s.id = t.shop_id
WHERE t.payment_status = 'confirmed'
  AND DATE(t.created_at) = $1
GROUP BY t.shop_id, s.name
ORDER BY total_gmv DESC
`, date)
if err != nil {
return EODReport{}, err
}

var totalGMV, totalUPI, totalCash float64
var totalTxns int
for i, s := range shops {
totalGMV += s.TotalGMV
totalUPI += s.UPITotal
totalCash += s.CashTotal
totalTxns += s.TxnCount
shops[i].BeellzCut = s.TotalGMV * 0.40
}

return EODReport{
Date:        date,
TotalGMV:    totalGMV,
TotalUPI:    totalUPI,
TotalCash:   totalCash,
TotalTxns:   totalTxns,
BeellzShare: totalGMV * 0.40,
Shops:       shops,
}, nil
}

func GeneratePDF(report EODReport) (string, error) {
pdf := gofpdf.New("P", "mm", "A4", "")
pdf.AddPage()

// Header
pdf.SetFont("Arial", "B", 20)
pdf.SetFillColor(30, 30, 30)
pdf.SetTextColor(255, 255, 255)
pdf.CellFormat(0, 14, "KumbhPOS — End of Day Settlement Report", "", 1, "C", true, 0, "")
pdf.Ln(4)

// Date & summary
pdf.SetFont("Arial", "", 11)
pdf.SetTextColor(0, 0, 0)
pdf.CellFormat(0, 8, fmt.Sprintf("Date: %s     Total Transactions: %d", report.Date, report.TotalTxns), "", 1, "L", false, 0, "")
pdf.Ln(2)

// Summary boxes
pdf.SetFont("Arial", "B", 12)
pdf.SetFillColor(240, 240, 240)
pdf.CellFormat(60, 10, fmt.Sprintf("Total GMV: Rs.%.2f", report.TotalGMV), "1", 0, "C", true, 0, "")
pdf.CellFormat(60, 10, fmt.Sprintf("UPI: Rs.%.2f", report.TotalUPI), "1", 0, "C", true, 0, "")
pdf.CellFormat(60, 10, fmt.Sprintf("Cash: Rs.%.2f", report.TotalCash), "1", 1, "C", true, 0, "")
pdf.Ln(2)

// Beellz profit share
pdf.SetFont("Arial", "B", 13)
pdf.SetFillColor(220, 240, 220)
pdf.CellFormat(0, 12,
fmt.Sprintf("Beellz Technologies 40%% Profit Share: Rs.%.2f", report.BeellzShare),
"1", 1, "C", true, 0, "")
pdf.Ln(6)

// Shop breakdown table header
pdf.SetFont("Arial", "B", 11)
pdf.SetFillColor(50, 50, 50)
pdf.SetTextColor(255, 255, 255)
pdf.CellFormat(70, 9, "Shop Name", "1", 0, "L", true, 0, "")
pdf.CellFormat(30, 9, "Txns", "1", 0, "C", true, 0, "")
pdf.CellFormat(35, 9, "GMV (Rs.)", "1", 0, "C", true, 0, "")
pdf.CellFormat(35, 9, "UPI (Rs.)", "1", 0, "C", true, 0, "")
pdf.CellFormat(20, 9, "Beellz 40%", "1", 1, "C", true, 0, "")

// Shop rows
pdf.SetFont("Arial", "", 10)
pdf.SetTextColor(0, 0, 0)
for i, s := range report.Shops {
if i%2 == 0 {
pdf.SetFillColor(255, 255, 255)
} else {
pdf.SetFillColor(248, 248, 248)
}
pdf.CellFormat(70, 8, s.ShopName, "1", 0, "L", true, 0, "")
pdf.CellFormat(30, 8, fmt.Sprintf("%d", s.TxnCount), "1", 0, "C", true, 0, "")
pdf.CellFormat(35, 8, fmt.Sprintf("%.2f", s.TotalGMV), "1", 0, "C", true, 0, "")
pdf.CellFormat(35, 8, fmt.Sprintf("%.2f", s.UPITotal), "1", 0, "C", true, 0, "")
pdf.CellFormat(20, 8, fmt.Sprintf("%.2f", s.BeellzCut), "1", 1, "C", true, 0, "")
}

// Save PDF
filename := fmt.Sprintf("settlement_%s.pdf", report.Date)
if err := pdf.OutputFileAndClose(filename); err != nil {
return "", err
}
return filename, nil
}
