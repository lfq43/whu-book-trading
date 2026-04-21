@echo off
echo 停止 Docker...
docker-compose down

echo 停止所有进程...
taskkill /F /IM go.exe
taskkill /F /IM node.exe