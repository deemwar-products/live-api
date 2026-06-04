package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"

	_ "embed"

	"github.com/gorilla/websocket"
	"google.golang.org/genai"
)

//go:embed live_streaming.html
var homeTemplate string

const geminiModel = "models/gemini-3.1-flash-live-preview"

var (
	addr = flag.String("addr", ":8080", "HTTP service address")

	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

func main() {
	flag.Parse()

	if os.Getenv("GEMINI_API_KEY") == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", homePage)
	mux.HandleFunc("/live", liveHandler)
	mux.HandleFunc("/proxyVideo", proxyVideo)

	log.Printf("Server listening on http://localhost%s", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.New("home").Parse(homeTemplate)
	if err != nil {
		http.Error(w, "template parse error", http.StatusInternalServerError)
		return
	}
	// Pass the base WS URL — the frontend appends ?mode=text or ?mode=audio
	wsURL := "ws://" + r.Host + "/live"
	if err := tmpl.Execute(w, wsURL); err != nil {
		log.Printf("template execute error: %v", err)
	}
}

// liveHandler proxies messages between the browser and the Gemini Live API.
// Query param ?mode=text uses text modality; default is audio.
func liveHandler(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "audio"
	}

	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket upgrade error: %v", err)
		return
	}
	defer clientConn.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  os.Getenv("GEMINI_API_KEY"),
		Backend: genai.BackendGeminiAPI,
		HTTPOptions: genai.HTTPOptions{
			APIVersion: "v1alpha",
		},
	})
	if err != nil {
		log.Printf("genai client error: %v", err)
		return
	}

	config := buildConfig(mode)
	session, err := client.Live.Connect(ctx, geminiModel, config)
	if err != nil {
		log.Printf("live connect error: %v", err)
		return
	}
	defer session.Close()

	log.Printf("session started [mode=%s]", mode)

	errCh := make(chan error, 2)

	// Gemini → browser
	go func() {
		for {
			msg, err := session.Receive()
			if err != nil {
				errCh <- fmt.Errorf("receive from gemini: %w", err)
				return
			}
			b, err := json.Marshal(msg)
			if err != nil {
				errCh <- fmt.Errorf("marshal gemini message: %w", err)
				return
			}
			if err := clientConn.WriteMessage(websocket.TextMessage, b); err != nil {
				errCh <- fmt.Errorf("write to browser: %w", err)
				return
			}
		}
	}()

	// browser → Gemini
	go func() {
		for {
			_, raw, err := clientConn.ReadMessage()
			if err != nil {
				errCh <- fmt.Errorf("read from browser: %w", err)
				return
			}
			if err := forwardToGemini(session, mode, raw); err != nil {
				errCh <- err
				return
			}
		}
	}()

	if err := <-errCh; err != nil {
		log.Printf("session closed [mode=%s]: %v", mode, err)
	}
}

func buildConfig(mode string) *genai.LiveConnectConfig {
	if mode == "text" {
		return &genai.LiveConnectConfig{
			ResponseModalities: []genai.Modality{genai.ModalityText},
		}
	}
	return &genai.LiveConnectConfig{
		ResponseModalities:       []genai.Modality{genai.ModalityAudio},
		InputAudioTranscription:  &genai.AudioTranscriptionConfig{},
		OutputAudioTranscription: &genai.AudioTranscriptionConfig{},
	}
}

func forwardToGemini(session *genai.Session, mode string, raw []byte) error {
	if mode == "text" {
		var input genai.LiveClientContentInput
		if err := json.Unmarshal(raw, &input); err != nil {
			return fmt.Errorf("unmarshal text message: %w", err)
		}
		return session.SendClientContent(input)
	}
	var input genai.LiveRealtimeInput
	if err := json.Unmarshal(raw, &input); err != nil {
		return fmt.Errorf("unmarshal audio message: %w", err)
	}
	return session.SendRealtimeInput(input)
}

func proxyVideo(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get("http://storage.googleapis.com/cloud-samples-data/video/animals.mp4")
	if err != nil {
		http.Error(w, "error fetching video", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(w, resp.Body) //nolint:errcheck — broken pipe on client disconnect is expected
}
