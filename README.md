# 中国环保机械行业协会静态化门户

参考 `D:\WebstormProjects\miic-portal` 的架构：**HTML/CSS/Vanilla JavaScript + Go 静态化服务 + MySQL 内容库 + Nginx**。

Go 在发布时渲染完整 HTML。正文、列表、分页链接、统一页头和页脚都写入生成文件，主要内容不依赖浏览器 JavaScript。JavaScript 仅负责本地索引搜索、标签/焦点图切换、弹窗和旧地址兼容。

## 本地预览

无需数据库或前端依赖，使用 Go 1.21+：

```powershell
cd backend
go run ./cmd/camie-static preview-server --config config.example.yaml
```

访问 [本地首页](http://127.0.0.1:4173/index.html)。修改模板或样式后重新运行生成。

也可以在根目录使用兼容命令：

```powershell
npm run preview
npm run dev
```

`npm run dev` 只提供已生成的 `dist/camie-preview`，不生成内容。Node 不是生产运行依赖。

## 页面与 URL

| 页面 | 输出地址 |
| --- | --- |
| 首页 | `index.html` |
| 新闻列表 | `news.html` |
| 视频列表 | `videos.html` |
| 新闻详情兼容入口 | `detail.html?id=102` |
| 视频详情兼容入口 | `video-detail.html` |
| 正式详情 | `article/YYYY/MM/{id}.html` |
| 栏目分页 | `list/{column_id}/{page}.html` |
| 汇总分页 | `list/news/{page}.html`、`list/videos/{page}.html` |
| 本地搜索 | `search.html?q=关键词` |

全站共用首页的七项导航：**首页、党建专栏、关于协会、部委动态、会员中心、协会服务、行业资讯**。

之前的 `#/news/102`、`#/videos`、`#/news?q=...` 等地址自动跳转到新的静态页。主要栏目、分页和详情使用普通 HTML 链接。

## 目录

```text
backend/
  cmd/camie-static/       CLI、预览服务、发布 API 入口
  internal/config/       YAML 配置及输出路径校验
  internal/model/        内容模型（参考 MIIC）
  internal/repository/   MySQL 发布内容读取（参考 MIIC）
  internal/demo/         无数据库预览内容
  internal/generator/    模板渲染、资源复制、链接校验、目录发布
  internal/httpapi/      鉴权、异步整站发布与任务查询
  templates/             公共页头页脚、首页、列表、详情 Go 模板
  db/                    可选开发库结构
  config.example.yaml    无密码的配置示例
styles.css               原设计样式及响应式适配
script.js                渐进增强交互及旧链接兼容
assets/、slice/          现有设计素材
design/                  五张原设计稿
server.cjs               可选 Node 本地预览服务
dist/camie-preview/      演示输出（不提交）
dist/camie-portal/       正式输出（不提交）
```

根目录 `index.html` 是开发入口；生产部署根目录必须指向生成的站点目录。

## 正式生成

```powershell
cd backend
Copy-Item config.example.yaml config.yaml
$env:CAMIE_DB_DSN='user:password@tcp(127.0.0.1:3306)/camie_portal?charset=utf8mb4&parseTime=true'
go run ./cmd/camie-static generate --config config.yaml
```

数据库表沿用参考工程的 `page`、`column`、`article`、`article_column_publish`、`article_attachment`。仅发布 `site.page_name` 对应页面内的栏目与文章，不生成同库其他门户页面的内容。

配置与数据库密码不会复制到发布目录。配置路径相对配置文件解析；输出限制为本工程 `dist` 的子目录。整站先在临时目录完成渲染及校验，再替换正式目录，上一版保留在 `.previous`。生成失败保留原站点，重建自动清理下线文章、旧日期路径和过期分页。

当前实现整站发布，没有照搬 MIIC 的单页/单栏目/单文章增量 API。发布期间使用文件锁，异常进程退出可能留下锁文件，确认原进程已退出后才能手工清理。

## 发布 API

```powershell
$env:CAMIE_STATIC_TOKEN='至少24字符的随机令牌'
go run ./cmd/camie-static serve --config config.yaml
```

默认监听 `127.0.0.1:9144`：

- `POST /api/static/site`：整站异步发布，返回 202 和任务 ID。
- `GET /api/static/jobs/{id}`：查询任务结果。
- `GET /healthz`：健康检查。

发布和任务接口使用 `Authorization: Bearer ...`。同一时刻只接受一个任务，重复提交返回 409。最近 100 个任务摘要保存在进程内存中，重启不保留。

## 验证与边界

```powershell
npm run check
cd backend
go test ./...
go vet ./...
go run ./cmd/camie-static preview --config config.example.yaml
```

测试覆盖静态正文、统一导航、链接、分页、下线清理、失败保护、输出路径限制、富文本清洗、未来发布时间过滤、MySQL 发布条件及 API 鉴权和任务状态。

本次使用演示数据验证，尚未连接真实 MySQL。预览中的联系方式、图片和文字来自设计素材或演示数据；没有媒体源的视频展示封面及提示，会员功能仍待对接。媒体与附件不会下载，配置 `media_base_url` 或映射上传目录后使用。

详见 [内容录入说明](backend/DATA_ENTRY_GUIDE.md)、[部署说明](DEPLOYMENT.md)。
