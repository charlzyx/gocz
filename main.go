package main

import (
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"strings"

	"github.com/charlzyx/gocz/config"
	"github.com/charmbracelet/huh"
)

type CommitMessage struct {
	Type        string
	Scope       string
	Subject     string
	Description string
}

// getPackageScopes 获取 packages 目录下的所有子目录名作为 scope 选项
func getPackageScopes() []string {
	scopes := []string{}
	entries, err := os.ReadDir("packages")
	if err == nil { // 如果目录存在就读取
		for _, entry := range entries {
			if entry.IsDir() {
				scopes = append(scopes, entry.Name())
			}
		}
	}
	return scopes
}

// askQuestions 交互式询问用户填写 commit 信息
// presetType: 预设的 commit 类型
// presetMessage: 预设的 commit 信息
// 返回填写好的 CommitMessage 结构体和可能的错误
func askQuestions(config *config.Config, presetType, presetMessage string) (*CommitMessage, error) {
	var commitType, commitScope, commitSubject, commitDescription string

	commitTypeOptions := make([]huh.Option[string], 0, len(config.Types))
	for key, ct := range config.Types {
		commitTypeOptions = append(commitTypeOptions,
			huh.NewOption[string](ct.Title, key))
	}

	if presetType != "" {
		for _, option := range commitTypeOptions {
			if strings.HasPrefix(option.Value, presetType) {
				commitType = option.Value
				break
			}
		}
	}
	if presetMessage != "" {
		commitSubject = presetMessage
	}

	// 合并配置的 scopes 和 packages 目录的 scopes
	allScopes := append(getPackageScopes(), config.Scopes...)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("类型 Type:").
				Options(commitTypeOptions...).
				Height(len(commitTypeOptions)).
				Filtering(true).
				Value(&commitType),
			huh.NewInput().
				Title("范围 Scope:").
				Placeholder(func() string {
					if len(allScopes) > 0 {
						return strings.Join(allScopes[:int(math.Min(float64(3), float64(len(allScopes))))], ", ")
					}
					return "eg. api, cli"
				}()).
				CharLimit(50).
				Suggestions(allScopes).
				Value(&commitScope),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("概述 Summary:").
				Placeholder("Short description").
				CharLimit(70).
				Value(&commitSubject).
				Validate(func(s string) error {
					if s == "" {
						return errors.New("Required")
					}
					return nil
				}),
			huh.NewText().
				Title("详情 Details:").
				CharLimit(80).
				Placeholder("Detailed description").
				Value(&commitDescription).
				WithHeight(4),
		),
	)

	if err := form.Run(); err != nil {
		if err.Error() == "user aborted" {
			os.Exit(0)
		}
		return nil, err
	}

	return &CommitMessage{
		Type:        commitType,
		Scope:       commitScope,
		Subject:     commitSubject,
		Description: commitDescription,
	}, nil
}

// checkGitStatus 检查 git 仓库状态
// 如果有未暂存的更改或未跟踪的文件则返回错误
func checkGitStatus() error {
	// 使用 LANG=C 强制 git 输出为英文
	cmd := exec.Command("git", "status")
	cmd.Env = append(cmd.Env, "LANG=C") // 设置环境变量 LANG=C

	cmdLocal := exec.Command("git", "status")

	output, err := cmd.Output()
	outputLocal, _ := cmdLocal.Output()

	if err != nil {
		return fmt.Errorf("无法检查 Git 仓库状态: %v", err)
	}

	outputStr := string(output)

	// 首先检查是否存在未跟踪文件或未暂存的更改
	if strings.Contains(outputStr, "Untracked files:") || strings.Contains(outputStr, "Changes not staged for commit:") {
		// 只有在需要显示错误时才处理本地语言输出
		outputStrLocal := formatLocalGitStatus(outputLocal)
		return fmt.Errorf(outputStrLocal)
	}

	return nil
}

// formatLocalGitStatus 格式化本地语言的 git status 输出
func formatLocalGitStatus(outputLocal []byte) string {
	lines := strings.Split(string(outputLocal), "\n")
	filteredLines := make([]string, 0)
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		// Skip empty lines and lines starting with "(" or "（"
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "(") && !strings.HasPrefix(trimmedLine, "（") {
			filteredLines = append(filteredLines, line)
		}
	}

	if len(filteredLines) > 2 {
		filteredLines = filteredLines[2:]
	}
	return strings.Join(filteredLines, "\n")
}

// formatPreview 格式化 git commit 命令预览
// 将多行 commit 信息格式化为易读的形式
func formatPreview(cmd string) string {
	// 将命令按 -m 参数分割并格式化
	parts := strings.Split(cmd, " -m ")
	formatted := parts[0]
	for i := 1; i < len(parts); i++ {
		formatted += "\n    -m " + parts[i]
	}
	return formatted
}

// showGitStatusError 显示 Git 状态错误对话框
// 提供添加未暂存文件或退出程序的选项
func showGitStatusError(title, message string) {
	var confirmed bool

	err := huh.NewConfirm().
		Title(title).
		Description(message).
		Affirmative("帮我 add 啊").
		Negative("退下!").
		Value(&confirmed).
		// WithHeight(20).
		Run()

	if err != nil && err.Error() == "user aborted" {
		os.Exit(0)
	}

	if confirmed {
		// 执行 git add .
		if err := exec.Command("git", "add", ".").Run(); err != nil {
			showError("❌ 执行错误", fmt.Sprintf("执行 git add 失败: %v", err))
		}
	} else {
		os.Exit(1)
	}
}

// showError 显示错误信息对话框并退出程序
func showError(title, message string) {
	var confirmed bool
	huh.NewConfirm().
		Title(title).
		Description(message).
		Affirmative("朕知道了").
		Negative("退下吧").
		Value(&confirmed).
		Run()
	os.Exit(1)
}

func main() {
	// 检查 git 状态
	if err := checkGitStatus(); err != nil {
		showGitStatusError("❌ 怎么回事啊?!", err.Error())
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		showError("❌ 配置错误", fmt.Sprintf("加载配置失败: %v", err))
	}

	// 获取命令行参数
	var presetType, presetMessage string
	args := os.Args[1:]
	if len(args) > 0 {
		presetType = args[0]
	}
	if len(args) > 1 {
		presetMessage = args[1]
	}

	message, err := askQuestions(cfg, presetType, presetMessage)
	if err != nil {
		showError("❌ 输入错误", err.Error())
	}

	// 生成主消息
	mainMessage := message.Type
	if message.Scope != "" {
		mainMessage += fmt.Sprintf("(%s)", message.Scope)
	}

	emoji := ""
	if selectedType, ok := cfg.Types[message.Type]; ok && len(selectedType.Title) > 0 {
		emoji = string([]rune(selectedType.Title)[0]) + " "
	}

	mainMessage += fmt.Sprintf(": %s%s", emoji, message.Subject)

	// 构建完整的 git commit 命令
	var cmdParts []string
	cmdParts = append(cmdParts, "git", "commit", "-m", fmt.Sprintf(`"%s"`, mainMessage))

	// 处理长描述，按行分割并为每行添加 -m 参数
	if message.Description != "" {
		for _, line := range strings.Split(strings.TrimSpace(message.Description), "\n") {
			if trimmedLine := strings.TrimSpace(line); trimmedLine != "" {
				cmdParts = append(cmdParts, "-m", fmt.Sprintf(`"%s"`, trimmedLine))
			}
		}
	}

	cmd := strings.Join(cmdParts, " ")

	// 显示预览并确认
	var confirmed bool = true
	err = huh.NewConfirm().
		Title("即将执行:").
		Description(formatPreview(cmd)).
		Affirmative("执行").
		Negative("取消").
		Value(&confirmed).
		Run()

	if err != nil && err.Error() == "user aborted" {
		os.Exit(0)
	}

	if confirmed {
		// 执行 git commit 命令
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			showError("❌ 执行失败", fmt.Sprintf("执行 git commit 失败: %v", err))
		}
	} else {
		os.Exit(0)
	}
}
