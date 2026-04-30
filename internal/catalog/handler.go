package catalog

import (
"net/http"
"github.com/gin-gonic/gin"
)

func CreateCategoryHandler(c *gin.Context) {
var req CreateCategoryRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
cat, err := CreateCategory(req)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusCreated, cat)
}

func ListCategoriesHandler(c *gin.Context) {
cats, err := ListCategories()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
if cats == nil { cats = []Category{} }
c.JSON(http.StatusOK, cats)
}

func CreateItemHandler(c *gin.Context) {
var req CreateItemRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
item, err := CreateItem(req)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusCreated, item)
}

func ListItemsHandler(c *gin.Context) {
items, err := ListItems()
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
if items == nil { items = []Item{} }
c.JSON(http.StatusOK, items)
}

func AssignItemHandler(c *gin.Context) {
shopID := c.Param("shop_id")
var req AssignItemRequest
if err := c.ShouldBindJSON(&req); err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
return
}
if err := AssignItemToShop(shopID, req); err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
c.JSON(http.StatusOK, gin.H{"message": "item assigned"})
}

func ShopCatalogHandler(c *gin.Context) {
shopID := c.Param("shop_id")
items, err := GetShopCatalog(shopID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}
if items == nil { items = []Item{} }
c.JSON(http.StatusOK, items)
}
