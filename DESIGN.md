# bubble-prompt — 设计文档 v0.1

> 基于 Bubble Tea 生态的交互式命令补全框架，彻底解决 go-prompt 的终端渲染问题，并提供更优雅的补全体验。

---

## 1. 项目目标

- **稳定**：利用 Bubble Tea 的 MVU 架构，从根本上消除 go-prompt 的连续按键乱码问题
- **美观**：通过 lipgloss 实现媲美现代 IDE 的补全 UI（圆角弹窗、高亮匹配、类型标注）
- **可组合**：本身就是一个标准 Bubble，可无缝嵌入任何 Bubble Tea TUI 应用
- **易用**：API 设计对标 go-prompt，迁移成本低

---

## 2. 依赖生态

```
charmbracelet/bubbletea   — 主事件循环 / MVU 框架
charmbracelet/bubbles     — TextInput 基础组件（可能选择自实现以获得更多控制）
charmbracelet/lipgloss    — 所有样式、布局
charmbracelet/harmonica   — 补全弹窗出现的弹簧动画（可选）
sahilm/fuzzy              — 模糊匹配算法
```

---

## 3. 核心概念

### 3.1 Suggestion（补全项）

```go
type Suggestion struct {
    Text        string  // 补全文本（实际插入的内容）
    Display     string  // 显示文本（可与 Text 不同，默认同 Text）
    Description string  // 右侧说明文字
    Category    string  // 分组标签（可选，用于分组展示）
}
```

### 3.2 Completer（补全函数）

```go
// Completer 根据当前输入和光标位置，返回补全建议列表
// document 提供了对输入内容的便捷访问方法
type Completer func(document Document) []Suggestion
```

### 3.3 Executor（执行函数）

```go
// Executor 在用户按下 Enter 时被调用
type Executor func(input string) error
```

### 3.4 Document（输入文档）

```go
type Document struct {
    Text            string  // 完整输入内容
    CursorPosition  int     // 光标字节位置（光标左侧内容长度）
}

// 常用辅助方法
func (d Document) TextBeforeCursor() string
func (d Document) TextAfterCursor() string
func (d Document) CurrentWord() string          // 光标所在单词
func (d Document) CurrentLineBeforeCursor() string
```

---

## 4. 用户 API

### 4.1 最简用法

```go
package main

import (
    "fmt"
    prompt "github.com/yourname/bubble-prompt"
)

func main() {
    p := prompt.New(
        executor,
        completer,
        prompt.WithPrefix(">>> "),
    )
    p.Run()
}

func executor(input string) error {
    fmt.Println("执行:", input)
    return nil
}

func completer(d prompt.Document) []prompt.Suggestion {
    return prompt.FilterHasPrefix([]prompt.Suggestion{
        {Text: "get",    Description: "获取资源"},
        {Text: "set",    Description: "设置资源"},
        {Text: "delete", Description: "删除资源"},
    }, d.CurrentWord(), true)
}
```

### 4.2 嵌入到更大的 TUI 中

```go
// bubble-prompt 导出标准的 bubbletea Model 接口
// 可以直接组合到父 Model 中

type AppModel struct {
    prompt prompt.Model
    output []string
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    var cmd tea.Cmd
    m.prompt, cmd = m.prompt.Update(msg)
    
    // 监听补全确认事件
    if ev, ok := msg.(prompt.ExecuteMsg); ok {
        m.output = append(m.output, ev.Input)
    }
    return m, cmd
}
```

---

## 5. 配置选项（Option 模式）

```go
// 前缀
prompt.WithPrefix("$ ")
prompt.WithDynamicPrefix(func() string { return time.Now().Format("15:04") + " > " })

// 历史记录
prompt.WithHistory([]string{"prev command 1", "prev command 2"})
prompt.WithHistoryLimit(1000)

// 补全窗口
prompt.WithMaxSuggestions(10)          // 最多显示几条
prompt.WithCompletionStyle(style)       // lipgloss.Style，自定义弹窗样式
prompt.WithSelectedStyle(style)         // 选中项样式
prompt.WithDescriptionStyle(style)      // 描述文字样式
prompt.WithCompletionPosition(prompt.CompletionBelow) // Below / Above（自动适配终端底部）

// 键绑定
prompt.WithKeyMap(customKeyMap)

// 初始文字 / placeholder
prompt.WithPlaceholder("输入命令，Tab 补全...")
```

---

## 6. 补全 UI 设计

```
┌─────────────────────────────────────────┐
│ >>> get us█                             │  ← 输入行，█ 表示光标
│ ╭──────────────────────╮                │
│ │ > user          用户  │                │  ← 选中项高亮
│ │   username      用户名│                │
│ │   user-list     列表  │                │
│ ╰──────────────────────╯                │
└─────────────────────────────────────────┘
```

关键视觉特性：
- **弹窗位置**：默认在光标正下方，靠近终端底部时自动向上弹出
- **模糊匹配高亮**：匹配字符以不同颜色标出，例如 `us` 匹配 `**us**er`
- **左右分栏**：左侧补全文本，右侧描述（用 lipgloss `JoinHorizontal` 实现）
- **滚动条**：超出 `MaxSuggestions` 时显示迷你滚动条
- **动画**：弹窗出现时有轻微的弹簧动效（可选关闭）

---

## 7. 内部架构

### 7.1 Model 结构

```
Model
├── textInput    *textInput      // 输入行（自实现或基于 bubbles/textinput）
├── completion   *completionList // 补全弹窗
├── history      *history        // 历史记录
├── completer    Completer        // 用户提供
├── executor     Executor         // 用户提供
└── styles       Styles           // 所有 lipgloss 样式
```

### 7.2 消息流

```
按键 tea.KeyMsg
    │
    ▼
Update()
    ├── Tab / Shift+Tab → 切换补全选项
    ├── ↑ / ↓         → (无补全时) 历史导航 / (有补全时) 移动选项
    ├── Enter          → 确认输入，emit ExecuteMsg
    ├── Esc            → 关闭补全弹窗
    └── 其他字符        → 更新 textInput，重新调用 Completer，刷新弹窗
    │
    ▼
View()
    ├── 渲染输入行（前缀 + 输入内容 + 光标）
    └── 渲染补全弹窗（若有候选项）
```

### 7.3 关键：为什么不会乱码

每次 `View()` 都从当前 `Model` 完整重新计算整个字符串，Bubble Tea 的 renderer 对比 diff 后刷新。**没有任何增量 ANSI 拼接**，按键再快也只是多次 Update，每次状态完全一致。

---

## 8. 辅助工具函数

```go
// 过滤补全项（对标 go-prompt，方便迁移）
func FilterHasPrefix(suggestions []Suggestion, sub string, ignoreCase bool) []Suggestion
func FilterContains(suggestions []Suggestion, sub string, ignoreCase bool) []Suggestion
func FilterFuzzy(suggestions []Suggestion, pattern string) []Suggestion  // 使用 sahilm/fuzzy
```

---

## 9. 开发路线图

### v0.1 — MVP
- [ ] 基础 TextInput（光标移动、退格、删除词）
- [ ] 静态补全列表（Tab 触发）
- [ ] 历史记录（↑/↓ 导航）
- [ ] Prefix 支持
- [ ] 最基础的 lipgloss 样式弹窗

### v0.2 — 体验打磨
- [ ] 模糊匹配 + 匹配高亮
- [ ] 弹窗自动上下翻转（靠近底部时）
- [ ] 分组补全（按 Category 分隔）
- [ ] 描述预览侧边栏

### v0.3 — 高级功能
- [ ] 异步补全（Completer 返回 tea.Cmd）
- [ ] 语法高亮（输入行内）
- [ ] 多行输入模式
- [ ] 弹窗出现动画（harmonica）

### v0.4 — 生态整合
- [ ] 与 cobra/urfave-cli 的集成示例
- [ ] 内置 Shell 风格补全（文件路径、环境变量）
- [ ] 完整文档站

---

## 10. 与 go-prompt 的 API 对比

| go-prompt API | bubble-prompt 对应 |
|--------------|-------------------|
| `prompt.New(executor, completer, opts...)` | `prompt.New(executor, completer, opts...)` ✓ |
| `prompt.OptionPrefix("$ ")` | `prompt.WithPrefix("$ ")` |
| `prompt.OptionHistory([]string{...})` | `prompt.WithHistory([]string{...})` |
| `prompt.OptionMaxSuggestion(10)` | `prompt.WithMaxSuggestions(10)` |
| `prompt.FilterHasPrefix(...)` | `prompt.FilterHasPrefix(...)` ✓ 完全兼容 |
| `p.Run()` 独占模式 | `p.Run()` + 可选嵌入模式 |

迁移成本：大部分代码只需修改 import 路径和少量 Option 名称。

---

## 11. 目录结构（规划）

```
bubble-prompt/
├── prompt.go           // 对外 API：New()、Run()、Model
├── model.go            // bubbletea Model 实现（Init/Update/View）
├── textinput.go        // 内部输入行组件
├── completion.go       // 补全弹窗组件
├── history.go          // 历史记录
├── document.go         // Document 类型及辅助方法
├── filter.go           // FilterHasPrefix / FilterFuzzy 等
├── styles.go           // 默认样式，基于 lipgloss
├── keymap.go           // 键绑定定义
├── options.go          // 所有 WithXxx 选项
├── _examples/
│   ├── simple/         // 最简示例
│   ├── kubernetes/     // 模拟 kubectl 补全
│   └── embedded/       // 嵌入更大 TUI
└── DESIGN.md           // 本文档
```
