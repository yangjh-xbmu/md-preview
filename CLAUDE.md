# md-preview

## 目标

将原有 CLI 预览改造为 Wails 桌面应用，并使用 React + Tailwind 作为渲染界面，仍保持：

- Markdown 渲染与 `goldmark` 一致性
- HTML 安全过滤
- 文件变更自动刷新

## 关键实现

- `main.go`: CLI 参数解析与 Wails 启动。
- `app.go`: `App` 结构体、`LoadMarkdown`、`CurrentVersion`、文件监听和事件发送（`markdown-updated`）。
- `frontend/src/App.tsx`: GitHub 风格 Markdown 展示界面，订阅 `markdown-updated`。
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
