package handlers

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/869413421/wechatbot/config"
	"github.com/eatmoreapple/openwechat"
)

// HandleObsidianMessage 处理笔记群消息
func HandleObsidianMessage(msg *openwechat.Message, group openwechat.Group) int {
	// === 1. 检查分隔符暗号 ===
	if msg.IsText() {
		cfg := config.LoadConfig()
		if cfg.SeparatorKeyword != "" && strings.TrimSpace(msg.Content) == cfg.SeparatorKeyword {
			insertSeparator()
			return 1
		}
	}

	os.MkdirAll("data/Daily", 0755)
	os.MkdirAll("data/Files", 0755)

	noteFileName := fmt.Sprintf("data/Daily/%s.md", time.Now().Format("20060102"))

	sender, err := msg.Sender()
	if err != nil {
		sender = &openwechat.User{NickName: "未知"}
	}

	content := getMessageContent(msg)

	// === 2. 清洗文字消息中的 @机器人 前缀 ===
	if msg.IsText() && msg.IsAt() {
		self := msg.Owner()
		if self != nil {
			replaceText := "@" + self.NickName
			content = strings.TrimSpace(strings.ReplaceAll(content, replaceText, ""))
		}
	}

	timeStr := time.Now().Format("15:04")
	line := fmt.Sprintf("- [%s] %s: %s\n", timeStr, sender.NickName, content)

	f, err := os.OpenFile(noteFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("❌ 保存笔记失败：%v", err)
		return 0
	}
	defer f.Close()

	f.WriteString(line)
	log.Printf("✅ 笔记已保存：%s", noteFileName)
	return 1
}

// insertSeparator 插入分隔线
func insertSeparator() int {
	os.MkdirAll("data/Daily", 0755)
	noteFileName := fmt.Sprintf("data/Daily/%s.md", time.Now().Format("20060102"))

	f, err := os.OpenFile(noteFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("❌ 打开笔记文件失败：%v", err)
		return 0
	}
	defer f.Close()

	separatorLine := fmt.Sprintf("\n---\n\n> 📅 新主题开始于：%s\n\n", time.Now().Format("15:04"))
	
	_, err = f.WriteString(separatorLine)
	if err != nil {
		log.Printf("❌ 写入分隔线失败：%v", err)
		return 0
	}

	log.Printf("✅ [暗号] 已在笔记中插入分隔线：%s", noteFileName)
	return 1
}

// getMessageContent 分发处理
func getMessageContent(msg *openwechat.Message) string {
	switch msg.MsgType {
	case 1:
		return msg.Content
	case 3:
		return handleImage(msg)
	case 34:
		return handleVoice(msg)
	case 49:
		return parseAppMessage(msg)
	default:
		return fmt.Sprintf("[不支持的消息类型：%d]", msg.MsgType)
	}
}

func handleImage(msg *openwechat.Message) string {
	fileName := msg.FileName
	if fileName == "" || !strings.Contains(fileName, ".") {
		fileName = fmt.Sprintf("img_%s.jpg", time.Now().Format("20060102_150405"))
	} else {
		fileName = fmt.Sprintf("%s_%s", time.Now().Format("20060102_150405"), fileName)
	}

	savePath := fmt.Sprintf("data/Files/%s", fileName)

	err := msg.SaveFileToLocal(savePath)
	if err != nil {
		log.Printf("❌ 保存图片失败：%v", err)
		return "🖼️ [图片下载失败]"
	}

	log.Printf("✅ 图片已下载：%s", savePath)
	return fmt.Sprintf("![图片](%s)", savePath)
}

func handleVoice(msg *openwechat.Message) string {
	fileName := fmt.Sprintf("voice_%s.amr", time.Now().Format("20060102_150405"))
	savePath := fmt.Sprintf("data/Files/%s", fileName)

	err := msg.SaveFileToLocal(savePath)
	if err != nil {
		log.Printf("❌ 保存语音失败：%v", err)
		return "🎤 [语音下载失败]"
	}

	log.Printf("✅ 语音已下载：%s", savePath)
	return fmt.Sprintf("🎤 [语音](%s)", savePath)
}

func parseAppMessage(msg *openwechat.Message) string {
	xmlContent := msg.Content
	
	// === 正则提取基础字段 ===
	typeRe := regexp.MustCompile(`<type>(\d+)</type>`)
	typeMatches := typeRe.FindStringSubmatch(xmlContent)
	appType := 0
	if len(typeMatches) > 1 {
		appType, _ = strconv.Atoi(typeMatches[1])
	}

	titleRe := regexp.MustCompile(`<title><!\[CDATA\[(.*?)\]\]></title>|<title>(.*?)</title>`)
	titleMatches := titleRe.FindStringSubmatch(xmlContent)
	title := ""
	if len(titleMatches) > 1 {
		if titleMatches[1] != "" { title = titleMatches[1] } else { title = titleMatches[2] }
	}
	
	// 优化标题：如果是“版本不支持”，尝试从 finderFeed 获取
	if strings.Contains(title, "当前微信版本不支持") {
		nickRe := regexp.MustCompile(`<nickname><!\[CDATA\[(.*?)\]\]></nickname>`)
		descRe := regexp.MustCompile(`<desc><!\[CDATA\[(.*?)\]\]></desc>`)
		nick := nickRe.FindStringSubmatch(xmlContent)
		desc := descRe.FindStringSubmatch(xmlContent)
		if len(nick) > 1 && len(desc) > 1 {
			title = fmt.Sprintf("%s: %s", nick[1], desc[1])
		}
	}

	urlRe := regexp.MustCompile(`<url><!\[CDATA\[(.*?)\]\]></url>|<url>(.*?)</url>`)
	urlMatches := urlRe.FindStringSubmatch(xmlContent)
	url := ""
	if len(urlMatches) > 1 {
		if urlMatches[1] != "" { url = urlMatches[1] } else { url = urlMatches[2] }
	}

	fileExtRe := regexp.MustCompile(`<fileext>(.*?)</fileext>`)
	fileExtMatches := fileExtRe.FindStringSubmatch(xmlContent)
	fileExt := ""
	if len(fileExtMatches) > 1 { fileExt = fileExtMatches[1] }

	// 提取描述
	desRe := regexp.MustCompile(`<des><!\[CDATA\[(.*?)\]\]></des>|<des>(.*?)</des>`)
	desMatches := desRe.FindStringSubmatch(xmlContent)
	description := ""
	if len(desMatches) > 1 {
		if desMatches[1] != "" { description = desMatches[1] } else { description = desMatches[2] }
	}

	log.Printf("🏷️ [解析] Type: %d, Title: %s", appType, title)

	// === 业务逻辑 ===

	// [链接] Type=5 (公众号/元宝)
	if appType == 5 {
		if title == "" { title = "无标题链接" }
		if url == "" { url = "无链接地址" }
		
		if description != "" {
			// ✅ 新格式：公众号文章 [标题](链接) + 摘要
			return fmt.Sprintf("公众号文章 [%s](%s)\n\n> **💡 摘要：** %s", title, url, description)
		}
		return fmt.Sprintf("公众号文章 [%s](%s)", title, url)
	}

	// [文件] Type=6
	if appType == 6 {
		fileName := title
		if fileExt != "" && !strings.HasSuffix(strings.ToLower(fileName), "."+strings.ToLower(fileExt)) {
			fileName += "." + fileExt
		}
		if fileName == "" { fileName = "未知文件" }
		savePath := fmt.Sprintf("data/Files/%s_%s", time.Now().Format("20060102_150405"), fileName)
		
		if err := msg.SaveFileToLocal(savePath); err != nil {
			return fmt.Sprintf("📎 %s (下载失败:%v)", fileName, err)
		}
		return fmt.Sprintf("📎 [%s](%s)", fileName, savePath)
	}

	// [视频号] Type=51 或 包含 finderFeed
	if appType == 51 || strings.Contains(xmlContent, "<finderFeed>") {
		log.Printf("📹 识别到视频号消息")
		
		res := fmt.Sprintf("📹 **[视频号]** %s", title)
		if description != "" {
			res += fmt.Sprintf("\n\n> 💡 **摘要：** %s", description)
		}
		res += "\n\n> ⚠️ *注：视频流暂不支持下载*"
		
		return res
	}

	return fmt.Sprintf("📄 [卡片] %s", title)
}