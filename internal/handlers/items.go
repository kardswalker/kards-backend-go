package handlers

import (
	_ "embed"
	"encoding/json"
	"log"
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

// GetItems 返回玩家物品列表和已装备物品（包括全量库 + 玩家物品）
func GetItems(c *gin.Context) {
	playerID := c.Param("player_id")

	var user models.User
	if err := database.DB.First(&user, "id = ?", playerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	// 反序列化Items
	if user.ItemsJSON != "" {
		if err := json.Unmarshal([]byte(user.ItemsJSON), &user.Items); err != nil {
			user.Items = []models.Item{}
		}
	} else {
		user.Items = []models.Item{}
	}

	// 反序列化EquippedItems
	if user.EquippedJSON != "" {
		if err := json.Unmarshal([]byte(user.EquippedJSON), &user.EquippedItems); err != nil {
			user.EquippedItems = []models.Item{}
		}
	} else {
		user.EquippedItems = []models.Item{}
	}

	// 合并全量库 + 玩家物品（去重）
	itemMap := make(map[string]models.Item)
	for _, item := range allItems {
		itemMap[item.ItemID] = item
	}
	for _, item := range user.Items {
		itemMap[item.ItemID] = item
	}

	// 转换为切片
	mergedItems := make([]models.Item, 0, len(itemMap))
	for _, item := range itemMap {
		mergedItems = append(mergedItems, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"player_id":      playerID,
		"items":          mergedItems,
		"equipped_items": user.EquippedItems,
	})
}

// ChangeItem 更换玩家装备的物品（匹配Node.js的equipItem逻辑）
func ChangeItem(c *gin.Context) {
	playerID := c.Param("player_id")

	// 接收完整的Item对象
	var item models.Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// 检查必需字段
	if item.ItemID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id is required"})
		return
	}
	if item.Slot == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "slot is required"})
		return
	}
	if item.Faction == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "faction is required"})
		return
	}

	var user models.User
	if err := database.DB.First(&user, "id = ?", playerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "player not found"})
		return
	}

	// 反序列化Items
	if user.ItemsJSON != "" {
		json.Unmarshal([]byte(user.ItemsJSON), &user.Items)
	}
	if user.Items == nil {
		user.Items = []models.Item{}
	}

	// 反序列化EquippedItems
	if user.EquippedJSON != "" {
		json.Unmarshal([]byte(user.EquippedJSON), &user.EquippedItems)
	}
	if user.EquippedItems == nil {
		user.EquippedItems = []models.Item{}
	}

	// 移除所有相同slot和faction的装备
	newEquipped := []models.Item{}
	for _, eq := range user.EquippedItems {
		// 如果slot和faction都相同，则跳过（删除）；否则保留
		if !(eq.Slot == item.Slot && eq.Faction == item.Faction) {
			newEquipped = append(newEquipped, eq)
		}
	}
	user.EquippedItems = newEquipped

	// 添加新装备
	user.EquippedItems = append(user.EquippedItems, item)

	// 序列化EquippedItems
	equippedJSON, err := json.Marshal(user.EquippedItems)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal equipped items"})
		return
	}
	user.EquippedJSON = string(equippedJSON)

	// 保存到数据库
	if err := database.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user"})
		return
	}

	// 调试日志
	log.Printf("装备物品成功: player_id=%s, item_id=%s, slot=%s, faction=%s", playerID, item.ItemID, item.Slot, item.Faction)

	c.JSON(http.StatusOK, gin.H{
		"player_id":      playerID,
		"equipped_items": user.EquippedItems,
	})

	// 调试日志
	log.Printf("装备物品成功: player_id=%s, item_id=%s, slot=%s, faction=%s", playerID, item.ItemID, item.Slot, item.Faction)
}
