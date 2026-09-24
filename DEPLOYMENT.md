# 静态门户部署

1. 部署源码与素材，从示例创建 `backend/config.yaml`。
2. 设置 `CAMIE_DB_DSN` 和发布 API 使用的 `CAMIE_STATIC_TOKEN`。
3. 在 `backend` 执行 `go build -trimpath -o bin/camie-static ./cmd/camie-static`。
4. 执行 `bin/camie-static generate --config config.yaml`，确认生成和链接校验成功。
5. Nginx 根目录指向生成的 `dist/camie-portal`，生产流量不经过 Node 或 Go。

Nginx 示例（域名、路径需替换）：

```nginx
server {
    listen 80;
    server_name portal.example.org;
    root /srv/camie-portal/dist/camie-portal;
    index index.html;
    location / {
        try_files $uri $uri/ =404;
    }
    location ~* \.(html|js|css)$ {
        add_header Cache-Control "no-cache";
    }
}
```

无需 SPA 的 `/index.html` 回退。上传媒体另配映射或独立域名。不要将项目源目录、`backend` 或配置目录作为网站根目录。

后台集成启动 `camie-static serve --config config.yaml`，默认只监听本机 9144。调用 `POST /api/static/site` 后，根据响应 `Location` 查询任务结果。

发布流程：临时目录生成 → 校验链接 → 当前目录移至 `.previous` → 新目录接替。失败时尝试恢复旧目录；成功后上一版保留至下一次发布。目录替换有短暂切换窗口，如果生产要求持续可读，可在运维层使用版本目录及 Nginx 符号链接切换。

异常退出可能遗留 `.lock`，确认没有生成进程运行后才能移除。回滚前停止发布任务，再将 `.previous` 切换为当前目录。

本次只进行了本地生成、自动化测试和静态预览，未执行数据库初始化、真实数据库连接或生产发布。
