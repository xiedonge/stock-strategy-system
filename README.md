# 股票策略选股系统（A股）

本项目是一个面向个人量化研究的 A 股选股系统样板，包含行情采集、本地持久化、策略管理、选股计算与回测可视化等核心能力。当前实现提供一个可运行的全栈骨架，并通过内置示例数据帮助你快速验证流程。

## 功能概览

- 日线与 30 分钟级别行情存储（SQLite）
- 策略管理（可扩展策略类型与参数）
- 选股计算（MA 交叉策略示例）
- 历史回测与权益曲线展示（ECharts）
- 前后端分离，适合后续扩展数据源与策略库

## 技术栈

- 前端：Vue 3 + Vite + ECharts + Pinia
- 后端：Go + Gin + GORM
- 数据库：SQLite（本地持久化）

## 项目结构

- `backend/` Go 服务端
- `frontend/` Vue 前端
- `stock-strategy-system.md` 需求说明

## 快速开始

### 一键启动（推荐）

```bash
./scripts/start.sh
```

脚本会检测依赖（Go、Node.js、npm）。如未安装，将尝试通过系统包管理器自动安装，可能需要 `sudo` 权限。
启动完成后会输出可访问的前端与后端 URL（含局域网地址）。

默认启动：

- 后端 `:8080`
- 前端 `:5173`

可选环境变量：

- `PORT`：后端端口
- `FRONTEND_PORT`：前端端口

### 一键卸载

```bash
./scripts/uninstall.sh
```

说明：会停止服务并清理 `backend/data/`、`frontend/node_modules/`、`frontend/dist/`、`.run/`。

### 后端

```bash
cd backend
# 初始化依赖
GO111MODULE=on go mod tidy

# 启动服务（默认端口 8080）
go run ./cmd/server
```

可选环境变量：

- `PORT`: 服务端口（默认 8080）
- `DB_PATH`: SQLite 文件路径（默认 `data/stock.db`）

### 前端

```bash
cd frontend
npm install
npm run dev
```

默认前端通过 `http://localhost:8080/api` 访问后端，如需调整可设置：

- `VITE_API_BASE`: 例如 `http://localhost:8080/api`

若前端通过服务器 IP 访问（如 `http://<服务器IP>:5173`），建议设置：

- `VITE_API_BASE`: `http://<服务器IP>:8080/api`

### 示例数据

- 在前端点击“导入示例数据”即可生成模拟行情并写入数据库。
- 也可通过接口调用：`POST /api/demo/seed`。

### AkShare 数据同步

使用 AkShare 获取 A 股行情（支持日线与 30 分钟级别）。第一次运行会自动创建 Python 虚拟环境并安装依赖。

```bash
./scripts/sync_akshare.sh --symbols 000001,600519 --mode all --start-date 20240101 --end-date 20241231 --min-start "2024-12-01 09:30:00" --min-end "2025-02-01 15:00:00" --period 30
```

参数说明：

- `--symbols`: 股票代码列表（逗号分隔）；不提供则按 `--limit` 自动取前 N 只
- `--mode`: `daily` / `minute` / `all`
- `--start-date` / `--end-date`: 日线区间（YYYYMMDD）
- `--min-start` / `--min-end`: 分钟线区间（YYYY-MM-DD HH:MM:SS）
- `--period`: 分钟线周期（默认 30）
- `--limit`: 未指定 symbols 时的默认数量（默认 50）

也可通过接口触发同步：

```bash
curl -X POST http://localhost:8080/api/sync/akshare \
  -H 'Content-Type: application/json' \
  -d '{"symbols":["000001","600519"],"mode":"all","start_date":"20240101","end_date":"20241231","min_start":"2024-12-01 09:30:00","min_end":"2025-02-01 15:00:00","period":"30"}'
```

## 核心接口

- `GET /api/health` 服务健康检查
- `POST /api/demo/seed` 生成示例行情
- `GET /api/stocks` 获取股票列表
- `GET /api/stocks/:code/klines?interval=1d&limit=200` 获取 K 线数据
- `GET /api/strategies` 策略列表
- `POST /api/strategies` 创建策略
- `POST /api/screen` 运行选股
- `POST /api/backtest` 运行回测
- `POST /api/sync/akshare` AkShare 行情同步

## 策略扩展建议

- 在 `backend/internal/strategy/` 添加新策略文件。
- 在 `backend/internal/services/analysis_service.go` 中注册策略执行逻辑。
- 前端可扩展策略参数表单以匹配新增策略。

## 备注

当前示例使用内置随机行情作为数据源，方便本地验证流程。若接入真实行情数据，可在后端新增采集任务，写入 `klines` 表即可复用现有选股与回测流程。
