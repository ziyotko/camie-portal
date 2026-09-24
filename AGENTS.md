# 项目约定

- 本工程是静态化门户，参考 `D:\WebstormProjects\miic-portal` 的 Go + MySQL + HTML 架构。
- 页面模板在 `backend/templates`，正式输出为 `dist/camie-portal`；演示输出为 `dist/camie-preview`。不要直接修改生成目录。
- 正文、列表和导航必须生成到 HTML 中，不改回依赖 JavaScript 渲染主体的哈希路由应用。
- 所有页面永远复用首页七项导航及相同顺序，维护公共 `layout.html.tmpl`。
- 视觉以本工程 `design/` 为准，复用 `assets/`、`slice/`，参考工程只用于架构和实现方式。
- 内容库仅发布配置页面范围内的已发布、已审核、未删除且到达发布时间的内容。无真实数据库配置时使用 preview 模式，不擅自复制其他项目连接信息。
- 修改模板、内容模型或生成流程后重新生成预览；按变更运行 Go 测试、静态检查及必要的桌面/手机浏览器验证。
