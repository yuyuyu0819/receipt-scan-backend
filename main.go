package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Expo から受け取るリクエスト
type OCRRequest struct {
	Base64 string `json:"base64"`
}

// Vision API に送るペイロード
type visionRequest struct {
	Requests []struct {
		Image struct {
			Content string `json:"content"`
		} `json:"image"`
		Features []struct {
			Type string `json:"type"`
		} `json:"features"`
	} `json:"requests"`
}

// Vision API から返ってくるレスポンスのうち、使う部分だけ
type visionResponse struct {
	Responses []struct {
		FullTextAnnotation struct {
			Text string `json:"text"`
		} `json:"fullTextAnnotation"`
	} `json:"responses"`
}

func loadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env が見つかりませんでした（本番環境などでは正常な場合もあります）")
	}
}

// CORS ヘッダを付ける便利関数
func withCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")              // 手元開発なら * でOK（本番は絞る）
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func ocrHandler(w http.ResponseWriter, r *http.Request) {
	withCORS(w)

	if r.Method == http.MethodOptions {
		// CORS preflight 用
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req OCRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Base64 == "" {
		http.Error(w, "base64 is required", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("VISION_API_KEY")
	if apiKey == "" {
		http.Error(w, "VISION_API_KEY not set", http.StatusInternalServerError)
		return
	}

	// Vision API に送るリクエストを組み立て
	var vReq visionRequest
	vReq.Requests = []struct {
		Image struct {
			Content string `json:"content"`
		} `json:"image"`
		Features []struct {
			Type string `json:"type"`
		} `json:"features"`
	}{
		{
			Image: struct {
				Content string `json:"content"`
			}{Content: req.Base64},
			Features: []struct {
				Type string `json:"type"`
			}{
				{Type: "TEXT_DETECTION"},
			},
		},
	}

	bodyBytes, err := json.Marshal(vReq)
	if err != nil {
		http.Error(w, "marshal error", http.StatusInternalServerError)
		return
	}

	visionURL := fmt.Sprintf("https://vision.googleapis.com/v1/images:annotate?key=%s", apiKey)
	resp, err := http.Post(visionURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		log.Println("Vision API call error:", err)
		http.Error(w, "Vision API error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Vision API non-200:", resp.Status)
		http.Error(w, "Vision API error", http.StatusBadGateway)
		return
	}

	var vRes visionResponse
	if err := json.NewDecoder(resp.Body).Decode(&vRes); err != nil {
		log.Println("Vision API response decode error:", err)
		http.Error(w, "Vision response decode error", http.StatusInternalServerError)
		return
	}

	// OCR テキストを取り出す
	var text string
	if len(vRes.Responses) > 0 {
		text = vRes.Responses[0].FullTextAnnotation.Text
	}

	// Expo 側にそのまま返す（ここで GPT に回してもOK）
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"text": text,
	})
}

func main() {
	loadEnv()

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	http.HandleFunc("/ocr", ocrHandler)

	log.Printf("Go OCR server listening on :%s ...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
