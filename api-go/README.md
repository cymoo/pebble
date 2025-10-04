# 使用说明

1. **初始化项目**
   ```bash
   go mod download
   cp .env.example .env
   ```

2. **运行数据库迁移**
   ```bash
   make migrate
   ```

3. **启动服务**
   ```bash
   make run
   ```