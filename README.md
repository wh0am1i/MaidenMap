# MaidenMap

自托管的梅登海德网格码（Maidenhead Locator）反查地名服务，面向业余无线电（HAM）爱好者。

- **后端**：Go 1.26 + Gin，离线数据（GeoNames + Natural Earth）
- **前端**：React 19 + Vite + TypeScript + shadcn/ui，中英双语，深/浅色跟随系统
- **部署**：Docker Compose（api + web 两容器），前面可加主机 nginx 做 TLS 终止

## 快速开始

```bash
# 首次：拉取 GeoNames + Natural Earth 数据到 ./data（几分钟，约下载 500 MB）
docker compose --profile update run --rm update-data

# 启动栈
docker compose up -d

# 打开
open http://127.0.0.1:8081/        # 前端
curl http://127.0.0.1:8180/api/grid/JO65ab    # 直连 API

# 停止
docker compose down
```

本地端口：Web `127.0.0.1:8081`，API `127.0.0.1:8180`（默认 8080 被占用时可在 `docker-compose.yml` 调整）。

## HTTP API

所有端点返回 JSON，编码 UTF-8，仅支持 GET。所有地名字段均为双语对象 `{en, zh}`；若无中文翻译，`zh` 为空串（不是 null）。

### `GET /api/health`

健康检查。返回数据集规模和最后更新时间，前端 topbar 的 `33.5k cities` 徽标也读这个端点。

```bash
curl http://127.0.0.1:8180/api/health
```

```json
{
  "status": "ok",
  "cities_count": 33558,
  "countries_count": 242,
  "data_updated_at": "2026-04-20T07:21:52Z"
}
```

### `GET /api/grid/:code`

单个网格查询。`:code` 支持 4/6/8 位大小写混合输入（不区分大小写），例如 `JO65`、`JO65ab`、`JO65ab11`。

```bash
curl http://127.0.0.1:8180/api/grid/JO65ab
```

```json
{
  "grid": "JO65ab",
  "center": { "lat": 55.0625, "lon": 12.0417 },
  "country": {
    "code": "DK",
    "name": { "en": "Denmark", "zh": "丹麦" }
  },
  "admin1": { "en": "Zealand", "zh": "西兰大区" },
  "admin2": { "en": "Vordingborg Kommune", "zh": "" },
  "city":   { "en": "Vordingborg", "zh": "" }
}
```

**字段含义**：
- `grid` — 归一化后的网格码
- `center.lat` / `center.lon` — 网格中心经纬度（WGS84，精度 4 位小数）
- `country` — 通过点-多边形判断的国家；海洋或南极空缺时为 `null`
- `admin1` / `admin2` — 最近城市所在的一级/二级行政区（与 country 不一致时丢弃 admin，避免边界误判）
- `city` — 最近城市名（来自 GeoNames cities15000）

**海洋点示例**（country 为 null）：

```bash
curl http://127.0.0.1:8180/api/grid/AA00
```

```json
{
  "grid": "AA00",
  "center": { "lat": -89.5, "lon": -179 },
  "country": null,
  "admin1": { "en": "", "zh": "" },
  "admin2": { "en": "", "zh": "" },
  "city":   { "en": "McMurdo Station", "zh": "" }
}
```

**错误 — 格式非法（HTTP 400）**：

```bash
curl -i http://127.0.0.1:8180/api/grid/BAD
```

```
HTTP/1.1 400 Bad Request
Content-Type: application/json

{"error":"invalid_grid","message":"invalid length 3: must be 4, 6, or 8"}
```

### `GET /api/grid?codes=A,B,C`

批量查询。最多 100 个代号，逗号分隔。成功项和错误项混合返回在同一数组里，靠 `error` 字段区分。

```bash
curl "http://127.0.0.1:8180/api/grid?codes=JO65ab,OM89,BAD"
```

```json
{
  "results": [
    {
      "grid": "JO65ab",
      "center": { "lat": 55.0625, "lon": 12.0417 },
      "country": { "code": "DK", "name": { "en": "Denmark", "zh": "丹麦" } },
      "admin1": { "en": "Zealand", "zh": "西兰大区" },
      "admin2": { "en": "Vordingborg Kommune", "zh": "" },
      "city":   { "en": "Vordingborg", "zh": "" }
    },
    {
      "grid": "OM89",
      "center": { "lat": 39.5, "lon": 117 },
      "country": { "code": "CN", "name": { "en": "People's Republic of China", "zh": "中华人民共和国" } },
      "admin1": { "en": "Tianjin", "zh": "天津市" },
      "admin2": { "en": "Tianjin Municipality", "zh": "天津市" },
      "city":   { "en": "Yangcun", "zh": "" }
    },
    {
      "grid": "BAD",
      "error": "invalid_grid",
      "message": "invalid length 3: must be 4, 6, or 8"
    }
  ]
}
```

**批量错误码**：
- `HTTP 400 missing_codes` — 未提供 `codes` 参数
- `HTTP 400 too_many_codes` — 超过 100 个

### 限流

每 IP 每分钟 60 次请求。超出返回 `HTTP 429`：

```json
{ "error": "rate_limited", "message": "too many requests" }
```

前端会 toast 提示并禁用输入 5 秒。

### 部署在主机 nginx 后

容器只监听 `127.0.0.1`，需要主机 nginx 反代。下面是一份生产可用的最小配置，包含 TLS、HSTS、速率限制、超时和安全响应头：

```nginx
# 定义一个给 /api 用的限流区（另一层防线，容器内 app 层还有 60 req/min/IP）
limit_req_zone $binary_remote_addr zone=maidenmap_api:10m rate=120r/m;

server_tokens off;

server {
    listen 80;
    server_name maidenmap.example.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name maidenmap.example.com;

    ssl_certificate     /etc/letsencrypt/live/maidenmap.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/maidenmap.example.com/privkey.pem;
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;
    ssl_prefer_server_ciphers on;

    # 安全响应头
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Content-Type-Options    "nosniff"   always;
    add_header X-Frame-Options           "DENY"      always;
    add_header Referrer-Policy           "strict-origin-when-cross-origin" always;

    client_max_body_size 64k;          # 防大 body 攻击；API 请求远小于此
    proxy_read_timeout   30s;
    proxy_connect_timeout 5s;

    location /api/ {
        limit_req zone=maidenmap_api burst=30 nodelay;
        proxy_pass         http://127.0.0.1:8180;
        proxy_set_header   Host              $host;
        proxy_set_header   X-Real-IP         $remote_addr;
        proxy_set_header   X-Forwarded-For   $proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto $scheme;
    }

    location / {
        proxy_pass http://127.0.0.1:8081;
        proxy_set_header Host $host;
    }
}
```

**`--trusted-proxies` 必须配对**：后端只在直连 peer 属于这个 CIDR 时才信 `X-Forwarded-For`；否则会 fallback 到直连 IP（防止攻击者直接打 API 并伪造 X-F-F 绕限流）。默认 `172.16.0.0/12` 覆盖 Docker 默认桥接网络；如果用 `network_mode: host` 或者主机 nginx 跑在容器外不同网段，要把对应地址加进去，比如 `--trusted-proxies=172.16.0.0/12,127.0.0.1/32`。

**Web 容器自己也带了一套安全头**（CSP / X-Frame-Options / Referrer-Policy），主机 nginx 的 header 会被它的响应覆盖；两层同向设置是深度防御的正常做法，不冲突。

## 数据更新

```bash
# 代码有更新的情况下，先重建镜像
git pull
docker compose build

# 重新拉取最新 GeoNames + Natural Earth 数据
docker compose --profile update run --rm update-data

# 数据写入 ./data/（原子写，服务不用重启，但要重启才会重新加载）
docker compose restart api
```

`update-data` 服务复用 `api` 的镜像（同一个 Dockerfile 和二进制，只是换了 entrypoint），所以 `docker compose build` 一次就够。如果只改数据、没改代码，`build` 那步可以省掉。

数据源：
- GeoNames cities15000、admin1CodesASCII、admin2Codes、alternateNamesV2（中文名）—— CC-BY
- Natural Earth `ne_10m_admin_0_countries.geojson` —— Public Domain（含 HK/MO/TW 独立多边形）

## 本地开发

```bash
# 终端 1：后端
cd api
go run ./cmd/server --data-dir=../data

# 终端 2：前端
cd web
npm install
npm run dev       # http://localhost:5173，/api 走 vite proxy 到 :8080
```

前端 Vite dev server 默认把 `/api/*` 代理到 `http://127.0.0.1:8080`，如果本地后端改了端口，改 `web/vite.config.ts`。

### 测试

```bash
# 后端
cd api && go test ./...

# 前端
cd web && npm run test
```

## 许可

代码以 [MIT License](LICENSE) 发布。

数据层保持各自原有许可：
- GeoNames — [CC-BY 4.0](https://creativecommons.org/licenses/by/4.0/)
- Natural Earth — Public Domain
