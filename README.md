# md-preview

一个小型本地 Markdown 预览桌面应用，使用 Go + Wails + React + Tailwind 构建。

项目核心目标是：给本地文件提供稳定、干净且可重复的 Markdown 渲染体验，直接在独立窗口中查看，避免启动浏览器或暴露额外服务。

## 设计初衷

- **本地优先**：只处理本地 Markdown 文件，不引入远程预览服务。
- **安全**：渲染后通过白名单清洗，过滤不安全标签，避免脚本注入风险。
- **轻量路径**：CLI 只需要一个文件参数即可打开预览窗口，默认监听文件变更自动刷新。
- **桌面体验**：保留应用窗口的交互感，满足“记事本式”预览需求。

## 兼容与技术约束

- 仅支持 `.md` 与 `.markdown`。
- 文件变更时通过 Wails 事件推送给前端，不依赖浏览器轮询。
- 默认渲染使用 `goldmark` + `extension.GFM`，再经过 `bluemonday` 清洗。

## 安装与运行

### 直接运行源码

```bash
go run . <file.md>
```

### 发布版本方式

```bash
wails build
.\build\bin\md-preview.exe <file.md>
```

## 命令参数

```text
Usage: md-preview [--browser] [--watch=false] <file.md>
```

- `--watch=false`  
  关闭文件监听。对于自动化脚本或单次检查更友好。
- `--browser`  
  保留兼容参数，当前仍以桌面模式启动。

## 给 Agent 的调用建议

如果你在自动化流程里调用该工具，建议按如下约定使用：

- 入口始终是单文件路径，例如：
  - `md-preview notes.md`
  - `md-preview --watch=false notes.md`
- 只要解析路径合法并成功启动预览，进程会持续运行直到窗口关闭；若参数或文件异常则快速返回非零退出码并输出错误。
- 当需要“可重复行为”时，优先使用 `--watch=false`。
- 不需要关注前端开发环境端口、浏览器地址或本地 HTTP 服务。

### 常见错误

- `file does not exist`：文件路径不存在或权限不足。
- `expected a Markdown file, got directory`：传入的是目录而非文件。
- `unsupported file extension`：请使用 `.md` 或 `.markdown`。

## 开发

```bash
wails dev
```

前端依赖与构建：

```bash
cd frontend
npm install
npm run build
```

## 验证

```bash
go test ./...
wails generate module
wails build
```

## 许可

MIT 或 Apache 2.0（二选一可自行补充到发布说明）
