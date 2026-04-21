@echo off
echo 启动 Docker...
docker-compose up -d

echo 启动后端...
start "Backend" cmd /c "go run main.go"

echo 启动前端...
cd frontend
start "Frontend" cmd /c "npm run dev"