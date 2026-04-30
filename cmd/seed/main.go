package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

const base = "http://localhost:8080/api/v1"

var token string

func post(path string, body map[string]any) map[string]any {
    b, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", base+path, bytes.NewBuffer(b))
    req.Header.Set("Content-Type", "application/json")
    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }
    res, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Fatalf("POST %s failed: %v", path, err)
    }
    defer res.Body.Close()
    var out map[string]any
    json.NewDecoder(res.Body).Decode(&out)
    return out
}

func main() {
    fmt.Println("Logging in...")
    login := post("/auth/login", map[string]any{
        "shop_id": "15e5fb6a-0c41-4297-bb87-1c5b73d216b6",
        "pin":     "1234",
    })
    t, ok := login["token"].(string)
    if !ok {
        log.Fatalf("Login failed: %v", login)
    }
    token = t
    fmt.Println("Logged in OK")

    shopID := "15e5fb6a-0c41-4297-bb87-1c5b73d216b6"

    fmt.Println("Creating categories...")
    cats := []map[string]any{
        {"name": "Hot Beverages", "description": "Tea, coffee and more"},
        {"name": "Cold Drinks",   "description": "Juices, lassi, shakes"},
        {"name": "Snacks",        "description": "Samosa, pakora, chaat"},
        {"name": "Main Course",   "description": "Rice, curry, roti"},
        {"name": "Sweets",        "description": "Mithai and desserts"},
        {"name": "Combos",        "description": "Value meal combos"},
    }
    catIDs := map[string]string{}
    for _, c := range cats {
        res := post("/categories", c)
        id, _ := res["id"].(string)
        catIDs[c["name"].(string)] = id
        fmt.Printf("  Category: %s -> %s\n", c["name"], id)
    }

    fmt.Println("Creating items...")
    type item struct {
        name  string
        cat   string
        price float64
        sku   string
    }
    items := []item{
        {"Masala Chai",         "Hot Beverages", 20,  "BEV-001"},
        {"Plain Chai",          "Hot Beverages", 15,  "BEV-002"},
        {"Filter Coffee",       "Hot Beverages", 25,  "BEV-003"},
        {"Ginger Lemon Tea",    "Hot Beverages", 30,  "BEV-004"},
        {"Hot Chocolate",       "Hot Beverages", 60,  "BEV-005"},
        {"Mango Lassi",         "Cold Drinks",   50,  "DRK-001"},
        {"Sweet Lassi",         "Cold Drinks",   40,  "DRK-002"},
        {"Sugarcane Juice",     "Cold Drinks",   30,  "DRK-003"},
        {"Nimbu Pani",          "Cold Drinks",   25,  "DRK-004"},
        {"Cold Coffee",         "Cold Drinks",   70,  "DRK-005"},
        {"Aam Panna",           "Cold Drinks",   35,  "DRK-006"},
        {"Samosa (2 pcs)",      "Snacks",        30,  "SNK-001"},
        {"Aloo Tikki",          "Snacks",        40,  "SNK-002"},
        {"Pani Puri (6 pcs)",   "Snacks",        50,  "SNK-003"},
        {"Bread Pakora",        "Snacks",        35,  "SNK-004"},
        {"Dhokla Plate",        "Snacks",        60,  "SNK-005"},
        {"Bhel Puri",           "Snacks",        45,  "SNK-006"},
        {"Vada Pav",            "Snacks",        25,  "SNK-007"},
        {"Dal Tadka + Rice",    "Main Course",  120,  "MAIN-001"},
        {"Chole Bhature",       "Main Course",  100,  "MAIN-002"},
        {"Rajma Chawal",        "Main Course",  110,  "MAIN-003"},
        {"Paneer Butter Masala","Main Course",  160,  "MAIN-004"},
        {"Jeera Rice",          "Main Course",   80,  "MAIN-005"},
        {"Plain Paratha (2)",   "Main Course",   50,  "MAIN-006"},
        {"Aloo Paratha",        "Main Course",   70,  "MAIN-007"},
        {"Gulab Jamun (2)",     "Sweets",        40,  "SWT-001"},
        {"Rasgulla (2)",        "Sweets",        40,  "SWT-002"},
        {"Gajar Halwa",         "Sweets",        60,  "SWT-003"},
        {"Kheer",               "Sweets",        55,  "SWT-004"},
        {"Jalebi (100g)",       "Sweets",        50,  "SWT-005"},
        {"Chai + Samosa",       "Combos",        45,  "CMB-001"},
        {"Thali (Full)",        "Combos",       200,  "CMB-002"},
        {"Snack Combo",         "Combos",       120,  "CMB-003"},
    }

    itemIDs := []string{}
    for _, it := range items {
        catID := catIDs[it.cat]
        res := post("/items", map[string]any{
            "name":        it.name,
            "category_id": catID,
            "price":       it.price,
            "sku":         it.sku,
        })
        id, _ := res["id"].(string)
        itemIDs = append(itemIDs, id)
        fmt.Printf("  Item: %-25s Rs.%-6.0f -> %s\n", it.name, it.price, id)
    }

    fmt.Println("Assigning items to shop...")
    for _, id := range itemIDs {
        if id == "" {
            continue
        }
        post(fmt.Sprintf("/shops/%s/items", shopID), map[string]any{
            "item_id": id,
        })
    }
    fmt.Printf("\nDone! %d items assigned to shop.\n", len(itemIDs))
    fmt.Println("Press R in the Flutter terminal to reload the catalog.")
}
