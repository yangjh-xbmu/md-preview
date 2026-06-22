# SESSION LOG

## 完成
- 2026-06-22 按 speckit 工作流（specify → plan → tasks → analyze → implement → PR）完成 Mermaid 渲染支持功能开发，产物在 specs/002-mermaid-support/。
- 2026-06-22 前端集成 mermaid 11：新增 frontend/src/mermaid.ts helper（封装 initialize/render/主题映射/按块错误隔离），App.tsx 加 [contentHtml, theme] 联合 effect 先于 Prism 替换 language-mermaid 块为 SVG，App.css 加 .md-mermaid 样式与 sepia 容器背景。
- 2026-06-22 app.go exportHTMLTemplate 加 Mermaid CDN 脚本与 DOMContentLoaded 初始化器，导出 HTML 在浏览器打开时自动渲染 mermaid 块，主题按导出主题注入 default/dark。
- 2026-06-22 新增 3 个 Go 测试（mermaid 块保留、导出脚本注入、主题映射），调整现有导出测试断言范围以区分用户 script 与模板自带 Mermaid script。
- 2026-06-22 README 加 Mermaid 示例章节，CLAUDE.md/AGENTS.md 更新集成点与 Domain Map。
- 2026-06-22 PR #1 squash 合并到 main，本地 feature branch 清理。
- 2026-06-22 用构建产物 md-preview.exe 打开 README.md 做 dogfood 预览，验证 Mermaid 渲染效果。
- 2026-06-22 完成 6 项 Mermaid smoke test 代码级验证：go test 全部通过（含 3 个 Mermaid 测试）、wails build 成功、前端构建成功。逐项确认 (1) Mermaid 块 SVG 渲染 (2) Go 代码块 Prism 高亮+复制按钮不受影响 (3) 主题切换重渲染 (4) sepia 暖色容器融合 (5) 导出 HTML 嵌入 Mermaid CDN 脚本 (6) 语法错误 in-page 占位符。
- 2026-06-19 修复 Markdown 文件以 UTF-8 BOM 开头时首行一级标题被当作普通段落的问题，并增加回归测试。
- 2026-06-19 初始化并提交 Spec Kit 与 speckit-superpowers-bridge 工作流，后续开发默认通过 spec、plan、tasks 和 handoff 执行。
- 2026-06-19 按 speckit bridge 流程将指定 SVG 转换为 md-preview 应用图标，更新 build/appicon.png 与 build/windows/icon.ico，并完成 Wails 构建验证。
- 2026-06-19 将 UTF-8 BOM 标题排查经验写入 Obsidian Inbox，并预览 README 确认当前文档展示合适。
- 2026-06-16 修复 Wails WebView2 中选择文本后无法粘贴的问题：添加 Ctrl+C 复制和 Ctrl+A 全选的 JS 层拦截，调用 ClipboardSetText 写入系统剪贴板
- 2026-06-16 添加 goldmark-wikilink 扩展，支持 [[页面名]]、[[文件.pdf]]、[[页面|别名]] 三种 wiki 链接语法渲染
- 2026-06-16 实现 wiki 链接点击跳转：前端拦截链接点击，后端 ResolveWikiLink 将 .html href 解码并查找同目录 .md 文件
- 2026-06-16 实现导航历史栈：Alt+← 返回、Alt+→ 前进，菜单添加 Back/Forward 按钮，状态栏提示快捷键
- 2026-06-16 创建 Wiki-Demo.md 演示文件，README wiki 链接指向真实文件，推送 v0.0.8 release tag
- 2026-06-06 修复 GitHub Actions Release workflow，经 7 次迭代使四个平台（Win/macOS Intel/macOS ARM/Linux）全部构建成功并生成 Draft Release
- 2026-06-06 清理失败标签和旧 Release（v1.0.0/v1.1.0/v0.0.1-v0.0.6），仅保留 v0.0.7
- 2026-06-06 左键选中正文文本自动复制到系统剪贴板，匹配 WezTerm 交互体验
- 2026-06-06 打印 PDF 时隐藏浮动菜单、TOC 目录、状态栏和面板圆角/边框/阴影，采用 @media print + JS .printing class 双方案兼容 WebView2
- 2026-06-06 窗口标题栏显示当前打开的文件名
- 2026-06-06 创建跨平台 GitHub Actions Release workflow（Windows/macOS Intel/Apple Silicon/Linux），tag push 自动构建并创建 Draft Release
- 2026-06-06 添加 CI README 同步检查 workflow，源码变更时 README 未更新则挂 warning
- 2026-06-06 更新 README.md 加入完整功能特性列表、快捷键表格和版本发布说明
- 2026-06-06 CLAUDE.md 加入功能变更后同步更新 README 的约束规则
- 2026-06-06 将 md-preview 从本地 Markdown 浏览器预览方案调整为 Wails 桌面应用方案，使用 Go + Wails + React + Tailwind 构建独立窗口预览。
- 2026-06-06 增加 GitHub 风格 Markdown 渲染、主题切换、目录导航、代码块高亮、复制按钮和行号等阅读功能。
- 2026-06-06 增加 HTML 导出、打印导出 PDF、文件选择、拖拽加载和运行中重新加载 Markdown 文件能力。
- 2026-06-06 修复 Wails 桌面包卡在静态启动页的问题，将 Vite 生产资源路径改为相对路径。
- 2026-06-06 修复 Prism 语言包加载顺序导致的 `class-name` 启动错误。

## 发现
- 2026-06-22 mermaid 11 传递依赖 @types/d3-dispatch 使用了 TS 5+ const 类型参数语法，TS 4.6 编译会报 TS1139。解法是升级 typescript 到 5.4 + 加 mermaid-shim.d.ts 走 tsconfig paths 绕开 node_modules 类型加载。
- 2026-06-22 Go fmt.Sprintf 模板里的 CSS 百分比（如 100%）必须转义成 100%%，否则 vet 报 '%; has unknown verb ;' 编译失败。
- 2026-06-22 speckit 工作流无 CLI，是 specs/<feature>/ 目录下的人工阶段流程，顺序为 specify → plan → tasks → analyze（checklists/）→ implement → PR，每个阶段有对应文件模板。
- 2026-06-22 Mermaid securityLevel: 'strict' 可阻断图表内的 HTML 标签与事件绑定，本地预览工具接受用户输入时建议双端（前端 helper + 导出 HTML 初始化器）都启用。
- 2026-06-22 Mermaid 无原生 sepia 主题，github-sepia 映射到 default 主题并叠加 CSS 暖色容器背景（rgba(234,213,167,0.35)）做视觉融合。
- 2026-06-19 goldmark 不会把带 UTF-8 BOM 前缀的首行 # 识别为 ATX 标题，渲染前应先去掉文件开头 BOM。
- 2026-06-19 Wails Windows 图标使用 build/windows/icon.ico，缺失时会从 build/appicon.png 生成，替换应用图标应同时维护这两个资产。
- 2026-06-19 本机全局 git ignore 会忽略 build/，需要用 git add -f 收纳 Wails 图标资产。
- 2026-06-19 specify extension add 的 --force 只表示覆盖已安装扩展，外部 URL 的非交互信任确认仍需用 yes y 管道输入。
- 2026-06-16 Wails v2 在 Windows 上默认设置 `AreBrowserAcceleratorKeysEnabled = false`，禁用 WebView2 内置的 Ctrl+C/V/X/A 等浏览器加速键，需在 JS 层手动拦截并调用 ClipboardSetText
- 2026-06-16 goldmark-wikilink（go.abhg.dev/goldmark/wikilink）默认将 [[Foo Bar]] 渲染为 `<a href="Foo%20Bar.html">Foo Bar</a>`，空格被 URL 编码，后端需反向解码再查找 .md 文件
- 2026-06-16 Wails WebView 中点击链接不会自动导航，需前端手动拦截 click 事件、调用后端方法加载目标文件
- 2026-06-06 GitHub Actions macOS Wails build 产物是 .app 包而非裸二进制，打包需用 tar czf md-preview.app
- 2026-06-06 GitHub Actions Windows runner 上 Wails build -o 生成的二进制可能不带 .exe 扩展名
- 2026-06-06 go install wails CLI 的正确路径是 github.com/wailsapp/wails/v2/cmd/wails@version，非 v2 裸包
- 2026-06-06 Ubuntu 24.04 移除了 libwebkit2gtk-4.0-dev，Wails v2 需用 ubuntu-22.04 runner
- 2026-06-06 GitHub Actions matrix 默认 fail-fast 会导致一个 job 失败即取消其余，跨平台构建需显式设 false
- 2026-06-06 WebView2 打印时 Vite 代码分割的 @media print CSS 可能不生效，需用 JS 在 window.print 前同步添加 print 类作为可靠方案
- 2026-06-06 Wails v2 Windows 不支持跨平台编译，需 GitHub Actions 分别用 windows/macos/ubuntu runner 构建三平台二进制
- 2026-06-06 Wails WindowSetTitle 会自动追加 "| appname" 后缀，只需传文件名即可
- 2026-06-06 Wails + Vite 桌面应用必须设置 `base: "./"`，否则生产构建的 `/assets/...` 绝对路径可能导致桌面端 JS 无法加载。
- 2026-06-06 Wails 桌面应用不能只验证浏览器或前端构建，必须执行 `wails build` 并启动真实 exe 检查资源加载和 WebView 模块执行。
- 2026-06-06 Prism 语言包存在隐式依赖，`cpp` 需要先加载 `clike` 和 `c`，`markdown` 需要先加载 `markup`。
- 2026-06-06 桌面应用应提供 HTML 静态兜底和 React 入口错误显示，便于区分窗口启动失败、资源加载失败和业务渲染失败。
- 2026-06-06 暴露给 Wails 前端绑定的 Go 结构体最好使用导出类型，生成的 TypeScript 类型更清晰稳定。
- 2026-06-06 Windows 原生菜单样式不易定制，轻量阅读器更适合使用不占文档流的浮动自定义菜单并保留快捷键。

## 待办

