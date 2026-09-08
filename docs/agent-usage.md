# Agent 调用指南

md-preview 接收一个本地 Markdown 文件路径，并持续显示其内容。它适合预览 Agent 生成的方案、技术说明和报告。

## 调用流程

1. 生成或更新本地 `.md` / `.markdown` 文件。
2. 首次生成时，用程序与文档的绝对路径启动预览。
3. 确认窗口和正文已显示。进程存在本身不能证明渲染完成。
4. 后续更新同一文件即可，已有窗口默认监听文件变更。

## 命令参数

```text
md-preview [--watch=false] <file.md>
```

默认启用文件监听，约每秒检查一次。`--watch=false` 用于固定文件版本的预览；它不会在加载后自动退出。无文件参数时启动空窗口，可以通过 Open File 选择文件。

历史参数 `--browser` 仍可接受，目前仍启动桌面窗口。

## Windows：让窗口持续运行

有 Node.js 的 Agent 环境可以使用独立子进程，避免终端退出时结束预览：

```javascript
const fs = require('node:fs');
const { spawn } = require('node:child_process');

const binary = 'C:/Tools/md-preview/md-preview.exe';
const document = 'D:/notes/plan.md';
fs.accessSync(binary);
fs.accessSync(document);

const preview = spawn(binary, [document], {
  detached: true,
  stdio: 'ignore',
});
preview.on('error', console.error);
preview.unref();
```

路径需要替换为当前机器的真实绝对路径。GUI 应保持可见，避免用隐藏窗口选项启动。

## macOS 与 Linux

macOS 使用应用包：

```bash
open /Applications/md-preview.app --args /absolute/path/plan.md
```

Linux 使用可执行文件：

```bash
nohup /absolute/path/md-preview /absolute/path/plan.md > /tmp/md-preview.log 2>&1 &
```

## 错误处理

多余的位置参数和未知选项会产生命令行错误。文件不存在、路径指向目录或扩展名不支持时，当前程序会向标准错误输出诊断信息，并打开空窗口供用户选择文件。自动化调用应先检查文件存在和扩展名，再启动程序。

这里的启动方式已在 Windows 桌面预览中验证；其他系统仍应检查实际窗口和渲染结果。

返回[项目介绍](../README.md)。
