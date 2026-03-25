package handlers

import (
	_ "embed"
	"encoding/json"
	"net/http"

	"kards-backend-go/internal/database"
	"kards-backend-go/internal/models"

	"github.com/gin-gonic/gin"
)

// 使用 embed 导入 JSON 文件
//
//go:embed items_library.json
var itemsJSON []byte

var allItems []models.Item

func init() {
	// 启动时解析 JSON 文件
	if err := json.Unmarshal(itemsJSON, &allItems); err != nil {
		panic("failed to parse items JSON: " + err.Error())
	}
}

// GetItems 返回玩家物品列表和已装备物品
func GetItems(c *gin.Context) {
	playerID := c.Param("player_id")

	var user models.User
	if err := database.DB.First(&user, "id = ?", playerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	// 如果用户数据库里没有物品，直接给他全量库
	if user.Items == nil || len(user.Items) == 0 {
		user.Items = allItems
	}

	if user.EquippedItems == nil {
		user.EquippedItems = []models.Item{}
	}

	c.JSON(http.StatusOK, gin.H{
		"player_id":      playerID,
		"items":          user.Items,
		"equipped_items": user.EquippedItems,
	})
}

// ChangeItem 更换玩家装备的物品
func ChangeItem(c *gin.Context) {
	playerID := c.Param("player_id")

	var req struct {
		Slot   string `json:"slot"`    // 要替换的槽位
		ItemID string `json:"item_id"` // 新装备的物品
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", playerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	// 如果没有物品就赋值全量库
	if user.Items == nil || len(user.Items) == 0 {
		user.Items = allItems
	}

	// 替换装备
	updated := false
	for i, eq := range user.EquippedItems {
		if eq.Slot == req.Slot {
			user.EquippedItems[i].ItemID = req.ItemID
			updated = true
			break
		}
	}
	if !updated {
		user.EquippedItems = append(user.EquippedItems, models.Item{
			ItemID: req.ItemID,
			Slot:   req.Slot,
		})
	}

	// 保存到数据库
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"player_id":      playerID,
		"equipped_items": user.EquippedItems,
	})
}
