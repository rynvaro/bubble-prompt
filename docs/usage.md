# bubble-prompt 使用文档

## 目录

1. [安装](#1-安装)
2. [核心概念](#2-核心概念)
3. [快速入门](#3-快速入门)
4. [Completer 编写指南](#4-completer-编写指南)
5. [配置选项](#5-配置选项)
6. [自定义样式](#6-自定义样式)
7. [Document API](#7-document-api)
8. [嵌入到更大的 TUI 中](#8-嵌入到更大的-tui-中)
9. [键位参考](#9-键位参考)

---

## 1. 安装

```bash
go get bubble-prompt
```

要求 Go 1.21+。

---

## 2. 核心概念

### Executor

```go
type Executor func(input string) error
```

用户按下 `Enter` 后被调用，`input` 是用户输入的完整字符串（已去除首尾空格由你自己处理）。将命令输出直接写到 `os.Stdout`/`os.Stderr` 即可，TUI 在执行期间自动暂停，输出会干净地出现在终端滚动区域。

### Completer

```go
type Completer func(d Document) []Suggestion
```

每次按键后调用，返回当前应该展示的补全候选列表。列表为空时弹窗自动关闭。

### Suggestion

```go
type Suggestion struct {
    Text        string // 补全时实际插入的内容（必填）
    Display     string // 弹窗里显示的文字，为空时使用 Text
    Description string // 右侧说明列
    Category    string // 分组标签（v0.2 后生效）
}
```

### Document

传入 Completer 的只读快照，包含当前输入和光标位置，见 [§7 Document API](#7-document-api)。

---

## 3. 快速入门

```go
package main

import (
    "fmt"
    "log"
    "strings"

    prompt "bubble-prompt"
)

func main() {
    p := prompt.New(
        executor,
        completer,
        prompt.WithPrefix("$ "),
        prompt.WithPlaceholder("输入命令，Tab 补全，Ctrl+C 退出"),
    )
    if err := p.Run(); err != nil {
        log.Fatal(err)
    }
}

func executor(input string) error {
    switch strings.TrimSpace(input) {
    case "help":
        fmt.Println("可用命令: get, set, delete, list")
    case "exit":
        fmt.Println("再见！")
    default:
        fmt.Printf("未知命令: %q\n", input)
    }
    return nil
}

var commands = []prompt.Suggestion{
    {Text: "get",    Description: "获取资源"},
    {Text: "set",    Description: "设置资源"},
    {Text: "delete", Description: "删除资源"},
    {Text: "list",   Description: "列出资源"},
    {Text: "help",   Description: "显示帮助"},
    {Text: "exit",   Description: "退出程序"},
}

func completer(d prompt.Document) []prompt.Suggestion {
    return prompt.FilterHasPrefix(commands, d.CurrentWord(), true)
}
```

运行效果：

```
$ get█
╭──────────────────╮
│ > get    获取资源 │
╰──────────────────╯
```

---

## 4. Completer 编写指南

### 4.1 使用内置过滤函数

```go
// 前缀匹配（推荐，性能最好）
prompt.FilterHasPrefix(suggestions, word, ignoreCase)

// 包含匹配
prompt.FilterContains(suggestions, word, ignoreCase)
```

### 4.2 多级补全（子命令）

根据 `Document.TextBeforeCursor()` 判断已输入的词，按阶段返回不同的候选列表：

```go
func completer(d prompt.Document) []prompt.Suggestion {
    args := strings.Fields(d.TextBeforeCursor())
    word := d.CurrentWord()

    // 第一个词：补全顶层命令
    if len(args) == 0 || (len(args) == 1 && word != "") {
        return prompt.FilterHasPrefix(topLevelCmds, word, true)
    }

    // 第二个词：根据第一个词补全子命令
    switch args[0] {
    case "get":
        return prompt.FilterHasPrefix(getSubCmds, word, true)
    case "set":
        return prompt.FilterHasPrefix(setSubCmds, word, true)
    }
    return nil
}
```

### 4.3 Flag 补全（`--xxx` 参数）

Flag 和子命令在框架层面没有区别，都是普通的 `Suggestion`。在 Completer 里用 `strings.HasPrefix(word, "-")` 检测当前词是否以 `-` 开头来决定是否提供 flag 列表：

```go
var getFlags = []prompt.Suggestion{
    {Text: "--namespace", Display: "--namespace", Description: "-n  指定命名空间"},
    {Text: "--output",    Display: "--output",    Description: "-o  输出格式 (json|yaml|wide)"},
    {Text: "--watch",     Display: "--watch",     Description: "-w  持续监听变化"},
}

func completer(d prompt.Document) []prompt.Suggestion {
    word := d.CurrentWord()
    args := strings.Fields(d.TextBeforeCursor())

    // 当前正在输入 flag
    if strings.HasPrefix(word, "-") {
        return prompt.FilterHasPrefix(getFlags, word, true)
    }
    // 刚输完资源名，还没输下一个词，自动提示 flag
    atBoundary := strings.HasSuffix(d.TextBeforeCursor(), " ")
    if atBoundary && len(args) >= 3 {
        return getFlags
    }
    // ... 其余多级逻辑
    return nil
}
```

### 4.4 分隔符感知（不替换整行，只替换当前词）

Completer 接受当前词，`prompt.New` 内部的 `acceptCompletion` 会自动只替换光标左侧的当前词，不影响已输入的其他内容。

例如用户输入 `kubectl get po█`，按 Tab 补全 `pod`，结果是 `kubectl get pod`，前面的 `kubectl get ` 保持不变。

---

## 5. 配置选项

所有选项以 `With` 开头，传给 `prompt.New` 的可变参数：

```go
p := prompt.New(executor, completer,
    prompt.WithPrefix(">>> "),
    prompt.WithDynamicPrefix(func() string {
        // 每次渲染时调用，可以在前缀中显示动态信息
        return time.Now().Format("15:04") + " > "
    }),
    prompt.WithPlaceholder("输入命令..."),
    prompt.WithHistory([]string{"get pod", "set config"}),
    prompt.WithMaxSuggestions(10),
    prompt.WithStyles(myStyles),
    prompt.WithCompletionPosition(prompt.CompletionAtCursor),
)
```

| 选项 | 说明 | 默认值 |
|------|------|--------|
| `WithPrefix(s)` | 静态前缀字符串 | `">>> "` |
| `WithDynamicPrefix(fn)` | 动态前缀函数，每帧调用（会覆盖 `WithPrefix`） | — |
| `WithPlaceholder(s)` | 输入为空时显示的灰色提示文字 | `""` |
| `WithHistory(entries)` | 预填充历史记录 | `nil` |
| `WithMaxSuggestions(n)` | 弹窗最多显示几条（超出显示滚动条）| `8` |
| `WithStyles(s)` | 完整替换样式 | `DefaultStyles()` |
| `WithCompletionPosition(p)` | 弹窗水平位置，见下表 | `CompletionAtPrefix` |

### 5.1 弹窗位置

`CompletionPosition` 控制补全弹窗的水平对齐：

| 常量 | 效果 |
|------|------|
| `CompletionAtPrefix`（默认）| 弹窗固定对齐输入起点（prefix 末尾），位置不随输入变化 |
| `CompletionAtCursor` | 弹窗跟随光标，始终出现在当前正在输入的字符正下方 |

```go
// 跟随光标（推荐用于多级子命令/flags 场景）
prompt.WithCompletionPosition(prompt.CompletionAtCursor)

// 固定位置（默认，适合 prefix 较短的场景）
prompt.WithCompletionPosition(prompt.CompletionAtPrefix)
```

当输入行较长导致终端换行时，框架会自动对弹窗缩进取模（`% terminalWidth`），保证弹窗始终可见。

---

## 6. 自定义样式

`Styles` 结构体的每个字段都是一个 `lipgloss.Style`，可以按需覆盖：

```go
import "github.com/charmbracelet/lipgloss"

s := prompt.DefaultStyles()

// 修改前缀颜色
s.Prefix = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#FF5F87")).
    Bold(true)

// 修改弹窗边框颜色
s.CompletionBox = lipgloss.NewStyle().
    Border(lipgloss.RoundedBorder()).
    BorderForeground(lipgloss.Color("#00ADD8")).
    Padding(0, 1)

// 修改选中项背景色
s.SelectedItem = lipgloss.NewStyle().
    Background(lipgloss.Color("#003153")).
    Foreground(lipgloss.Color("#FFFFFF")).
    Bold(true)

p := prompt.New(executor, completer,
    prompt.WithStyles(s),
)
```

### Styles 字段一览

| 字段 | 应用位置 |
|------|---------|
| `Prefix` | 提示前缀（如 `>>> `）|
| `Placeholder` | 输入为空时的提示文字 |
| `CompletionBox` | 补全弹窗的外框（边框、内边距、背景）|
| `Item` | 弹窗中未选中的候选项 |
| `SelectedItem` | 弹窗中当前高亮的候选项 |
| `Description` | 候选项右侧说明列 |
| `Scrollbar` | 右下角的 `n/total` 滚动指示器 |

---

## 7. Document API

`Document` 在每次按键后传给 Completer，提供对当前输入的便捷访问：

```go
type Document struct {
    Text           string // 完整输入内容
    CursorPosition int    // 光标位置（rune 下标，光标左侧有多少个字符）
}
```

| 方法 | 说明 |
|------|------|
| `d.TextBeforeCursor()` | 光标左侧的文本 |
| `d.TextAfterCursor()` | 光标右侧的文本 |
| `d.CurrentWord()` | 光标紧邻左侧的单词（以空白字符分隔）|

**示例：**

```go
// 输入: "kubectl get po█d"（█ = 光标）
d.Text              // "kubectl get pod"
d.TextBeforeCursor() // "kubectl get po"
d.TextAfterCursor()  // "d"
d.CurrentWord()      // "po"   ← 光标左侧到最近空格之间的内容
```

---

## 8. 嵌入到更大的 TUI 中

`prompt.Model` 实现了标准的 `tea.Model` 接口，可以直接作为子组件嵌入：

```go
package main

import (
    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/lipgloss"
    prompt "bubble-prompt"
)

type AppModel struct {
    prompt  prompt.Model
    history []string
}

func (m AppModel) Init() tea.Cmd {
    return m.prompt.Init()
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // 检测提交事件（prompt 在内部 quit 之前会把 submitted=true）
    // 嵌入模式下需要在父 Model 里调用 prompt.Update，
    // 并在 prompt.View() 之外渲染其他内容。
    var cmd tea.Cmd
    newPrompt, cmd := m.prompt.Update(msg)
    m.prompt = newPrompt.(prompt.Model)
    return m, cmd
}

func (m AppModel) View() string {
    header := lipgloss.NewStyle().Bold(true).Render("My App")
    return header + "\n\n" + m.prompt.View()
}
```

> **注意：** 嵌入模式下没有内置的"提交后调用 executor"逻辑，需要在父 Model 的 `Update` 里自行检测并处理。`Prompt.Run()` 是独占 REPL 循环的便捷封装，嵌入时不需要调用它。

---

## 9. 键位参考

### 光标移动

| 按键 | 动作 |
|------|------|
| `←` / `→` | 左移 / 右移一个字符 |
| `Ctrl+←` / `Alt+B` | 左移一个单词 |
| `Ctrl+→` / `Alt+F` | 右移一个单词 |
| `Home` / `Ctrl+A` | 移到行首 |
| `End` / `Ctrl+E` | 移到行尾 |

### 编辑

| 按键 | 动作 |
|------|------|
| `Backspace` | 删除光标左侧字符 |
| `Delete` | 删除光标右侧字符 |
| `Ctrl+W` | 删除光标左侧的整个单词 |
| `Ctrl+U` | 删除光标到行首的所有内容 |
| `Ctrl+K` | 删除光标到行尾的所有内容 |

### 补全

| 按键 | 动作 |
|------|------|
| `Tab` | 触发补全弹窗 / 选择下一个候选项 |
| `Shift+Tab` | 选择上一个候选项 |
| `↑` / `↓` | 补全弹窗打开时：上下移动选项 |
| `Enter` | 补全弹窗打开时：接受当前选中项并关闭弹窗 |
| `Esc` | 关闭补全弹窗，保留当前输入 |

### 历史

| 按键 | 动作 |
|------|------|
| `↑` | 补全关闭时：加载上一条历史命令 |
| `↓` | 补全关闭时：加载下一条历史命令（或恢复正在编辑的内容）|

### 其他

| 按键 | 动作 |
|------|------|
| `Enter` | 补全关闭时：提交输入，调用 Executor |
| `Ctrl+C` / `Ctrl+D` | 退出 prompt |
