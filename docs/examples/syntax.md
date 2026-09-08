# Markdown 语法示例

将本文件拖入 md-preview，查看公式、图表、脚注和代码的实际效果。

## 脚注

Markdown 可以把补充说明放到文末，让正文保持简洁。[^reading]

[^reading]: 脚注支持跳转和返回，适合讲义、研究记录与长文档。

```markdown
Markdown 可以把补充说明放到文末，让正文保持简洁。[^reading]

[^reading]: 脚注支持跳转和返回，适合讲义、研究记录与长文档。
```

## Wiki 链接

试试 [[Wiki-Demo]]，也可以使用别名 [[Wiki-Demo|打开演示页]]。

```markdown
[[Wiki-Demo]]
[[Wiki-Demo|打开演示页]]
[[docs/examples/preview|预览样例]]
```

链接会先查找当前目录，再查找 Obsidian Vault 或 Git 工作区根目录，最后按文件名或路径后缀搜索。使用 `Alt+←` 返回，使用 `Alt+→` 前进。

跨目录查找需要祖先目录中存在 `.obsidian` 或 `.git`。同名笔记建议写明路径。这里提供 Markdown 文件导航；PDF 阅读、反向链接索引和 Obsidian 插件不在支持范围内。

## Mermaid 图表

```mermaid
flowchart LR
    A[打开 Markdown] --> B[渲染图表与正文]
    B --> C[阅读]
    C --> D[导出 HTML 或打印]
```

在代码围栏后标注 `mermaid` 即可。图表会跟随主题切换；导出的 HTML 使用 CDN 加载 Mermaid，需要联网渲染图表。

## 数学公式

行内公式：$h \geq 2$。

$$
\sqrt{x^2+1} + \sum_{i=1}^{n} i = \frac{n(n+1)}{2}
$$

```markdown
行内公式：$h \geq 2$。

$$
\sqrt{x^2+1} + \sum_{i=1}^{n} i = \frac{n(n+1)}{2}
$$
```

## 代码、表格和任务列表

```json
{
  "course": "数据新闻",
  "students": [
    {"student_id": "S001", "name": "张三"},
    {"student_id": "S002", "name": "李四"}
  ]
}
```

| 功能 | 支持 |
| --- | --- |
| 表格、删除线 | GFM |
| 数学公式 | KaTeX |
| 图表 | Mermaid |

- [x] 打开本地文件
- [x] 阅读公式和代码
- [ ] 修改文件，观察自动刷新

返回[项目介绍](../../README.md)。
