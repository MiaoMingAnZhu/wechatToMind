
昨天看到 [[wechatbot]]，可以在微信中和大模型对话，于是就萌生一个想法：

那么是否可以通过对话将消息保存到 Obsidian 中呢？本项目基于：
 - [openwechat](https://github.com/eatmoreapple/openwechat)
 - [wechatbot](https://github.com/869413421/wechatbot.git)

获取项目：
- `git clone https://github.com/zigholding/WechatObsidian.git`

进入项目目录：
- `cd WechatObsidian`

复制配置文件：
- `copy config.dev.json config.json`

修改配置文件：
- `vi config.json`
- `api_key`：[[deepseek]] API
- `obsidian_group`：保存 Obsidian 笔记的群聊名称
- `obsidian_daily_note`：日记笔记路径，会替换 `{year}{month}{day}` 为当日日期；

启动项目，扫码登录：
- `go run main.go`

项目支持 [[wechatbot]] 原生的功能（修复了 wechatbot 每次热加载时重复回复的bug）：
- 群聊@回复
- 私聊回复
- 自动通过回复

例如问它：你知道“知更鸟在屋顶”这个公众号吗？

![Pasted image 20250104135431.png](./assets/Pasted%20image%2020250104135431.png)

咳，Kimi 就知道：
![Pasted image 20250104135530.png](./assets/Pasted%20image%2020250104135530.png)

额外添加了功能：在 `obsidian_group` 里发消息时，会将消息自动同步到 `obsidian_daily_note` 的日志笔记中：

![Pasted image 20250104134326.png](./assets/Pasted%20image%2020250104134326.png)




顺便说下，deepseek 的API比OpenAI好用多了！我都找不到 OpenAI 在哪儿充值，是我傻还是它蠢？

注册 [[deepseek api]] 开放平台后，点击 https://platform.deepseek.com/top_up 就是充值页面：

![Pasted image 20250103202507.png](./assets/Pasted%20image%2020250103202507.png)

充值后点击 API keys，输入名称，创建后复制 key，粘贴到 `config.json` 对应的 `api_key` 字段中即可。

![Pasted image 20250103202550.png](./assets/Pasted%20image%2020250103202550.png)

送佛送到西，项目是用 `golang` 写的，需要到 [Go官网下载页面](https://go.dev/dl/)，下载 windows 安装包。
> https://go.dev/dl/

![Pasted image 20250103201728.png](./assets/Pasted%20image%2020250103201728.png)

点击安装包，按默认项安装，之后在 `git bash` 里执行 `go version`，能输出版本号信息说明安装成功。

`golang` 的源国内使用不为，还需要设置镜像源，同样在 `git bash` 里输入：

```bash
export GOPROXY=https://goproxy.cn
```

