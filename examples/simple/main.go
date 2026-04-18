// Package main demonstrates multi-level (kubectl-style) tab completion with
// cursor-following popup position.
//
// Level 1 – verb:          get / describe / delete / apply / logs
// Level 2 – resource type: pod / deployment / service / configmap / node
// Level 3 – resource name: live-style names per type
// Level 4 – flags:         --namespace / --output / --selector / ...
package main

import (
	"fmt"
	"log"
	"strings"

	prompt "github.com/rynvaro/bubble-prompt"
)

func main() {
	p := prompt.New(
		executor,
		completer,
		prompt.WithPrefix("kubectl> "),
		prompt.WithPlaceholder("输入命令，Tab 多级补全，Ctrl+C 退出"),
		// popup follows the cursor instead of staying fixed at the prefix
		prompt.WithCompletionPosition(prompt.CompletionAtCursor),
	)
	if err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func executor(input string) error {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil
	}
	args := strings.Fields(input)
	switch args[0] {
	case "exit", "quit":
		fmt.Println("Goodbye!")
		return prompt.ErrExit
	default:
		fmt.Printf("▶ kubectl %s\n", input)
	}
	return nil
}

// ---- suggestion tables ------------------------------------------------------

var verbs = []prompt.Suggestion{
	{Text: "get", Description: "显示一个或多个资源"},
	{Text: "describe", Description: "显示资源的详细信息"},
	{Text: "delete", Description: "删除资源"},
	{Text: "apply", Description: "通过文件应用配置"},
	{Text: "logs", Description: "打印 Pod 中容器的日志"},
	{Text: "exit", Description: "退出"},
}

var resources = []prompt.Suggestion{
	{Text: "pod", Display: "pod", Description: "运行中的容器组"},
	{Text: "deployment", Display: "deployment", Description: "声明式应用部署"},
	{Text: "service", Display: "service", Description: "网络服务暴露"},
	{Text: "configmap", Display: "configmap", Description: "配置数据"},
	{Text: "node", Display: "node", Description: "集群节点"},
	{Text: "namespace", Display: "namespace", Description: "资源隔离空间"},
}

// fakeNames pretends to be the list of live resources for each type.
var fakeNames = map[string][]prompt.Suggestion{
	"pod": {
		{Text: "nginx-7d8b9c-xk2pq", Description: "Running"},
		{Text: "redis-5f6d8b-mnjlp", Description: "Running"},
		{Text: "worker-abc12", Description: "Pending"},
	},
	"deployment": {
		{Text: "nginx", Description: "3/3 ready"},
		{Text: "redis", Description: "1/1 ready"},
	},
	"service": {
		{Text: "kubernetes", Description: "ClusterIP"},
		{Text: "nginx-svc", Description: "LoadBalancer"},
	},
	"configmap": {
		{Text: "kube-dns", Description: "system"},
		{Text: "app-config", Description: "default"},
	},
	"node": {
		{Text: "node-1", Description: "Ready"},
		{Text: "node-2", Description: "Ready"},
	},
	"namespace": {
		{Text: "default", Description: ""},
		{Text: "kube-system", Description: ""},
		{Text: "production", Description: ""},
	},
}

// flags available after a resource name for each verb.
var verbFlags = map[string][]prompt.Suggestion{
	"get": {
		{Text: "--namespace", Display: "--namespace", Description: "-n  指定命名空间"},
		{Text: "--output", Display: "--output", Description: "-o  输出格式 (json|yaml|wide)"},
		{Text: "--selector", Display: "--selector", Description: "-l  标签选择器"},
		{Text: "--all-namespaces", Display: "--all-namespaces", Description: "-A  所有命名空间"},
		{Text: "--watch", Display: "--watch", Description: "-w  持续监听变化"},
		{Text: "--show-labels", Display: "--show-labels", Description: "    显示所有标签"},
	},
	"describe": {
		{Text: "--namespace", Display: "--namespace", Description: "-n  指定命名空间"},
		{Text: "--selector", Display: "--selector", Description: "-l  标签选择器"},
	},
	"delete": {
		{Text: "--namespace", Display: "--namespace", Description: "-n  指定命名空间"},
		{Text: "--force", Display: "--force", Description: "    强制删除"},
		{Text: "--grace-period", Display: "--grace-period", Description: "    优雅退出等待秒数"},
		{Text: "--selector", Display: "--selector", Description: "-l  标签选择器"},
	},
	"logs": {
		{Text: "--namespace", Display: "--namespace", Description: "-n  指定命名空间"},
		{Text: "--follow", Display: "--follow", Description: "-f  持续输出"},
		{Text: "--tail", Display: "--tail", Description: "    输出最后 N 行"},
		{Text: "--previous", Display: "--previous", Description: "-p  上一个容器实例的日志"},
		{Text: "--container", Display: "--container", Description: "-c  指定容器名"},
	},
	"apply": {
		{Text: "--filename", Display: "--filename", Description: "-f  文件路径或 URL"},
		{Text: "--namespace", Display: "--namespace", Description: "-n  指定命名空间"},
		{Text: "--dry-run", Display: "--dry-run", Description: "    预演，不实际变更"},
		{Text: "--recursive", Display: "--recursive", Description: "-R  递归处理目录"},
	},
}

// ---- completer --------------------------------------------------------------

func completer(d prompt.Document) []prompt.Suggestion {
	before := d.TextBeforeCursor()
	args := strings.Fields(before)
	word := d.CurrentWord()
	atWordBoundary := strings.HasSuffix(before, " ")

	// Count how many non-flag args we've seen so far.
	// Flags (--xxx) can appear anywhere after level-3 and are always offered
	// when the current word starts with "-".
	nonFlagArgs := make([]string, 0, len(args))
	for _, a := range args {
		if !strings.HasPrefix(a, "-") {
			nonFlagArgs = append(nonFlagArgs, a)
		}
	}

	// If the user is typing a flag, offer flags for the current verb.
	if strings.HasPrefix(word, "-") && len(nonFlagArgs) >= 1 {
		if flags, ok := verbFlags[nonFlagArgs[0]]; ok {
			return prompt.FilterHasPrefix(flags, word, true)
		}
		return nil
	}
	// If we're at a word boundary and there are already ≥3 non-flag args,
	// offer flags as a next-step hint.
	if atWordBoundary && len(nonFlagArgs) >= 3 {
		if flags, ok := verbFlags[nonFlagArgs[0]]; ok {
			return flags
		}
		return nil
	}

	switch {
	case len(args) == 0:
		return prompt.FilterHasPrefix(verbs, word, true)

	case len(args) == 1 && !atWordBoundary:
		return prompt.FilterHasPrefix(verbs, args[0], true)

	case len(args) == 1 && atWordBoundary:
		if verbTakesResource(args[0]) {
			return resources
		}
		return nil

	case len(args) == 2 && !atWordBoundary:
		if verbTakesResource(args[0]) {
			return prompt.FilterHasPrefix(resources, args[1], true)
		}
		return nil

	case len(args) == 2 && atWordBoundary:
		if names, ok := fakeNames[args[1]]; ok {
			return names
		}
		return nil

	case len(args) == 3 && !atWordBoundary:
		if names, ok := fakeNames[args[1]]; ok {
			return prompt.FilterHasPrefix(names, args[2], true)
		}
		return nil
	}

	return nil
}

func verbTakesResource(verb string) bool {
	switch verb {
	case "get", "describe", "delete", "logs":
		return true
	}
	return false
}
