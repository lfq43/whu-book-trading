.PHONY: dev up down clean

# 配置
BACKEND_PORT := 8082
FRONTEND_DIR := frontend

dev: up
	@echo "🚀 启动后端服务..."
	@cd backend && go run main.go &
	@echo "🚀 启动前端服务..."
	@cd $(FRONTEND_DIR) && npm run dev
	@echo "✅ 所有服务已启动"
	@echo "📝 按 Ctrl+C 停止所有服务"

up:
	@echo "🐳 启动 Docker 容器..."
	@docker-compose up -d
	@echo "⏳ 等待 Docker 就绪..."
	@sleep 3

down:
	@echo "🛑 停止 Docker 容器..."
	@docker-compose down
	@echo "🛑 停止所有服务..."
	@-pkill -f "go run main.go" 2>/dev/null || true
	@-pkill -f "npm run dev" 2>/dev/null || true
	@echo "✅ 已停止所有服务"

clean: down
	@echo "🧹 清理 Docker 数据和缓存..."
	@docker-compose down -v
	@echo "✅ 清理完成"

# 单独启动某个服务
backend:
	@cd backend && go run main.go

frontend:
	@cd $(FRONTEND_DIR) && npm run dev

docker:
	@docker-compose up -d

# 查看服务状态
status:
	@echo "📊 Docker 容器状态:"
	@docker-compose ps
	@echo ""
	@echo "📊 后端进程:"
	@-ps aux | grep "go run main.go" | grep -v grep || echo "未运行"
	@echo ""
	@echo "📊 前端进程:"
	@-ps aux | grep "npm run dev" | grep -v grep || echo "未运行"

# 查看日志
logs:
	@docker-compose logs -f