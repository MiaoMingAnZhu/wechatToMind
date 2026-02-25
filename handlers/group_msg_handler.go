package handlers

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/869413421/wechatbot/gtp"
	"github.com/eatmoreapple/openwechat"
)

var _ MessageHandlerInterface = (*GroupMessageHandler)(nil)

// GroupMessageHandler 群消息处理
type GroupMessageHandler struct {
}

// handle 处理消息
func (g *GroupMessageHandler) handle(msg *openwechat.Message) error {
	// 1. 所有群消息都先保存到 Obsidian
	if msg.IsSendByGroup() {
		sender, err := msg.Sender()
		if err != nil {
			sender = &openwechat.User{NickName: "未知"}
		}
		group := openwechat.Group{sender}

		// 保存消息
		flag := HandleObsidianMessage(msg, group)

		// 2. 回复逻辑优化
		// 排除纯文字消息 (MsgType 1)，避免刷屏
		if flag != 0 && msg.MsgType != 1 {
			
			// 判断是否是视频号 (MsgType 49 且 内容包含 finderFeed)
			isFinder := false
			if msg.MsgType == 49 && strings.Contains(msg.Content, "<finderFeed>") {
				isFinder = true
			}

			if isFinder {
				// 视频号特殊回复
				_, err := msg.ReplyText("⚠️ 视频号已记录标题/摘要，视频流暂无法下载。")
				if err != nil {
					log.Printf("回复视频号通知失败：%v", err)
				}
			} else {
				// 其他类型 (图片/文件/链接) 成功回复
				_, err := msg.ReplyText("✅ 消息已保存到笔记")
				if err != nil {
					log.Printf("回复保存通知失败：%v", err)
				}
			}
		}
	}

	// 3. 只有文字消息才继续处理 (AI 回复逻辑)
	if msg.IsText() {
		return g.ReplyText(msg)
	}
	return nil
}

// NewGroupMessageHandler 创建群消息处理器
func NewGroupMessageHandler() MessageHandlerInterface {
	return &GroupMessageHandler{}
}

// ReplyText 发送文本消息到群 (包含 AI 回答及自动保存)
func (g *GroupMessageHandler) ReplyText(msg *openwechat.Message) error {
	sender, err := msg.Sender()
	if err != nil {
		return err
	}

	group := openwechat.Group{sender}

	timestamp := time.Unix(msg.CreateTime, 0)
	formattedTime := timestamp.Format("20060102 15:04:05")

	log.Printf("Received Group：【%v】；Msg: 【%v】；Time：【%s】", group.NickName, msg.Content, formattedTime)

	if time.Since(timestamp) > 10*time.Second {
		return nil
	}

	if !msg.IsAt() {
		return nil
	}

	self := msg.Owner()
	if self == nil {
		return errors.New("owner is nil")
	}
	replaceText := "@" + self.NickName

	requestText := strings.TrimSpace(strings.ReplaceAll(msg.Content, replaceText, ""))
	
	reply, err := gtp.Completions(requestText)
	if err != nil {
		log.Printf("gtp request error: %v \n", err)
		msg.ReplyText("机器人神了，我一会发现了就去修。")
		return err
	}
	if reply == "" {
		return nil
	}

	groupSender, err := msg.SenderInGroup()
	if err != nil {
		log.Printf("get sender in group error :%v \n", err)
		return err
	}

	reply = strings.TrimSpace(reply)
	reply = strings.Trim(reply, "\n")
	atText := "@" + groupSender.NickName
	replyText := atText + reply
	
	// 1. 发送回答到群里
	_, err = msg.ReplyText(replyText)
	if err != nil {
		log.Printf("response group error: %v \n", err)
		return err
	}

	// 2. 手动将 AI 的回答保存到笔记
	saveAIReplyToNote(group, replyText)

	return nil
}

// saveAIReplyToNote 辅助函数：专门保存 AI 回答到当天的笔记
func saveAIReplyToNote(group openwechat.Group, content string) {
	os.MkdirAll("data/Daily", 0755)
	noteFileName := fmt.Sprintf("data/Daily/%s.md", time.Now().Format("20060102"))

	timeStr := time.Now().Format("15:04")
	line := fmt.Sprintf("- [%s] 机器人: %s\n", timeStr, content)

	f, err := os.OpenFile(noteFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("❌ 保存 AI 回答失败：%v", err)
		return
	}
	defer f.Close()

	f.WriteString(line)
	log.Printf("✅ AI 回答已自动保存：%s", content[:min(30, len(content))]+"...")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}