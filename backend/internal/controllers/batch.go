package controllers

import (
	"encoding/json" // 新增：标准 JSON 库
	"net/http"
	"strconv"

	"book-trading/backend/internal/database"
	"book-trading/backend/internal/models"

	"github.com/gin-gonic/gin"
)

// CreateBatch 发布一批书（包含书名列表）
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

	// 初始化售出状态（全部 false）
	soldStatus := make([]bool, len(req.BookNames))
	for i := range soldStatus {
		soldStatus[i] = false
	}

	// 将切片转换为 JSON 字符串（使用标准库 encoding/json）
	bookNamesJSON, err := json.Marshal(req.BookNames)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "数据序列化失败",
			Data:    nil,
		})
		return
	}

	soldStatusJSON, err := json.Marshal(soldStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "数据序列化失败",
			Data:    nil,
		})
		return
	}

	batch := models.Batch{
		Title:       req.Title,
		Description: req.Description,
		Image:       req.Image,
		BookNames:   string(bookNamesJSON),  // []byte 转 string
		SoldStatus:  string(soldStatusJSON), // []byte 转 string
		Contact:     req.Contact,
		Status:      "available",
		UserID:      userID.(uint),
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

// UpdateBookSoldStatus 更新单本书的售出状态（打勾/取消打勾）
func UpdateBookSoldStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "无效的ID",
			Data:    nil,
		})
		return
	}

	var req struct {
		BookIndex int  `json:"book_index" ` //required要求不能为0值，但是序号从0开始
		Sold      bool `json:"sold"`
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

	// 解析当前售出状态（从 JSON 字符串解析为 []bool）
	var soldStatus []bool
	if err := json.Unmarshal([]byte(batch.SoldStatus), &soldStatus); err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "数据解析失败",
			Data:    nil,
		})
		return
	}

	// 检查索引是否有效
	if req.BookIndex < 0 || req.BookIndex >= len(soldStatus) {
		c.JSON(http.StatusBadRequest, models.Response{
			Code:    400,
			Message: "书本索引无效",
			Data:    nil,
		})
		return
	}

	// 更新状态
	soldStatus[req.BookIndex] = req.Sold

	// 重新序列化为 JSON 字符串
	newSoldStatusJSON, err := json.Marshal(soldStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.Response{
			Code:    500,
			Message: "数据序列化失败",
			Data:    nil,
		})
		return
	}
	batch.SoldStatus = string(newSoldStatusJSON)

	// 检查是否全部售出
	allSold := true
	for _, sold := range soldStatus {
		if !sold {
			allSold = false
			break
		}
	}
	if allSold {
		batch.Status = "sold"
	} else {
		batch.Status = "available"
	}

	database.DB.Save(&batch)

	c.JSON(http.StatusOK, models.Response{
		Code:    0,
		Message: "状态更新成功",
		Data:    batch,
	})
}

// GetBatchDetail 获取批次详情（公开，用于展示）
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

// GetBatchList 获取批次列表（公开接口）
// 支持搜索：标题、描述、书籍列表中的书名
func GetBatchList(c *gin.Context) {
	// 1. 获取查询参数
	keyword := c.Query("keyword")                                  // 搜索关键词
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))           // 页码，默认第1页
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12")) // 每页数量，默认12条

	// 2. 参数校验
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 12
	}
	if pageSize > 50 {
		pageSize = 50 // 限制最大50条
	}

	// 3. 构建查询（只显示在售的批次）
	query := database.DB.Model(&models.Batch{}).
		Preload("User").
		Where("status != ?", "sold")

	// 4. 关键词搜索（核心逻辑）
	if keyword != "" {
		// 使用 OR 条件同时搜索三个字段：
		// - title: 批次标题
		// - description: 批次描述
		// - book_names: JSON 数组中的书名
		//
		// JSON_SEARCH(book_names, 'one', '%keyword%') 的作用：
		// 在 book_names 这个 JSON 数组中搜索包含 keyword 的字符串
		// 如果找到，返回该元素的路径（如 "$[0]"），否则返回 NULL
		// IS NOT NULL 表示找到了匹配的书籍名称
		query = query.Where(
			"title LIKE ? OR description LIKE ? OR JSON_SEARCH(book_names, 'all', ?) IS NOT NULL",
			"%"+keyword+"%", // 标题包含关键词
			"%"+keyword+"%", // 描述包含关键词
			"%"+keyword+"%", // 书籍列表中有书名包含关键词
		)
	}

	// 5. 统计总数
	var total int64
	query.Count(&total)

	// 6. 分页查询（按发布时间倒序）
	var batches []models.Batch
	offset := (page - 1) * pageSize
	query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&batches)

	// 7. 返回结果
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

// GetMyBatches 获取我的发布
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
	database.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&batches)

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

	userID, _ := c.Get("userID")
	result := database.DB.Where("id = ? AND user_id = ?", id, userID).Delete(&models.Batch{})
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
