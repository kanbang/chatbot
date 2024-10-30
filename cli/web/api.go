package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/kevwan/chatbot/bot"
	"github.com/kevwan/chatbot/bot/adapters/logic"
	"github.com/kevwan/chatbot/bot/adapters/storage"
)

type ChatBotResponse struct {
	Answers []string `json:"answers"`
}

// getEnv 获取环境变量的值，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

var chatbot *bot.ChatBot

func ask(w http.ResponseWriter, r *http.Request) {
	// 获取响应
	startTime := time.Now()
	
	// 从请求中获取问题
	question := r.URL.Query().Get("question")
	if question == "" {
		http.Error(w, "Question parameter is missing", http.StatusBadRequest)
		return
	}

	answers := chatbot.GetResponse(question)
	if len(answers) == 0 {
		http.Error(w, "No answer found", http.StatusNotFound)
		return
	}

	// 准备响应数据
	response := ChatBotResponse{
		Answers: make([]string, len(answers)),
	}
	for i, answer := range answers {
		response.Answers[i] = answer.Content
	}

	// 设置响应头和编码
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	fmt.Println("Response time:", time.Since(startTime))
}

func main() {
	// 加载 .env 文件
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 从环境变量中读取端口号
	port := getEnv("PORT", "8090")

	// 从环境变量中读取语料库路径
	corpusPath := getEnv("CORPUS_PATH", "./corpus.gob")

	// 初始化存储和 ChatBot
	store, err := storage.NewSeparatedMemoryStorage(corpusPath)
	if err != nil {
		log.Fatal("Failed to initialize storage")
	}

	// 从环境变量中读取匹配数量
	matchCountStr := getEnv("MATCH_COUNT", "5")
	matchCount, err := strconv.Atoi(matchCountStr)
	if err != nil {
		log.Fatal("Invalid match count")
	}

	// 初始化 ChatBot
	chatbot = &bot.ChatBot{
		LogicAdapter: logic.NewClosestMatch(store, matchCount),
	}

	http.HandleFunc("/ask", ask)
	fmt.Printf("Server is running on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}