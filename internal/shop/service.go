package shop

import (
"github.com/beellz/kumbhpos/pkg/db"
)

type Shop struct {
ID       string `db:"id"       json:"id"`
Name     string `db:"name"     json:"name"`
Location string `db:"location" json:"location"`
IsActive bool   `db:"is_active" json:"is_active"`
}

type CreateShopRequest struct {
Name     string `json:"name"     binding:"required"`
Location string `json:"location"`
}

func CreateShop(req CreateShopRequest) (Shop, error) {
var s Shop
err := db.DB.QueryRowx(
`INSERT INTO shops (name, location) VALUES ($1, $2)
 RETURNING id, name, location, is_active`,
req.Name, req.Location,
).StructScan(&s)
return s, err
}

func ListShops() ([]Shop, error) {
var shops []Shop
err := db.DB.Select(&shops, `SELECT id, name, location, is_active FROM shops ORDER BY name`)
return shops, err
}
