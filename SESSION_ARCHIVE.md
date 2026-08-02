# SESSION ARCHIVE

## 2026-06

### 完成
- 2026-06-06 修复 Prism 语言包加载顺序导致的 `class-name` 启动错误。
- 2026-06-06 修复 Wails 桌面包卡在静态启动页的问题，将 Vite 生产资源路径改为相对路径。
- 2026-06-06 增加 HTML 导出、打印导出 PDF、文件选择、拖拽加载和运行中重新加载 Markdown 文件能力。
- 2026-06-06 增加 GitHub 风格 Markdown 渲染、主题切换、目录导航、代码块高亮、复制按钮和行号等阅读功能。
- 2026-06-06 将 md-preview 从本地 Markdown 浏览器预览方案调整为 Wails 桌面应用方案，使用 Go + Wails + React + Tailwind 构建独立窗口预览。
- 2026-06-06 CLAUDE.md 加入功能变更后同步更新 README 的约束规则
- 2026-06-06 撰写并推送 Obsidian 技术笔记《Wails 与 Vite 桌面应用空白页排查》到 MyNotes Inbox。
- 2026-06-06 将页面内固定工具区先迁移到 Wails 原生菜单，随后替换为更美观的右上角浮动自定义命令菜单。
- 2026-06-06 将 Go 绑定返回类型从 `previewPayload` 调整为导出的 `PreviewPayload`，并重新生成 Wails 前端绑定。
- 2026-06-06 增加 HTML 级静态兜底和 React 模块加载错误显示，避免桌面窗口纯白。

