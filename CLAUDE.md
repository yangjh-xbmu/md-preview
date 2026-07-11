# md-preview

## 目标

将原有 CLI 预览改造为 Wails 桌面应用，并使用 React + Tailwind 作为渲染界面，仍保持：

- Markdown 渲染与 `goldmark` 一致性
- HTML 安全过滤
- 文件变更自动刷新

## 关键实现

- `main.go`: CLI 参数解析、Wails 启动，以及内嵌前端资源与本地图片 Handler 的装配。
- `app.go`: `App` 结构体、`LoadMarkdown`、`CurrentVersion`、本地图片白名单、文件监听和事件发送（`markdown-updated`）。`exportHTMLTemplate` 内嵌 Mermaid CDN 与初始化脚本，导出 HTML 在浏览器打开时自动渲染 mermaid 代码块。
- `local_assets.go`: 解析 Markdown 相对图片路径，生成不暴露磁盘路径的资源 ID，只向 Wails WebView 提供当前文档渲染时登记的图片文件。
- `frontend/src/App.tsx`: GitHub 风格 Markdown 展示界面，订阅 `markdown-updated`，将 HTTP、HTTPS 和邮件链接交给系统默认应用。`contentHtml` 与 `theme` 联合 effect 调用 `renderMermaidBlocks`，先于 Prism 把 `pre > code.language-mermaid` 替换为 `<div class="md-mermaid">` 并渲染 SVG。
- `frontend/src/mermaid.ts`: Mermaid 渲染助手。封装 `mermaid.initialize`、`mermaid.render`、主题映射（light/sepia → default，dark → dark）和按块错误处理。源码保存在 `data-mermaid-source` 属性，主题切换时据此重渲染。
- `frontend/src/mermaid-shim.d.ts`: Mermaid 11 自带类型依赖 TS 5+ 语法，项目通过 `tsconfig.json` 的 `paths` 指向本地 shim，跳过 `node_modules/mermaid` 类型加载，避免拉入 `@types/d3-dispatch` 的语法错误。
- `frontend/src/style.css`: Tailwind 入口。
- `frontend/src/App.css`: 渲染内容细节样式。
- `wails.json`: 前端构建和前端文件服务配置。

## 验证命令

```bash
go test ./...
wails generate module
wails dev
wails build
npm --prefix frontend install
npm --prefix frontend run build
```

## 约束

- 仅提交必要文件，不要提交 `frontend/node_modules`。
- 桌面应用端口监听、HTML 输出和渲染策略保持最小复杂度，优先依赖 Wails 生命周期与事件机制。
- 新增功能或交互变更后，必须同步更新 `README.md` 中的功能特性列表和快捷键表格。
