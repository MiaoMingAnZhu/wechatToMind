package bootstrap

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/869413421/wechatbot/handlers"
	"github.com/eatmoreapple/openwechat"
)

func Run() {
	// 创建 Bot（桌面模式）
	bot := openwechat.DefaultBot(openwechat.Desktop)
   
    
	// 注册消息处理函数
	bot.MessageHandler = handlers.Handler

	// 自定义 UUIDCallback：保存 URL + 启动 HTTP 服务
	bot.UUIDCallback = func(uuid string) {
		log.Println("🔄 获取到登录 UUID")

		// 确保 data 目录存在
		os.MkdirAll("data", 0755)

		// 拼接登录 URL
		qrUrl := fmt.Sprintf("https://login.weixin.qq.com/qrcode/%s", uuid)

		// 保存到文件
		urlFile := "data/login_url.txt"
		os.WriteFile(urlFile, []byte(qrUrl), 0644)
		log.Printf("✅ 登录 URL 已保存至：%s", urlFile)

		// 启动 HTTP 服务显示 HTML 页面
		go startHTTPServer(qrUrl)
	}

	// 创建热存储对象（路径改为 data/storage.json）
	reloadStorage := openwechat.NewJsonFileHotReloadStorage("data/storage.json")

	// 尝试热登录
	err := bot.HotLogin(reloadStorage)
	if err != nil {
		log.Printf("⚠️ 热登录失败：%v", err)
		log.Println("🌐 请访问 http://localhost:8080 获取登录链接")

		// 普通登录（会触发 UUIDCallback）
		if err = bot.Login(); err != nil {
			log.Printf("❌ 登录失败：%v", err)
			return
		}
	} else {
		log.Println("✅ 热登录成功！")
		// 热登录成功，启动状态页
		go startStatusPage()
	}

	// 阻塞等待消息
	bot.Block()
}

// 启动 HTTP 服务显示登录页面
func startHTTPServer(qrUrl string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>微信机器人登录</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; 
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .container {
            background: white;
            padding: 40px;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            max-width: 500px;
            width: 90%%;
            text-align: center;
        }
        h1 { color: #07c160; margin-bottom: 20px; font-size: 24px; }
        .url-box {
            background: #f5f5f5;
            padding: 15px;
            border-radius: 8px;
            font-family: "Courier New", monospace;
            font-size: 14px;
            word-break: break-all;
            margin: 20px 0;
            border: 1px solid #ddd;
        }
        .btn {
            display: inline-block;
            background: #07c160;
            color: white;
            padding: 12px 30px;
            border-radius: 8px;
            text-decoration: none;
            font-weight: bold;
            margin: 10px 5px;
        }
        .btn:hover { background: #06ad56; }
        .btn-secondary { background: #667eea; }
        .btn-secondary:hover { background: #5a6fd6; }
        .tips {
            margin-top: 20px;
            padding: 15px;
            background: #fff3cd;
            border-radius: 8px;
            font-size: 13px;
            color: #856404;
            text-align: left;
        }
        .tips ol { margin-left: 20px; }
        .tips li { margin: 8px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>📱 微信机器人登录</h1>
        <p style="color: #666; margin: 20px 0;">复制下方链接到微信打开，或浏览器访问后扫码</p>
        
        <div class="url-box">%s</div>
        
        <a href="%s" class="btn" target="_blank">🔗 点击打开链接</a>
        <a href="javascript:void(0)" class="btn btn-secondary" onclick="copyUrl()">📋 复制链接</a>
        
        <div class="tips">
            <strong>使用说明：</strong>
            <ol>
                <li>点击"点击打开链接"或复制链接到浏览器</li>
                <li>微信会显示登录二维码</li>
                <li>用手机微信扫码并确认登录</li>
                <li>页面会自动检测登录状态</li>
            </ol>
            <p style="margin-top: 10px; color: #856404;">⚠️ 链接有效期约 5 分钟，过期请刷新页面</p>
        </div>
    </div>
    
    <script>
        function copyUrl() {
            const url = "%s";
            navigator.clipboard.writeText(url).then(() => {
                alert("✅ 链接已复制到剪贴板！");
            }).catch(() => {
                prompt("请手动复制链接：", url);
            });
        }
        
        // 每 5 秒检查一次登录状态
        setInterval(() => {
            fetch('/status').then(r => r.json()).then(d => {
                if (d.logged_in) {
                    document.querySelector('.container').innerHTML = 
                        '<h1>✅ 登录成功！</h1><p style="color: #666; margin: 20px 0;">机器人已启动，正在接收消息...</p>';
                    setTimeout(() => window.close(), 3000);
                }
            });
        }, 5000);
    </script>
</body>
</html>`, qrUrl, qrUrl, qrUrl)
		w.Write([]byte(html))
	})

	// 状态检查接口
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, err := os.Stat("data/storage.json")
		loggedIn := err == nil
		fmt.Fprintf(w, `{"logged_in": %v}`, loggedIn)
	})

	log.Println("🌐 HTTP 服务已启动：http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

// 热登录成功后的状态页
func startStatusPage() {
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`<h1>✅ 机器人已运行中</h1><p>热登录成功，无需扫码</p>`))
    })
    
    log.Println("🌐 状态页已启动：http://localhost:8080")
    http.ListenAndServe(":8080", mux)
}