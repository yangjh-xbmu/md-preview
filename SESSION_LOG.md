# SESSION LOG

## 完成
- 2026-09-08 修复 Prism 运算符、实体、URL 与 CSS 字符串的半透明白底，消除 JSON 冒号背景方块；3 种主题与 4 种语言的浏览器回归通过，提交 d802bbc。
- 2026-09-08 前端构建、Windows Wails 构建和 Go 测试通过，并通过实际桌面窗口截图确认 README 正常渲染。
- 2026-09-08 发布脚本统一使用 Bash 注入版本号，推送 v0.1.8 并完成 Windows、macOS 双架构与 Linux 发布包，GitHub Release 设为最新版。
- 2026-09-08 按独立 Markdown 阅读器定位完善 GitHub 简介、10 个 Topics 和下载主页，重整 README 并加入真实截图、窗口自动刷新动图、Agent 指南和语法示例，提交 7165fa0。
- 2026-09-08 经用户确认加入 MIT LICENSE 并更新 README，GitHub 已识别 MIT；文档链接检查及 README CI 通过，提交 4f7551f。
- 2026-08-02 排查公式不渲染问题并确认非回归：git 全历史无 KaTeX/MathJax 痕迹，goldmark 从首个版本起只挂 GFM，宏哥此前所见公式渲染来自其他工具。
- 2026-08-02 实现 KaTeX 公式渲染：app.go 挂 goldmark-katex 扩展，$...$/$$...$$ 服务端渲染为 span+MathML+内联 SVG，bluemonday 精确白名单放行对应标签、class、em 度量样式和 SVG 属性。
- 2026-08-02 前端 main.tsx 引入 katex.min.css；导出 HTML 不带 KaTeX CSS，模板加 .katex-html{display:none} 回退浏览器原生 MathML。
- 2026-08-02 新增 TestLoadMarkdownRendersMath 与 TestExportHTMLMathFallsBackToMathML 两个测试，go test 全绿，wails build 成功并重建本地 exe。
- 2026-08-02 README 加数学公式示例章节，CLAUDE.md/AGENTS.md 同步更新；提交 7cfb209/1f4e46b 并推送，打 v0.1.6 标签，Release workflow 成功产出四平台产物。
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

## 发现
- 2026-09-08 Prism 默认主题会给 operator、entity、url 和 CSS string 添加半透明白底，定制代码块主题时需要同时覆盖这些 token 的背景。
- 2026-09-08 Windows 预览启动验证应检查窗口正文或截图，单独看到进程存在无法证明文档已显示；共享桌面录制可用 PrintWindow 捕获指定窗口，避免录入其他窗口内容。
- 2026-09-08 跨平台 GitHub Actions 使用 Bash 参数展开注入版本号时，需要显式指定 shell: bash，避免 Windows 默认 PowerShell 解释不同。
- 2026-08-02 goldmark-katex 用 modernc.org/quickjs 在 Go 侧执行 KaTeX，产物是 span 嵌套 + MathML + 内联 SVG（根号、伸缩括号），bluemonday 需 AllowStyles 放行 em 度量内联样式并精确白名单 MathML/SVG 属性，否则公式被消毒成空壳。
- 2026-08-02 不带 KaTeX CSS 的静态导出 HTML 可用 .katex-html{display:none} 隐藏视觉层，回退到浏览器原生 MathML 渲染，避免 MathML 与 KaTeX HTML 双重显示。
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

## 待办
无
