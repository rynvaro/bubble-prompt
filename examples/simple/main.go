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
		prompt.WithPrefix(">>> "),
		prompt.WithPlaceholder("输入命令，Tab 补全，Ctrl+C 退出"),
	)
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func executor(input string) error {
	input = strings.TrimSpace(input)
	switch input {
	case "exit", "quit":
		fmt.Println("Goodbye!")
		// In a real app you'd call os.Exit here.
	default:
		fmt.Printf("You entered: %q\n", input)
	}
	return nil
}

var commands = []prompt.Suggestion{
	{Text: "get", Description: "获取资源"},
	{Text: "set", Description: "设置资源"},
	{Text: "delete", Description: "删除资源"},
	{Text: "list", Description: "列出资源"},
	{Text: "describe", Description: "查看资源详情"},
	{Text: "apply", Description: "应用配置"},
	{Text: "exit", Description: "退出程序"},
	{Text: "help", Description: "显示帮助"},
}

func completer(d prompt.Document) []prompt.Suggestion {
	return prompt.FilterHasPrefix(commands, d.CurrentWord(), true)
}
