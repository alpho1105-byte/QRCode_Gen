package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

const baseURL = "http://localhost:8080"

func main() {
	/*
		// 1. 建立 QR code
		fmt.Println("=== POST /v1/qr_code ===")
		body := []byte(`{"url":"http://google.com"}`)
		resp, err := http.Post(baseURL+"/v1/qr_code", "application/json", bytes.NewBuffer(body))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Println("Status:", resp.StatusCode)
		fmt.Println("Body:", string(respBody))

		// 從回應中取出 qr_token
		var result map[string]string
		json.Unmarshal(respBody, &result)
		token := result["qr_token"]
		fmt.Println("Token:", token)
	*/
	token := "od21YZRI"
	// 1.5 取得 QR code 圖片
	fmt.Println("\n=== GET /v1/qr_code_image/{token} ===")
	resp, err := http.Get(baseURL + "/v1/qr_code_image/" + token + "?dimension=300")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	imgData, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Content-Type:", resp.Header.Get("Content-Type"))
	fmt.Println("Image size:", len(imgData), "bytes")

	// 把圖片存到檔案，可以用圖片檢視器打開
	os.WriteFile("qr_output.png", imgData, 0644)
	fmt.Println("Saved to qr_output.png")

	// 2. 查詢原始 URL
	fmt.Println("\n=== GET /v1/qr_code/{token} ===")
	resp, err = http.Get(baseURL + "/v1/qr_code/" + token)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	respBody, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Body:", string(respBody))

	// 3. 測試 redirect（不自動跟隨）
	fmt.Println("\n=== GET /r/{token} (redirect) ===")
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // 不自動跟隨 redirect
		},
	}
	resp, err = client.Get(baseURL + "/r/" + token)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	resp.Body.Close()
	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Location:", resp.Header.Get("Location"))

	/*
		// 4. 刪除
		fmt.Println("\n=== DELETE /v1/qr_code/{token} ===")
		req, _ := http.NewRequest("DELETE", baseURL+"/v1/qr_code/"+token, nil)
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		resp.Body.Close()
		fmt.Println("Status:", resp.StatusCode)

		// 5. 再查一次（應該 404）
		fmt.Println("\n=== GET /v1/qr_code/{token} (after delete) ===")
		resp, err = http.Get(baseURL + "/v1/qr_code/" + token)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	*/
	respBody, _ = io.ReadAll(resp.Body)
	resp.Body.Close()
	fmt.Println("Status:", resp.StatusCode)
	fmt.Println("Body:", string(respBody))
}
