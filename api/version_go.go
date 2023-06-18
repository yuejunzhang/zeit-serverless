package handler
 

import (
	
	"io"
	"os"
	"fmt"
	"log"
	"net/http"
	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	
	"freechatgpt/internal/chatgpt"
	"freechatgpt/internal/tokens"
	typings "freechatgpt/internal/typings"
	"freechatgpt/internal/typings/responses"
	"./chatgpt"
	"encoding/json"
)

var (
	jar     = tls_client.NewCookieJar()
	options = []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(360),
		//tls_client.WithClientProfile(tls_client.Chrome_112),
		tls_client.WithClientProfile(tls_client.Firefox_110),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
		tls_client.WithInsecureSkipVerify(),
	}
	client, _  = tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	user_agent = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/112.0.0.0 Safari/537.36"
	
)
func Handler(w http.ResponseWriter, r *http.Request) {
	var url string
	var err error
	var request_method string
	var request *fhttp.Request
	var response *fhttp.Response
	
	var original_request typings.APIRequest
	err := json.NewDecoder(r.Body).Decode(&original_request)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// 将聊天请求转换为 ChatGPT 请求
	translated_request := chatgpt.ConvertAPIRequest(original_request)
 	// 获取访问令牌
	token := ACCESS_TOKENS.GetToken()

	response, err := chatgpt.SendRequest(translated_request, token)
	if err != nil {
	    http.Error(w, response.Status, response.StatusCode)
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
		    w.WriteHeader(http.StatusInternalServerError)
		    w.Write([]byte(fmt.Sprintf(`{"error": {"message": "Unknown error", "type": "internal_server_error", "param": null, "code": "500", "details": %q}}`, string(body))))
		    return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
		    "error": map[string]interface{}{
		        "message": error_response["detail"],
		        "type":    response.Status,
		        "param":   nil,
		        "code":    "error",
		    },
		})

		return
	}
	// 创建一个从响应体中读取数据的 bufio.Reader、、、、、、、、、、、、、、、、、、、、、、、、、、、、、、、、、
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

	// var url string
	// var err error
	// var request_method string
	// var request *fhttp.Request
	// var response *fhttp.Response

	// if r.URL.RawQuery != "" {
	// 	url = "https://chat.openai.com/backend-api" + r.URL.Path + "?" + r.URL.RawQuery
	// } else {
	// 	url = "https://chat.openai.com/backend-api" + r.URL.Path
	// }

	// request_method = r.Method

	// request, err = fhttp.NewRequest(request_method, url, r.Body)
	// if err != nil {
	// 	fmt.Fprintf(w, "<h1>Error creating request: %v</h1>", err)
	// 	fmt.Fprintf(w, "<h2>Request:</h2>")
	// 	fmt.Fprintf(w, "<pre>%v</pre>", request)
	// 	return
	// }
	// request.Header.Set("Host", "chat.openai.com")
	// request.Header.Set("Origin", "https://chat.openai.com/chat")
	// request.Header.Set("Connection", "keep-alive")
	// request.Header.Set("Content-Type", "application/json")
	// request.Header.Set("Keep-Alive", "timeout=360")
	// request.Header.Set("Authorization", r.Header.Get("Authorization"))
	// request.Header.Set("sec-ch-ua", "\"Chromium\";v=\"112\", \"Brave\";v=\"112\", \"Not:A-Brand\";v=\"99\"")
	// request.Header.Set("sec-ch-ua-mobile", "?0")
	// request.Header.Set("sec-ch-ua-platform", "\"Linux\"")
	// request.Header.Set("sec-fetch-dest", "empty")
	// request.Header.Set("sec-fetch-mode", "cors")
	// request.Header.Set("sec-fetch-site", "same-origin")
	// request.Header.Set("sec-gpc", "1")
	// request.Header.Set("user-agent", user_agent)
 // 	//request.Header.Set("Accept",text/event-stream)
	// if os.Getenv("PUID") != "" {
	// 	// request.AddCookie(&fhttp.Cookie{Name: "_puid", Value: os.Getenv("PUID")})
	// 	request.Header.Set("Cookie", "_puid="+os.Getenv("PUID")+";")
	// }
	// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// response, err = client.Do(request)
	// if err != nil {
	// 	fmt.Fprintf(w, "<h1>Error creating request: %v</h1>", err)
	// 	fmt.Fprintf(w, "<h2>Request:=========</h2>")
	// 	fmt.Fprintf(w, "<pre>%v</pre>", request)
	// 	return
	// }

	// defer response.Body.Close()
	// w.Header().Set("Content-Type", response.Header.Get("Content-Type"))
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Headers", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "*")

	// // Get status code
	// w.WriteHeader(response.StatusCode)

	// /////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// buf := make([]byte, 4096)
	// for {
	// 	n, err := response.Body.Read(buf)
	// 	if n > 0 {
	// 		_, writeErr :=w.Write(buf[:n])
	// 		if writeErr != nil {
	// 			log.Printf("Error writing to client: %v", writeErr)
	// 			break
	// 		}
			 
	// 	}
	// 	if err == io.EOF {
	// 		break
	// 	}
	// 	if err != nil {
	// 		log.Printf("Error reading from response body: %v", err)
	// 		break
	// 	}
	// }




	// statusText := http.StatusText(response.StatusCode)
	// fmt.Fprintf(w, url+statusText)
	// // fmt.Fprintf(w, "<h1>response: %v</h1>", err)
	// // fmt.Fprintf(w, "<h2>response:</h2>")
	// // fmt.Fprintf(w, "<pre>%v</pre>", response)	
	
}

