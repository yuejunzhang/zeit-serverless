package main

import (
	"bufio"
	"encoding/json"
	"freechatgpt/internal/chatgpt"
	"freechatgpt/internal/tokens"
	typings "freechatgpt/internal/typings"
	"freechatgpt/internal/typings/responses"
	"io"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// passwordHandler 用于更新管理员密码
func passwordHandler(c *gin.Context) {
	// 从请求中获取密码并更新密码
	type password_struct struct {
		Password string `json:"password"`
	}
	var password password_struct
	err := c.BindJSON(&password)
	if err != nil {
		c.String(400, "password not provided")
		return
	}
	ADMIN_PASSWORD = password.Password
	// 设置环境变量
	os.Setenv("ADMIN_PASSWORD", ADMIN_PASSWORD)
	c.String(200, "password updated")
}

// puidHandler 用于更新 PUID
func puidHandler(c *gin.Context) {
	// 从请求中获取 PUID 并更新
	type puid_struct struct {
		PUID string `json:"puid"`
	}
	var puid puid_struct
	err := c.BindJSON(&puid)
	if err != nil {
		c.String(400, "puid not provided")
		return
	}
	// 设置环境变量
	os.Setenv("PUID", puid.PUID)
	c.String(200, "puid updated")
}

// tokensHandler 用于更新访问令牌
func tokensHandler(c *gin.Context) {
	// 从请求中获取请求令牌并更新
	var request_tokens []string
	err := c.BindJSON(&request_tokens)
	if err != nil {
		c.String(400, "tokens not provided")
		return
	}
	ACCESS_TOKENS = tokens.NewAccessToken(request_tokens)
	c.String(200, "tokens updated")
}

// optionsHandler 用于设置 CORS 响应头
func optionsHandler(c *gin.Context) {
	// 设置 CORS 响应头
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST")
	c.Header("Access-Control-Allow-Headers", "*")
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// nightmare 是一个聊天请求处理函数
func nightmare(c *gin.Context) {
	var original_request typings.APIRequest
	err := c.BindJSON(&original_request)
	if err != nil {
		c.JSON(400, gin.H{"error": gin.H{
			"message": "Request must be proper JSON",
			"type":    "invalid_request_error",
			"param":   nil,
			"code":    err.Error(),
		}})
	}
	// 将聊天请求转换为 ChatGPT 请求
	translated_request := chatgpt.ConvertAPIRequest(original_request)

	// 获取访问令牌
	token := ACCESS_TOKENS.GetToken()

	response, err := chatgpt.SendRequest(translated_request, token)
	if err != nil {
		c.JSON(response.StatusCode, gin.H{
			"error":   "error sending request",
			"message": response.Status,
		})
		return
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		// 尝试将响应体解析为 JSON
		var error_response map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&error_response)
		if err != nil {
			// 读取响应体
			body, _ := io.ReadAll(response.Body)
			c.JSON(500, gin.H{"error": gin.H{
				"message": "Unknown error",
				"type":    "internal_server_error",
				"param":   nil,
				"code":    "500",
				"details": string(body),
			}})
			return
		}
		c.JSON(response.StatusCode, gin.H{"error": gin.H{
			"message": error_response["detail"],
			"type":    response.Status,
			"param":   nil,
			"code":    "error",
		}})
		return
	}
	// 创建一个从响应体中读取数据的 bufio.Reader
	reader := bufio.NewReader(response.Body)

	var fulltext string

	// 逐字节读取响应，直到遇到换行符
	if original_request.Stream {
		// 响应内容类型为 text/event-stream
		c.Header("Content-Type", "text/event-stream")
	} else {
		// 响应内容类型为 application/json
		c.Header("Content-Type", "application/json")
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return
		}
		if len(line) < 6 {
			continue
		}
		// 去掉行开头的 "data: "
		line = line[6:]
		// 检查行是否以 [DONE] 开头
		if !strings.HasPrefix(line, "[DONE]") {
			// 将行解析为 JSON
			var original_response responses.Data
			err = json.Unmarshal([]byte(line), &original_response)
			if err != nil {
				continue
			}
			if original_response.Error != nil {
				return
			}
			if original_response.Message.Content.Parts == nil {
				continue
			}
			if original_response.Message.Content.Parts[0] == "" || original_response.Message.Author.Role != "assistant" {
				continue
			}
			if original_response.Message.Metadata.Timestamp == "absolute" {
				continue
			}
			tmp_fulltext := original_response.Message.Content.Parts[0]
			original_response.Message.Content.Parts[0] = strings.ReplaceAll(original_response.Message.Content.Parts[0], fulltext, "")
			translated_response := responses.NewChatCompletionChunk(original_response.Message.Content.Parts[0])

			// 将响应流式传输到客户端
			response_string := translated_response.String()
			if original_request.Stream {
				_, err = c.Writer.WriteString("data: " + string(response_string) + "\n\n")
				if err != nil {
					return
				}
			}

			// 刷新响应写入缓冲区，以确保客户端接收到每一行数据
			c.Writer.Flush()
			fulltext = tmp_fulltext
		} else {
			if !original_request.Stream {
				full_response := responses.NewChatCompletion(fulltext)
				if err != nil {
					return
				}
				c.JSON(200, full_response)
				return
			}
			final_line := responses.StopChunk()
			c.Writer.WriteString("data: " + final_line.String() + "\n\n")

			c.String(200, "data: [DONE]\n\n")
			return

		}
	}

}
