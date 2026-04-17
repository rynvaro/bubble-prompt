Output demo.gif

Set Shell "bash"
Set FontSize 15
Set Width 680
Set Height 360
Set FontFamily "Hack"
Set Theme "Catppuccin Mocha"
Set Padding 20
Set PlaybackSpeed 0.8

# 静默设置 PATH、切换目录、清屏，画面不展示
Hide
Type "export PATH=/usr/local/go/bin:$PATH"
Enter
Type "cd /Users/emac/spx/rynvaro/bubble-prompt/examples/simple"
Enter
Type "clear"
Enter
Sleep 200ms
Show

# 画面从这里开始录制
Type "go run ."
Sleep 200ms
Enter
Sleep 1s

# 演示 1：输入 "g" 触发动词补全
Type "g"
Sleep 400ms
Tab
Sleep 600ms

# 选第一项 get（已高亮），再 Tab 确认
Tab
Sleep 400ms
Enter
Sleep 800ms

# 演示 2：输入 "get " 后 Tab 展开资源类型
Type "get "
Sleep 400ms
Tab
Sleep 600ms

# 用 Tab 导航到 pod
Tab
Sleep 400ms
Tab
Sleep 400ms

# 接受 pod，再 Tab 展开具体 pod 名称
Enter
Sleep 400ms
Tab
Sleep 600ms

# 接受第一个 pod 名
Enter
Sleep 300ms
Enter
Sleep 1s

# 退出
Type "exit"
Enter
Sleep 800ms
