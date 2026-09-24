# 新版静态门户预览

本目录对应 2026-09-24 提供的新版 UI 设计稿，是独立于旧版 Go 静态化工程的第一阶段静态前端成果。

## 预览

```powershell
npm run preview:new
```

浏览器访问 `http://127.0.0.1:4174/`。页面不依赖数据库、Go 服务或前端框架。

## 页面

| 设计稿 | 页面 |
| --- | --- |
| 首页 | `index.html` |
| 党建 / 党建内容页 | `pages/party.html` / `pages/party-detail.html` |
| 部委动态 / 详情 | `pages/ministry.html` / `pages/ministry-detail.html` |
| 关于协会 | `pages/about.html` |
| 会员中心 | `pages/member.html` |
| 交流培训 | `pages/training.html` |
| 科技标准 | `pages/standards.html` |
| 政策研究 | `pages/policy.html` |
| 专家委员会 | `pages/experts.html` |
| 新闻中心 | `pages/news.html` |
| 视频专区 / 视频详情 | `pages/videos.html` / `pages/video-detail.html` |

公共视觉位于 `css/common.css`，首页专属样式位于 `css/home.css`，渐进增强交互位于 `js/main.js`。设计切图复制并按用途重命名到 `assets/images/`。

页面主体内容均已生成进 HTML。`build-pages.mjs` 只用于维护阶段批量同步公共页头、页脚和演示内容；预览时无需运行它。

## 后续动态化区域

- 首页焦点图、通知公告、栏目标签内容、分支机构、专家内容及副会长单位。
- 各栏目侧栏、列表、日期、分页总数和分页链接。
- 文章标题、来源、日期、正文及上一篇/下一篇。
- 视频缩略图、标题、摘要、发布日期与媒体地址。

后续接入统一静态化平台时，应保留现有语义化容器与 class，仅把演示条目替换为模板循环和实际字段，不要改为浏览器端生成主体内容。
