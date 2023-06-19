package chatgpt

import (
	typings "freechatgpt/internal/typings" // 导入 typings 包，该包定义了 APIRequest 和 ChatGPTRequest 等类型
	"strings" // 导入 strings 包，该包提供了处理字符串的函数
)

// ConvertAPIRequest 函数将 APIRequest 类型的请求转换为 ChatGPTRequest 类型的请求
func ConvertAPIRequest(api_request typings.APIRequest) ChatGPTRequest {
	// 创建 ChatGPTRequest 类型的请求
	chatgpt_request := NewChatGPTRequest()

	// 如果 APIRequest 的 Model 字段以 "gpt-4" 开头，则将 ChatGPTRequest 的 Model 字段设置为 "gpt-4"
	if strings.HasPrefix(api_request.Model, "gpt-4") {
		chatgpt_request.Model = "gpt-4"
		// 如果 APIRequest 的 Model 字段是 "gpt-4-browsing"、"gpt-4-plugins"、"gpt-4-mobile" 或 "gpt-4-code-interpreter" 中的一个，则将 ChatGPTRequest 的 Model 字段设置为相应的值
		if api_request.Model == "gpt-4-browsing" || api_request.Model == "gpt-4-plugins" || api_request.Model == "gpt-4-mobile" || api_request.Model == "gpt-4-code-interpreter" {
			chatgpt_request.Model = api_request.Model
		}
	}

	// 遍历 APIRequest 中的每个消息
	for _, api_message := range api_request.Messages {
		// 如果消息的 Role 字段为 "system"，则将其设置为 "critic"
		if api_message.Role == "system" {
			api_message.Role = "critic"
		}
		// 将消息添加到 ChatGPTRequest 中
		chatgpt_request.AddMessage(api_message.Role, api_message.Content)
	}

	// 返回转换后的 ChatGPTRequest 请求
	return chatgpt_request
}
