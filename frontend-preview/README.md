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
| 站内搜索 | `pages/search.html?q=关键词` |
| 视频专区 / 视频详情 | `pages/videos.html` / `pages/video-detail.html` |

公共视觉位于 `css/common.css`，首页专属样式位于 `css/home.css`，渐进增强交互位于 `js/main.js`。设计切图复制并按用途重命名到 `assets/images/`。

页面主体内容均已生成进 HTML。`build-pages.mjs` 只用于维护阶段批量同步公共页头、页脚和演示内容；预览时无需运行它。

## 后续动态化区域

- 首页焦点图、通知公告、栏目标签内容、分支机构、专家内容及副会长单位。
- 各栏目侧栏、列表、日期、分页总数和分页链接。
- 文章标题、来源、日期、正文及上一篇/下一篇。
- 视频缩略图、标题、摘要、发布日期与媒体地址。

后续接入统一静态化平台时，应保留现有语义化容器与 class，仅把演示条目替换为模板循环和实际字段，不要改为浏览器端生成主体内容。

## 与后台管理平台的栏目映射

最终静态化以 `dia-platform` 的数据关系和 2026 年 9 月版《协会网站栏目分级列表》为准。最新完整层级及静态化编码统一维护在 `column-structure.json`：一个门户页面对应 `template`，页面内每个可运营内容块对应 `column`，展示条目来自该栏目发布的 `article`。首页焦点图、通知公告、活动标签、服务入口、专题、分支机构、专家和副会长单位均已通过 `data-column-code` 标记稳定栏目编码；其中焦点图是首页运营位，内容来源栏目为“新闻中心 / 热点关注”，并由每篇文章自己的详情链接生成整块可点击内容。

静态生成时只替换栏目与文章数据，不改变页面结构：

- `column.code` 对应 `data-column-code`，作为模板取数的稳定键，不依赖可修改的栏目名称。
- `column.display_type` 决定轮播、列表、图片、广告或链接等展示方式。
- `article_column_publish` 决定文章在哪个栏目发布及置顶、加粗、颜色等栏目内展示属性。
- `article.url` 有值时按后台外链规则处理，否则生成站内静态详情页链接；所有详情链接必须在生成阶段写入 HTML。
- `column-structure.json` 是页面栏目名称、层级与建议编码的单一清单；校验会拒绝未登记的栏目编码，并检查清单中的栏目是否已在生成页面中体现。
