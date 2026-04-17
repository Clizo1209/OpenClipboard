package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gorilla/websocket"
)

// 文件名地址格式：(ws)[host]{port} 或 (wss)[host]{port}，端口可省略
// 示例：opencb-client(wss)[myserver.com]{8443}.exe
//        opencb-client(ws)[myserver.com].exe
var addrRe = regexp.MustCompile(`\((wss?)\)\[([^\]]+)\](?:\{(\d+)\})?`)

// parseAddrFromFilename 从文件名中提取服务端 WebSocket 地址
// (ws)[host]{port}  → ws://host:port/ws
// (wss)[host]{port} → wss://host:port/ws
func parseAddrFromFilename(exePath string) string {
	name := exePath
	if idx := strings.LastIndexAny(exePath, `/\`); idx >= 0 {
		name = exePath[idx+1:]
	}
	m := addrRe.FindStringSubmatch(name)
	if m == nil {
		return ""
	}
	host := m[2]
	if m[3] != "" {
		host = host + ":" + m[3]
	}
	return m[1] + "://" + host + "/ws"
}

// resolveServerURL 从文件名提取地址，否则交互式询问用户
func resolveServerURL(exePath string) string {
	if url := parseAddrFromFilename(exePath); url != "" {
		fmt.Printf("[配置] 从文件名读取到地址: %s\n", url)
		return url
	}
	fmt.Print("请输入服务器地址（如 ws://localhost:8080）: ")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		url := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(url, "ws://") || strings.HasPrefix(url, "wss://") {
			return url + "/ws"
		}
		fmt.Print("地址须以 ws:// 或 wss:// 开头，请重新输入: ")
	}
	fmt.Fprintln(os.Stderr, "未输入有效地址，退出")
	os.Exit(1)
	return ""
}

// ─── WebSocket & 状态 ─────────────────────────────────────────────────────────

type Message struct {
	Type    string `json:"type"`
	Room    string `json:"room,omitempty"`
	Content string `json:"content,omitempty"`
	Users   int    `json:"users,omitempty"`
	Error   string `json:"error,omitempty"`
}

type roomState struct {
	mu      sync.RWMutex
	room    string
	content string
	users   int
}

func (s *roomState) setRoom(room string) {
	s.mu.Lock()
	s.room = room
	s.mu.Unlock()
}

func (s *roomState) setSync(content string, users int) {
	s.mu.Lock()
	s.content = content
	s.users = users
	s.mu.Unlock()
}

func (s *roomState) get() (room, content string, users int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.room, s.content, s.users
}

// autoMode 控制自动同步方向："" 关闭，"push" 推送，"pull" 接收，"both" 双向
var (
	st       roomState
	conn     *websocket.Conn
	connMu   sync.Mutex
	autoMode string
	autoMu   sync.RWMutex
)

func getAutoMode() string {
	autoMu.RLock()
	defer autoMu.RUnlock()
	return autoMode
}

func setAutoMode(m string) {
	autoMu.Lock()
	autoMode = m
	autoMu.Unlock()
}

// startAutoPush 每 500ms 轮询本地剪切板，内容变化时推送到房间
func startAutoPush() {
	last, _ := clipboard.ReadAll()
	for {
		time.Sleep(500 * time.Millisecond)
		m := getAutoMode()
		if m != "push" && m != "both" {
			continue
		}
		room, _, _ := st.get()
		if room == "" {
			continue
		}
		text, err := clipboard.ReadAll()
		if err != nil || text == last {
			continue
		}
		last = text
		if err := send(Message{Type: "update", Content: text}); err == nil {
			fmt.Printf("\n[自动推送] %d 字符 → 房间 %s\n> ", len([]rune(text)), room)
		}
	}
}

func send(msg Message) error {
	data, _ := json.Marshal(msg)
	connMu.Lock()
	defer connMu.Unlock()
	if conn == nil {
		return fmt.Errorf("未连接到服务器")
	}
	return conn.WriteMessage(websocket.TextMessage, data)
}

func readLoop() {
	for {
		connMu.Lock()
		c := conn
		connMu.Unlock()
		if c == nil {
			return
		}
		_, raw, err := c.ReadMessage()
		if err != nil {
			fmt.Println("\n[断开] 与服务器的连接已断开:", err)
			return
		}
		var msg Message
		if json.Unmarshal(raw, &msg) != nil {
			continue
		}
		switch msg.Type {
		case "sync":
			st.setSync(msg.Content, msg.Users)
			if msg.Room != "" {
				st.setRoom(msg.Room)
			}
			room, _, users := st.get()
			m := getAutoMode()
			if (m == "pull" || m == "both") && msg.Content != "" {
				if err := clipboard.WriteAll(msg.Content); err == nil {
					fmt.Printf("\n[自动拉取] %d 字符 ← 房间 %s\n> ", len([]rune(msg.Content)), room)
				} else {
					fmt.Printf("\n[同步] 房间 %s | 在线 %d 人 | 内容 %d 字符\n> ", room, users, len([]rune(msg.Content)))
				}
			} else {
				fmt.Printf("\n[同步] 房间 %s | 在线 %d 人 | 内容 %d 字符\n> ", room, users, len([]rune(msg.Content)))
			}
		case "users":
			st.mu.Lock()
			st.users = msg.Users
			room := st.room
			st.mu.Unlock()
			fmt.Printf("\n[通知] 房间 %s 人数变化: %d 人\n> ", room, msg.Users)
		case "error":
			fmt.Printf("\n[错误] %s\n> ", msg.Error)
		}
	}
}

func printHelp() {
	fmt.Print(`
命令列表:
  join <房间号>   加入指定房间（房间号自动转大写）
  push            将本机剪切板内容发送到房间
  pull            将房间内容写入本机剪切板
  show            显示当前房间内容
  status          显示当前房间和人数
  help            显示此帮助
  quit / exit     退出程序

自动同步:
  auto push           本机剪切板变化时自动推送到房间
  auto pull           房间内容更新时自动写入本机剪切板
  auto both           双向自动同步
  auto off            关闭自动同步

文件名携带服务器地址示例（首次启动后地址永久保存至旁路文件）:
  opencb-client(ws)[myserver.com]{8080}.exe
  opencb-client(wss)[myserver.com]{8443}.exe
  opencb-client(wss)[myserver.com].exe         （省略端口）
`)
}

func main() {
	exePath, err := os.Executable()
	if err != nil {
		exePath = os.Args[0]
	}

	fmt.Println("=== OpenClipboard 命令行客户端 ===")
	serverURL := resolveServerURL(exePath)
	fmt.Printf("服务器: %s\n", serverURL)
	fmt.Println("输入 help 查看命令列表")
	fmt.Println()

	c, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "连接服务器失败: %v\n", err)
		os.Exit(1)
	}
	connMu.Lock()
	conn = c
	connMu.Unlock()
	fmt.Println("[连接] 已连接到服务器")

	// 自动创建房间（空 Room 字段触发服务端分配）
	if err := send(Message{Type: "join"}); err != nil {
		fmt.Fprintf(os.Stderr, "创建房间失败: %v\n", err)
		os.Exit(1)
	}

	go readLoop()
	go startAutoPush()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\n正在退出...")
		connMu.Lock()
		if conn != nil {
			conn.Close()
		}
		connMu.Unlock()
		os.Exit(0)
	}()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("> ")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			fmt.Print("> ")
			continue
		}

		parts := strings.Fields(line)
		cmd := strings.ToLower(parts[0])

		switch cmd {
		case "help":
			printHelp()

		case "join":
			if len(parts) < 2 {
				fmt.Println("用法: join <房间号>")
				break
			}
			roomID := strings.ToUpper(parts[1])
			if err := send(Message{Type: "join", Room: roomID}); err != nil {
				fmt.Println("发送失败:", err)
			} else {
				fmt.Printf("正在加入房间 %s...\n", roomID)
			}

		case "push":
			room, _, _ := st.get()
			if room == "" {
				fmt.Println("尚未加入任何房间，请先 join <房间号>")
				break
			}
			text, err := clipboard.ReadAll()
			if err != nil {
				fmt.Println("读取剪切板失败:", err)
				break
			}
			if text == "" {
				fmt.Println("剪切板为空")
				break
			}
			if err := send(Message{Type: "update", Content: text}); err != nil {
				fmt.Println("发送失败:", err)
			} else {
				fmt.Printf("已推送 %d 字符到房间 %s\n", len([]rune(text)), room)
			}

		case "pull":
			_, content, _ := st.get()
			if content == "" {
				fmt.Println("房间内容为空")
				break
			}
			if err := clipboard.WriteAll(content); err != nil {
				fmt.Println("写入剪切板失败:", err)
			} else {
				fmt.Printf("已拉取 %d 字符到本机剪切板\n", len([]rune(content)))
			}

		case "show":
			room, content, users := st.get()
			if room == "" {
				fmt.Println("尚未加入任何房间")
				break
			}
			fmt.Printf("房间: %s | 在线: %d 人\n--- 内容 ---\n", room, users)
			if content == "" {
				fmt.Println("(空)")
			} else {
				fmt.Println(content)
			}
			fmt.Println("------------")

		case "status":
			room, content, users := st.get()
			if room == "" {
				fmt.Println("尚未加入任何房间")
			} else {
				fmt.Printf("房间: %s | 在线: %d 人 | 内容: %d 字符\n", room, users, len([]rune(content)))
			}

		case "auto":
			dir := ""
			if len(parts) >= 2 {
				dir = strings.ToLower(parts[1])
			}
			switch dir {
			case "push", "pull", "both":
				setAutoMode(dir)
				fmt.Printf("自动同步已开启: %s\n", dir)
			case "off", "stop", "":
				setAutoMode("")
				fmt.Println("自动同步已关闭")
			default:
				fmt.Println("用法: auto push | pull | both | off")
			}

		case "quit", "exit":
			fmt.Println("再见！")
			connMu.Lock()
			if conn != nil {
				conn.Close()
			}
			connMu.Unlock()
			os.Exit(0)

		default:
			fmt.Printf("未知命令: %q，输入 help 查看命令列表\n", cmd)
		}

		fmt.Print("> ")
	}
}
