package main

import (
	"fmt"
	"net/http"
	"pdb_proxy/conf"
	"pdb_proxy/pdb"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 打印启动信息
	printStartupInfo()

	app := fiber.New()
	app.Use(recover.New())

	// 添加根路径处理器，为浏览器访问提供友好界面
	app.Get("/", handleRootAccess)

	// 原有的pdb查询路由
	app.Get("/download/symbols/:pdbname/:pdbhash/:pdbname", pdb.PdbQuery)

	// 处理其他404情况，如果是浏览器访问则返回友好页面
	app.Use(handleNotFound)

	// 打印服务器启动完成信息
	fmt.Printf("🚀 服务器启动完成！访问 http://%s 获取配置说明\n", conf.ServerPort)
	fmt.Println("📝 按 Ctrl+C 停止服务器")
	fmt.Println(strings.Repeat("=", 60))

	app.Listen(conf.ServerPort)
}

// printStartupInfo 打印启动信息
func printStartupInfo() {
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("🔧 PDB Proxy Server - Microsoft 符号服务器代理")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("⏰ 启动时间: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("🌐 监听地址: %s\n", conf.ServerPort)
	fmt.Printf("📁 PDB 缓存目录: %s\n", conf.PdbDir)
	fmt.Printf("⏱️ PDB 缓存 TTL: %s\n", formatCacheTTL(conf.PdbCacheTTL))
	fmt.Printf("🔗 上游服务器: %s\n", conf.PdbServer)
	fmt.Println("📋 可用端点:")
	fmt.Println("   GET  /                                    - 配置说明页面")
	fmt.Println("   GET  /download/symbols/{name}/{hash}/{name} - PDB文件下载")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Println("�� 正在启动服务器...")
}

func formatCacheTTL(ttl time.Duration) string {
	if ttl == 0 {
		return "永久"
	}
	return ttl.String()
}

// handleRootAccess 处理根路径访问
func handleRootAccess(c *fiber.Ctx) error {
	// 检查是否为浏览器访问
	if isBrowserRequest(c) {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Status(http.StatusOK).SendString(getHelpHTML())
	} else {
		return c.Status(http.StatusNotFound).SendString("PDB Proxy Server - Use /download/symbols/[pdbname]/[pdbhash]/[pdbname] for PDB files")
	}
}

// handleNotFound 处理404情况
func handleNotFound(c *fiber.Ctx) error {
	// 检查是否为浏览器访问
	if isBrowserRequest(c) {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.Status(http.StatusNotFound).SendString(getHelpHTML())
	} else {
		return c.Status(http.StatusNotFound).SendString("pdb not exist")
	}
}

// isBrowserRequest 检查是否为浏览器请求
func isBrowserRequest(c *fiber.Ctx) bool {
	userAgent := c.Get("User-Agent")
	accept := c.Get("Accept")

	// 检查User-Agent中是否包含常见浏览器标识
	browserIdentifiers := []string{"Mozilla", "Chrome", "Safari", "Firefox", "Edge", "Opera"}
	for _, identifier := range browserIdentifiers {
		if strings.Contains(userAgent, identifier) {
			return true
		}
	}

	// 检查Accept头是否包含text/html
	if strings.Contains(accept, "text/html") {
		return true
	}

	return false
}

// getHelpHTML 返回帮助页面HTML
func getHelpHTML() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PDB Proxy</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; max-width: 700px; margin: 60px auto; padding: 0 20px; color: #333; line-height: 1.6; }
        h1 { font-size: 24px; border-bottom: 1px solid #eaeaea; padding-bottom: 20px; margin-bottom: 30px; font-weight: 600; }
        .card { background: #f8f9fa; border: 1px solid #e9ecef; border-radius: 6px; padding: 20px; margin-bottom: 30px; }
        .label { font-size: 14px; color: #666; margin-bottom: 8px; display: block; }
        .code-area { display: flex; background: #fff; border: 1px solid #ddd; border-radius: 4px; overflow: hidden; }
        #symbol-path { flex-grow: 1; padding: 12px; font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, Courier, monospace; font-size: 14px; color: #24292e; overflow-x: auto; white-space: nowrap; line-height: 20px; }
        .copy-btn { background: #f8f9fa; border: none; border-left: 1px solid #ddd; padding: 0 20px; cursor: pointer; color: #555; font-size: 14px; transition: all 0.2s; white-space: nowrap; }
        .copy-btn:hover { background: #e9ecef; color: #333; }
        .copy-btn:active { background: #dde0e3; }
        .help-text { font-size: 14px; color: #666; }
        ol { padding-left: 20px; margin: 0; }
        li { margin-bottom: 8px; }
        a { color: #0366d6; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <h1>PDB Proxy Server</h1>
    
    <div class="card">
        <span class="label">Windows 环境变量配置 (_NT_SYMBOL_PATH)</span>
        <div class="code-area">
            <div id="symbol-path">srv*C:\Symbols*<span class="current-server"></span>/download/symbols</div>
            <button class="copy-btn" onclick="copyToClipboard(this)">复制</button>
        </div>
    </div>

    <div class="help-text">
        <p><strong>设置步骤：</strong></p>
        <ol>
            <li>打开系统属性 → 高级 → 环境变量</li>
            <li>新建/修改用户变量 <code>_NT_SYMBOL_PATH</code></li>
            <li>填入上方地址</li>
            <li>保存生效</li>
        </ol>
        <p style="margin-top: 30px; border-top: 1px solid #eaeaea; padding-top: 20px;">
            项目地址: <a href="https://github.com/szdyg/pdb_proxy" target="_blank">https://github.com/szdyg/pdb_proxy</a>
        </p>
    </div>

    <script>
        document.addEventListener('DOMContentLoaded', () => {
            const server = window.location.protocol + '//' + window.location.host;
            document.querySelectorAll('.current-server').forEach(el => el.textContent = server);
        });
        
        function copyToClipboard(btn) {
            const text = document.getElementById('symbol-path').textContent;
            navigator.clipboard.writeText(text).then(() => {
                const originalText = btn.textContent;
                btn.textContent = '已复制';
                setTimeout(() => btn.textContent = originalText, 2000);
            }).catch(err => {
                const textArea = document.createElement('textarea');
                textArea.value = text;
                document.body.appendChild(textArea);
                textArea.select();
                document.execCommand('copy');
                document.body.removeChild(textArea);
                btn.textContent = '已复制';
                setTimeout(() => btn.textContent = '复制', 2000);
            });
        }
    </script>
</body>
</html>`
}
