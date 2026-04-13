package models

//数据传输对象格式表

// RegisterRequest 注册请求
type RegisterRequest struct {
	Account  string `json:"account" binding:"required,min=3,max=50"`  // 必填，账号名，3-50字符
	Username string `json:"username" binding:"required,min=3,max=50"` // 必填，显示用户名，3-50字符
	Password string `json:"password" binding:"required,min=6"`        // 必填，至少6位
	Email    string `json:"email" binding:"omitempty,email"`          // 可选，但如果填了必须是邮箱格式
	// VerificationCode 用户收到的邮箱验证码（若填写邮箱则需提供）
	VerificationCode string `json:"verification_code" binding:"omitempty,len=6"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Account  string `json:"account" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string `json:"token"` // JWT令牌
	User  User   `json:"user"`  // 用户信息（不包含密码）
}

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`    // 状态码：0=成功，其他=错误
	Message string      `json:"message"` // 提示信息
	Data    interface{} `json:"data"`    // 实际数据
}

// BookCreateRequest 单本书创建请求
type BookCreateRequest struct {
	Title         string  `json:"title" binding:"required,min=1,max=200"`
	Author        string  `json:"author" binding:"max=100"`
	ISBN          string  `json:"isbn" binding:"max=20"`
	Price         float64 `json:"price" binding:"required,gt=0"`
	OriginalPrice float64 `json:"original_price" binding:"gte=0"`
	Condition     string  `json:"condition" binding:"oneof=全新 几乎全新 良好 有笔记 破损"`
	Description   string  `json:"description" binding:"max=5000"`
	Images        string  `json:"images"`
}

// BatchCreateRequest 批量创建请求
type BatchCreateRequest struct {
	Books []BookCreateRequest `json:"books" binding:"required,min=1,max=20"` // 一次最多20本
}

// BookListRequest 书籍列表查询参数
type BookListRequest struct {
	Keyword   string  `form:"keyword"`              // 搜索关键词（书名/作者）
	MinPrice  float64 `form:"min_price"`            // 最低价
	MaxPrice  float64 `form:"max_price"`            // 最高价
	Condition string  `form:"condition"`            // 新旧程度
	Status    string  `form:"status"`               // 状态
	Page      int     `form:"page,default=1"`       // 页码
	PageSize  int     `form:"page_size,default=12"` // 每页数量
}
