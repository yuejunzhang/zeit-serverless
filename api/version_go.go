package handler
 

import (
	
	"io"
	"fmt"
	"log"
	"net/http"
	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

var (
	jar     = tls_client.NewCookieJar()
	options = []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(360),
		tls_client.WithClientProfile(tls_client.Chrome_112),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar), // create cookieJar instance and pass it as argument
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

	if r.URL.RawQuery != "" {
		url = "https://chat.openai.com/backend-api" + r.URL.Path + "?" + r.URL.RawQuery
	} else {
		url = "https://chat.openai.com/backend-api" + r.URL.Path
	}

	request_method = r.Method

	request, err = fhttp.NewRequest(request_method, url, r.Body)
	if err != nil {
		fmt.Fprintf(w, "<h1>Error creating request: %v</h1>", err)
		fmt.Fprintf(w, "<h2>Request:</h2>")
		fmt.Fprintf(w, "<pre>%v</pre>", request)
		return
	}
	request.Header.Set("Host", "chat.openai.com")
	request.Header.Set("Origin", "https://chat.openai.com/chat")
	request.Header.Set("Connection", "keep-alive")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Keep-Alive", "timeout=360")
	request.Header.Set("Authorization", r.Header.Get("Authorization"))
	request.Header.Set("sec-ch-ua", "\"Chromium\";v=\"112\", \"Brave\";v=\"112\", \"Not:A-Brand\";v=\"99\"")
	request.Header.Set("sec-ch-ua-mobile", "?0")
	request.Header.Set("sec-ch-ua-platform", "\"Linux\"")
	request.Header.Set("sec-fetch-dest", "empty")
	request.Header.Set("sec-fetch-mode", "cors")
	request.Header.Set("sec-fetch-site", "same-origin")
	request.Header.Set("sec-gpc", "1")
	request.Header.Set("user-agent", user_agent)
 	request.Header.Set("Accept",text/event-stream)

	response, err = client.Do(request)
	if err != nil {
		fmt.Fprintf(w, "<h1>Error creating request: %v</h1>", err)
		fmt.Fprintf(w, "<h2>Request:=========</h2>")
		fmt.Fprintf(w, "<pre>%v</pre>", request)
		return
	}

	defer response.Body.Close()
	w.Header().Set("Content-Type", response.Header.Get("Content-Type"))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	statusText := http.StatusText(response.StatusCode)
	w.Write([]byte("Response Status Code: " + strconv.Itoa(response.StatusCode)
	// Get status code
	w.WriteHeader(response.StatusCode)


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




	
//fmt.Fprintf(w, url)
	
}

