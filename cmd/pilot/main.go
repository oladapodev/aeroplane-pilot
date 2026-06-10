package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/oladapodev/aeroplane-pilot/internal/aeroplane"
	"github.com/oladapodev/aeroplane-pilot/internal/llm"
)

// ============ TELEGRAM TYPES ============

type Update struct {
	UpdateID      int            `json:"update_id"`
	Message       *Message       `json:"message"`
	CallbackQuery *CallbackQuery `json:"callback_query"`
}

type Message struct {
	MessageID int             `json:"message_id"`
	Chat      Chat             `json:"chat"`
	Text      string           `json:"text"`
	Entities  []MessageEntity  `json:"entities"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type MessageEntity struct {
	Type   string `json:"type"`
	Offset int    `json:"offset"`
	Length int    `json:"length"`
}

type CallbackQuery struct {
	ID      string `json:"id"`
	Message *Message `json:"message"`
	Data    string `json:"data"`
}

// ============ SEND METHODS ============

func sendMessage(token string, chatID int64, text string, replyMarkup *InlineKeyboardMarkup) error {
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	return postJSON(token, "sendMessage", payload)
}

func editMessageText(token string, chatID int64, messageID int, text string, replyMarkup *InlineKeyboardMarkup) error {
	payload := map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
		"text":       text,
		"parse_mode": "HTML",
	}
	if replyMarkup != nil {
		payload["reply_markup"] = replyMarkup
	}
	return postJSON(token, "editMessageText", payload)
}

func answerCallbackQuery(token, queryID, text string) error {
	payload := map[string]any{
		"callback_query_id": queryID,
		"text":              text,
		"show_alert":        false,
	}
	return postJSON(token, "answerCallbackQuery", payload)
}

type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data,omitempty"`
	URL          string `json:"url,omitempty"`
}

// ============ GET METHODS ============

func getUpdates(token string, offset, timeout int) ([]Update, error) {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=%d", token, offset, timeout)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var result struct {
		Result []Update `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Result, nil
}

func setCommands(token string) error {
	commands := []map[string]string{
		{"command": "start", "description": "Start the bot"},
		{"command": "status", "description": "System status"},
		{"command": "services", "description": "List all services"},
		{"command": "health", "description": "Health check"},
		{"command": "recent", "description": "Recent deployments"},
		{"command": "failures", "description": "Failed deployments"},
		{"command": "stats", "description": "System stats"},
		{"command": "help", "description": "Help menu"},
	}
	payload := map[string]any{"commands": commands}
	return postJSON(token, "setMyCommands", payload)
}

func sendChatAction(token string, chatID int64, action string) error {
	payload := map[string]any{"chat_id": chatID, "action": action}
	return postJSON(token, "sendChatAction", payload)
}

func postJSON(token, method string, payload map[string]any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/%s", token, method)
	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	io.ReadAll(resp.Body)
	resp.Body.Close()
	return nil
}

// ============ BOT LOGIC ============

type Bot struct {
	token   string
	router  *Router
	llm     *llm.OpenAIClient
	dbPath  string
}

type Router struct {
	client *aeroplane.Client
}

func NewBot(token, dbPath string, llmClient *llm.OpenAIClient) *Bot {
	client, _ := aeroplane.NewClient(dbPath)
	return &Bot{
		token:  token,
		router: &Router{client: client},
		llm:    llmClient,
		dbPath: dbPath,
	}
}

func (b *Bot) mainMenu(chatID int64) string {
	return `<b>🚀 Aeroplane Pilot</b>

Welcome! Use the buttons below or type commands.

<i>System monitoring for your VPS</i>

<b>Quick Stats:</b>
` + b.getQuickStats() + `

<b>Commands:</b>
• /status - System status
• /services - All services  
• /health - Health check
• /recent - Recent deployments
• /failures - Failed deployments
• /stats - Full system stats`
}

func (b *Bot) getQuickStats() string {
	var lines []string
	
	if b.router.client != nil {
		services, err := b.router.client.ListServices()
		if err == nil {
			running, stopped := 0, 0
			for _, s := range services {
				if strings.EqualFold(s.Status, "running") {
					running++
				} else {
					stopped++
				}
			}
			lines = append(lines, fmt.Sprintf("  ▸ Services: %d running, %d stopped", running, stopped))
		}
	}
	
	// System stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	lines = append(lines, fmt.Sprintf("  ▸ Memory: %.1f MB used", float64(m.Alloc)/1024/1024))
	lines = append(lines, fmt.Sprintf("  ▸ Uptime: %s", time.Since(startTime).Round(time.Second)))
	
	return strings.Join(lines, "\n")
}

func (b *Bot) handleCommand(chatID int64, cmd string) (string, *InlineKeyboardMarkup) {
	switch cmd {
	case "/start", "/help":
		return b.mainMenu(chatID), b.mainKeyboard()
	case "/status":
		return b.status(), b.statusKeyboard()
	case "/services":
		return b.services(), b.servicesKeyboard()
	case "/health":
		return b.health(), b.healthKeyboard()
	case "/recent":
		return b.recent(), b.actionKeyboard("recent")
	case "/failures":
		return b.failures(), b.actionKeyboard("failures")
	case "/stats":
		return b.stats(), b.statsKeyboard()
	}
	return "Unknown command. Use /help", nil
}

func (b *Bot) mainKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "📊 Status", CallbackData: "status"}, {Text: "🔧 Services", CallbackData: "services"}},
			{{Text: "❤️ Health", CallbackData: "health"}, {Text: "📈 Stats", CallbackData: "stats"}},
			{{Text: "🚀 Recent", CallbackData: "recent"}, {Text: "❌ Failures", CallbackData: "failures"}},
		},
	}
}

func (b *Bot) statusKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🔄 Refresh", CallbackData: "status"}, {Text: "📊 Services", CallbackData: "services"}},
			{{Text: "🔙 Menu", CallbackData: "menu"}},
		},
	}
}

func (b *Bot) servicesKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🔄 Refresh", CallbackData: "services"}, {Text: "❤️ Health", CallbackData: "health"}},
			{{Text: "🔙 Menu", CallbackData: "menu"}},
		},
	}
}

func (b *Bot) healthKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🔄 Refresh", CallbackData: "health"}, {Text: "❌ Failures", CallbackData: "failures"}},
			{{Text: "🔙 Menu", CallbackData: "menu"}},
		},
	}
}

func (b *Bot) statsKeyboard() *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🔄 Refresh", CallbackData: "stats"}, {Text: "📊 Status", CallbackData: "status"}},
			{{Text: "🔙 Menu", CallbackData: "menu"}},
		},
	}
}

func (b *Bot) actionKeyboard(action string) *InlineKeyboardMarkup {
	return &InlineKeyboardMarkup{
		InlineKeyboard: [][]InlineKeyboardButton{
			{{Text: "🔄 Refresh", CallbackData: action}, {Text: "🔙 Menu", CallbackData: "menu"}},
		},
	}
}

func (b *Bot) status() string {
	if b.router.client == nil {
		return "⚠️ Database not connected"
	}
	services, err := b.router.client.ListServices()
	if err != nil {
		return fmt.Sprintf("❌ Error: %v", err)
	}
	
	running, stopped, error := 0, 0, 0
	var lines []string
	for _, s := range services {
		status := "🟢"
		if strings.EqualFold(s.Status, "running") {
			running++
		} else if strings.EqualFold(s.Status, "error") {
			error++
			status = "🔴"
		} else {
			stopped++
			status = "🟡"
		}
		lines = append(lines, fmt.Sprintf("%s <b>%s</b>", status, s.Name))
	}
	
	return fmt.Sprintf(`<b>📊 System Status</b>

<b>Services:</b>
%v

<b>Summary:</b>
  🟢 Running: %d
  🟡 Stopped: %d
  🔴 Error: %d

<b>Updated:</b> %s`, strings.Join(lines, "\n"), running, stopped, error, time.Now().Format("15:04:05"))
}

func (b *Bot) services() string {
	if b.router.client == nil {
		return "⚠️ Database not connected"
	}
	services, err := b.router.client.ListServices()
	if err != nil {
		return fmt.Sprintf("❌ Error: %v", err)
	}
	
	var lines []string
	for _, s := range services {
		status := "🟢"
		if !strings.EqualFold(s.Status, "running") {
			status = "🟡"
		}
		lines = append(lines, fmt.Sprintf("%s <b>%s</b> — %s", status, s.Name, s.Status))
	}
	
	if len(lines) == 0 {
		return "📭 No services found"
	}
	
	return fmt.Sprintf(`<b>🔧 Services (%d)</b>

%v

<i>Updated %s</i>`, len(services), strings.Join(lines, "\n"), time.Now().Format("15:04"))
}

func (b *Bot) health() string {
	if b.router.client == nil {
		return "⚠️ Database not connected"
	}
	services, _ := b.router.client.ListServices()
	failed, _ := b.router.client.FailedDeployments(5)
	
	healthy, unhealthy := 0, 0
	for _, s := range services {
		if strings.EqualFold(s.Status, "running") {
			healthy++
		} else {
			unhealthy++
		}
	}
	
	emoji := "✅"
	if unhealthy > 0 || len(failed) > 0 {
		emoji = "⚠️"
	}
	if unhealthy > len(services)/2 {
		emoji = "🚨"
	}
	
	return fmt.Sprintf(`%s <b>Health Check</b>

<b>Services:</b>
  ✅ Healthy: %d
  ❌ Unhealthy: %d

<b>Deployments:</b>
  ❌ Failed (24h): %d

<b>Status:</b> %s`, emoji, healthy, unhealthy, len(failed), b.getHealthStatus(healthy, unhealthy, len(failed)))
}

func (b *Bot) getHealthStatus(healthy, unhealthy, failed int) string {
	if failed > 0 {
		return "DEGRADED - Check failures"
	}
	if unhealthy > 0 {
		return "WARNING - Some services down"
	}
	return "ALL SYSTEMS NOMINAL"
}

func (b *Bot) recent() string {
	if b.router.client == nil {
		return "⚠️ Database not connected"
	}
	deps, err := b.router.client.LatestDeployments(10)
	if err != nil {
		return fmt.Sprintf("❌ Error: %v", err)
	}
	
	var lines []string
	for _, d := range deps {
		status := "⏳"
		if strings.EqualFold(d.Status, "success") {
			status = "✅"
		} else if strings.EqualFold(d.Status, "failed") {
			status = "❌"
		}
		lines = append(lines, fmt.Sprintf("%s <b>%s</b>\n   Service: %s | %s", status, d.ID[:8], d.ServiceID, d.FinishedAt.String))
	}
	
	if len(lines) == 0 {
		return "📭 No recent deployments"
	}
	
	return fmt.Sprintf(`<b>🚀 Recent Deployments</b>

%v

<i>Updated %s</i>`, strings.Join(lines, "\n\n"), time.Now().Format("15:04"))
}

func (b *Bot) failures() string {
	if b.router.client == nil {
		return "⚠️ Database not connected"
	}
	deps, err := b.router.client.FailedDeployments(10)
	if err != nil {
		return fmt.Sprintf("❌ Error: %v", err)
	}
	
	var lines []string
	for _, d := range deps {
		lines = append(lines, fmt.Sprintf("❌ <b>%s</b>\n   Trigger: %s\n   Error: %s", d.ID[:8], d.Trigger, d.Status))
	}
	
	if len(lines) == 0 {
		return "✅ <b>No Failed Deployments</b>\n\nAll deployments are healthy!"
	}
	
	return fmt.Sprintf(`<b>❌ Failed Deployments</b>

%v

<i>Updated %s</i>`, strings.Join(lines, "\n\n"), time.Now().Format("15:04"))
}

func (b *Bot) stats() string {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	uptime := time.Since(startTime)
	
	return fmt.Sprintf(`<b>📈 System Statistics</b>

<b>Application:</b>
  ▸ Uptime: %s
  ▸ Memory: %.1f MB / %.1f MB
  ▸ Goroutines: %d
  ▸ Version: 1.0.0

<b>Go Runtime:</b>
  ▸ Alloc: %.1f MB
  ▸ Total Alloc: %.1f MB
  ▸ Sys: %.1f MB

<b>Database:</b>
  ▸ Path: %s
  ▸ Status: %s

<b>Updated:</b> %s`, 
		uptime.Round(time.Second),
		float64(m.Alloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		runtime.NumGoroutine(),
		float64(m.Alloc)/1024/1024,
		float64(m.TotalAlloc)/1024/1024,
		float64(m.Sys)/1024/1024,
		b.dbPath,
		b.getDBStatus(),
		time.Now().Format("15:04:05"),
	)
}

func (b *Bot) getDBStatus() string {
	if b.router.client == nil {
		return "❌ Not Connected"
	}
	return "✅ Connected"
}

var startTime = time.Now()

func (b *Bot) freeText(text string) string {
	if b.llm == nil {
		return "❌ LLM not configured"
	}
	resp, err := b.llm.Generate(text)
	if err != nil {
		return fmt.Sprintf("❌ LLM Error: %v", err)
	}
	return resp
}

// ============ MAIN ============

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	llmKey := os.Getenv("LLM_API_KEY")
	if llmKey == "" {
		llmKey = os.Getenv("OPENAI_API_KEY")
	}

	if token == "" || llmKey == "" {
		fmt.Println("Usage: TELEGRAM_BOT_TOKEN=xxx LLM_API_KEY=xxx ./pilot")
		os.Exit(1)
	}

	// Database path
	dbPath := "/opt/aeroplane/source/data/aeroplane.db"
	if home := os.Getenv("AEROPLANE_HOME"); home != "" {
		dbPath = home + "/source/data/aeroplane.db"
	}

	// LLM
	llmModel := os.Getenv("LLM_MODEL")
	if llmModel == "" {
		llmModel = "google/gemma-4-31b-it:free"
	}
	llmClient := llm.NewOpenAIClient(llmKey, llmModel, "")

	// Create bot
	bot := NewBot(token, dbPath, llmClient)

	// Register commands with BotFather
	setCommands(token)

	fmt.Println("🚀 Aeroplane Pilot running...")
	fmt.Println("   Bot: @ocmeogfegbot")
	fmt.Println("   DB: " + dbPath)

	// Main loop
	offset := 0
	for {
		updates, err := getUpdates(token, offset, 50)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		for _, u := range updates {
			offset = u.UpdateID + 1

			// Handle callback query (button press)
			if u.CallbackQuery != nil {
				chatID := u.CallbackQuery.Message.Chat.ID
				messageID := u.CallbackQuery.Message.MessageID
				data := u.CallbackQuery.Data

				var text string
				var markup *InlineKeyboardMarkup

				switch data {
				case "menu":
					text = bot.mainMenu(chatID)
					markup = bot.mainKeyboard()
				case "status":
					text = bot.status()
					markup = bot.statusKeyboard()
				case "services":
					text = bot.services()
					markup = bot.servicesKeyboard()
				case "health":
					text = bot.health()
					markup = bot.healthKeyboard()
				case "stats":
					text = bot.stats()
					markup = bot.statsKeyboard()
				case "recent":
					text = bot.recent()
					markup = bot.actionKeyboard("recent")
				case "failures":
					text = bot.failures()
					markup = bot.actionKeyboard("failures")
				default:
					text = "Unknown action"
				}

				editMessageText(token, chatID, messageID, text, markup)
				answerCallbackQuery(token, u.CallbackQuery.ID, "")
				continue
			}

			// Handle message
			if u.Message != nil && u.Message.Text != "" {
				chatID := u.Message.Chat.ID
				text := strings.TrimSpace(u.Message.Text)

				// Show typing indicator
				sendChatAction(token, chatID, "typing")

				// Check if it's a command
				if strings.HasPrefix(text, "/") {
					reply, markup := bot.handleCommand(chatID, text)
					sendMessage(token, chatID, reply, markup)
				} else {
					// Free text - use LLM
					sendChatAction(token, chatID, "typing")
					reply := bot.freeText(text)
					sendMessage(token, chatID, reply, nil)
				}
			}
		}
	}
}