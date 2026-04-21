package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"book-trading/backend/internal/database"
	"book-trading/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CreateBatch 发布一批书
func CreateBatch(c *gin.Context) {
	var req struct {
		Title       string   `json:"title" binding:"required,min=2,max=200"`
		Description string   `json:"description" binding:"max=2000"`
		Image       string   `json:"image"`
		BookNames   []string `json:"book_names" binding:"required,min=1,max=50"`
		Contact     string   `json:"contact" binding:"required,min=3,max=100"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "请先登录",
			Data:    nil,
		})
		return
	}

	// 序列化 JSON
	bookNamesJSON, _ := json.Marshal(req.BookNames)

	// 已售出列表初始为空
	soldBookNamesJSON, _ := json.Marshal([]string{})

	// 确定初始状态
	status := "available"
	if len(req.BookNames) == 0 {
		status = "sold"
	}

	batch := models.Batch{
		Title:         req.Title,
		Description:   req.Description,
		Image:         req.Image,
		BookNames:     string(bookNamesJSON),
		SoldBookNames: string(soldBookNamesJSON),
		Contact:       req.Contact,
		Status:        status,
		UserID:        userID.(uint),
	}

	if err := database.DB.Create(&batch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "发布失败: " + err.Error(),
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "发布成功",
		Data:    batch,
	})
}

// UpdateBookSoldStatus 更新单本书的售出状态（移动到已售出列表或移回）
func UpdateBookSoldStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "无效的ID",
			Data:    nil,
		})
		return
	}

	var req struct {
		BookName string `json:"book_name" binding:"required"` // 书名
		Sold     bool   `json:"sold"`                         // true:标记售出, false:取消售出
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	userID, _ := c.Get("userID")
	var batch models.Batch
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&batch).Error; err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "批次不存在或无权限",
			Data:    nil,
		})
		return
	}

	// 解析当前列表
	var bookNames []string
	var soldBookNames []string

	if err := json.Unmarshal([]byte(batch.BookNames), &bookNames); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "数据解析失败",
			Data:    nil,
		})
		return
	}

	if err := json.Unmarshal([]byte(batch.SoldBookNames), &soldBookNames); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "数据解析失败",
			Data:    nil,
		})
		return
	}

	if req.Sold {
		// 标记售出：从 bookNames 移到 soldBookNames
		found := false
		newBookNames := []string{}
		for _, name := range bookNames {
			if name == req.BookName {
				found = true
				continue
			}
			newBookNames = append(newBookNames, name)
		}

		if !found {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "书籍不存在于未售出列表中",
				Data:    nil,
			})
			return
		}

		bookNames = newBookNames
		soldBookNames = append(soldBookNames, req.BookName)
	} else {
		// 取消售出：从 soldBookNames 移回 bookNames
		found := false
		newSoldBookNames := []string{}
		for _, name := range soldBookNames {
			if name == req.BookName {
				found = true
				continue
			}
			newSoldBookNames = append(newSoldBookNames, name)
		}

		if !found {
			c.JSON(http.StatusBadRequest, models.Response{
				Code:    400,
				Message: "书籍不存在于已售出列表中",
				Data:    nil,
			})
			return
		}

		soldBookNames = newSoldBookNames
		bookNames = append(bookNames, req.BookName)
	}

	// 更新状态
	status := "available"
	if len(bookNames) == 0 && len(soldBookNames) > 0 {
		status = "sold"
	} else if len(bookNames) > 0 && len(soldBookNames) > 0 {
		status = "partial"
	} else {
		status = "available"
	}

	// 序列化保存
	newBookNamesJSON, _ := json.Marshal(bookNames)
	newSoldBookNamesJSON, _ := json.Marshal(soldBookNames)

	batch.BookNames = string(newBookNamesJSON)
	batch.SoldBookNames = string(newSoldBookNamesJSON)
	batch.Status = status

	database.DB.Save(&batch)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "状态更新成功",
		Data:    batch,
	})
}

// GetBatchDetail 获取批次详情
func GetBatchDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "无效的ID",
			Data:    nil,
		})
		return
	}

	var batch models.Batch
	if err := database.DB.Preload("User").First(&batch, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "批次不存在",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data:    batch,
	})
}

// UpdateBatchImage 更新批次图片
func UpdateBatchImage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "无效的ID",
			Data:    nil,
		})
		return
	}

	var req struct {
		Image string `json:"image" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "参数错误: " + err.Error(),
			Data:    nil,
		})
		return
	}

	userID, _ := c.Get("userID")
	var batch models.Batch
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&batch).Error; err != nil {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "批次不存在或无权限",
			Data:    nil,
		})
		return
	}

	batch.Image = req.Image
	database.DB.Save(&batch)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "图片更新成功",
		Data:    batch,
	})
}

// GetBatchList 获取批次列表（只显示还有未售出书籍的批次）
func GetBatchList(c *gin.Context) {
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 12
	}
	if pageSize > 50 {
		pageSize = 50
	}

	// 基础查询：只显示还有未售出书籍的批次（book_names 不为空数组）
	query := database.DB.Model(&models.Batch{}).
		Preload("User").
		Where("JSON_LENGTH(book_names) > 0")

	// 关键词搜索（支持标题、描述、未售出的书籍名称）
	if keyword != "" {
		query = query.Where(
			"title LIKE ? OR description LIKE ? OR JSON_SEARCH(book_names, 'one', ?) IS NOT NULL",
			"%"+keyword+"%",
			"%"+keyword+"%",
			"%"+keyword+"%",
		)
	}

	// 分页查询（按更新时间倒序，有未售出书籍的都在这里）
	var batches []models.Batch
	offset := (page - 1) * pageSize
	query.Order("updated_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&batches)

	// 统计总数
        var total int64
        query.Count(&total)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data: gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
			"batches":   batches,
		},
	})
}

// GetMyBatches 获取我的发布（全部，包括已售完的）
func GetMyBatches(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.Response{
			Code:    401,
			Message: "请先登录",
			Data:    nil,
		})
		return
	}

	var batches []models.Batch
	database.DB.Where("user_id = ?", userID).Order("updated_at DESC").Find(&batches)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "success",
		Data:    batches,
	})
}

// DeleteBatch 删除批次
func DeleteBatch(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "无效的ID",
			Data:    nil,
		})
		return
	}

	isAdmin, _ := c.Get("isAdmin")
	var result *gorm.DB
	if isAdminBool, ok := isAdmin.(bool); ok && isAdminBool {
		result = database.DB.Where("id = ?", id).Delete(&models.Batch{})
	} else {
		userID, _ := c.Get("userID")
		result = database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Batch{})
	}

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "删除失败",
			Data:    nil,
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, models.Response{
			Code:    404,
			Message: "批次不存在或无权限",
			Data:    nil,
		})
		return
	}

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "删除成功",
		Data:    nil,
	})
}
