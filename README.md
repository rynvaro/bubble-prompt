# bubble-prompt

基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 生态的交互式命令行补全框架。

> 彻底解决 go-prompt 的连续按键乱码问题，同时提供更现代、更优雅的补全 UI。

![demo](docs/demo.gif)

## 特性

- **稳定** — MVU 架构，每帧从 Model 完整重建 View，连续按键再快也不会乱码
- **美观** — lipgloss 圆角弹窗、左右双栏（补全文本 + 描述）、自适应终端宽度
- **可组合** — 本身就是一个标准 `tea.Model`，可无缝嵌入任何 Bubble Tea 应用
- **易迁移** — API 设计对标 go-prompt，大多数项目只需修改 import 路径

## 快速开始

```bash
go get bubble-prompt
```

```go
package main

import (
    "fmt"
    "log"

    prompt "bubble-prompt"
)

func main() {
    p := prompt.New(
        func(input string) error {
            fmt.Println("执行:", input)
            return nil
        },
        func(d prompt.Document) []prompt.Suggestion {
            return prompt.FilterHasPrefix([]prompt.Suggestion{
                {Text: "get",    Description: "获取资源"},
                {Text: "set",    Description: "设置资源"},
                {Text: "delete", Description: "删除资源"},
            }, d.CurrentWord(), true)
        },
        prompt.WithPrefix(">>> "),
    )
    if err := p.Run(); err != nil {
        log.Fatal(err)
    }
}
```

## 键位

| 按键 | 动作 |
|------|------|
| `Tab` | 触发补全 / 选择下一项 |
| `Shift+Tab` | 选择上一项 |
| `↑` / `↓` | 历史导航（补全关闭时）/ 补全列表导航（补全打开时）|
| `Enter` | 提交输入（补全打开时：接受当前选中项）|
| `Esc` | 关闭补全弹窗 |
| `←` / `→` | 移动光标 |
| `Ctrl+A` / `Home` | 移到行首 |
| `Ctrl+E` / `End` | 移到行尾 |
| `Ctrl+W` | 删除光标前的单词 |
| `Ctrl+U` | 删除到行首 |
| `Ctrl+K` | 删除到行尾 |
| `Ctrl+C` / `Ctrl+D` | 退出 |

## 路线图

- [x] v0.1 — 基础 REPL、Tab 补全、历史导航、lipgloss 弹窗
- [ ] v0.2 — 模糊匹配高亮、弹窗自动上下翻转、分组补全
- [ ] v0.3 — 异步补全、语法高亮、多行输入
- [ ] v0.4 — cobra 集成、文件路径补全

## 许可证

MIT
