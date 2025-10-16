package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/2bitburrito/http-implementation/internal/headers"
	"github.com/2bitburrito/http-implementation/internal/request"
	"github.com/2bitburrito/http-implementation/internal/response"
	"github.com/2bitburrito/http-implementation/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, Handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

type rtnMsg struct {
	Title   string
	Status  string
	Message string
}

//go:embed static/template.html
var htmlTemplate string

func Handler(w *response.Writer, r *request.Request) {
	pStr := strings.Trim(r.RequestLine.RequestTarget, "/")
	paths := strings.Split(pStr, "/")
	switch paths[0] {
	case "yourproblem":
		handle400(w, r)

	case "myproblem":
		handle500(w, r)

	case "httpbin":
		handleHTTPBin(w, r, paths)

	case "video":
		serveVideo(w, r)

	default:
		handleDefault(w, r)
	}
}

func handle400(w *response.Writer, _ *request.Request) {
	msg := rtnMsg{}
	tmpl := template.New("template")
	tmpl.Parse(htmlTemplate)
	buf := bytes.Buffer{}
	status := "Bad Request"
	msg.Title = fmt.Sprintf("%d %s", 400, status)
	msg.Status = status
	msg.Message = "Your request honestly kinda sucked."
	if err := tmpl.Execute(&buf, msg); err != nil {
		fmt.Printf("Bad template execution")
		return
	}
	w.WriteStatusLine(400)
	h := response.GetDefaultHeaders(buf.Len())
	h["Content-Type"] = "text/html"
	w.WriteHeaders(h)
	w.WriteBody(buf.Bytes())
}

func handle500(w *response.Writer, _ *request.Request) {
	msg := rtnMsg{}
	tmpl := template.New("template")
	tmpl.Parse(htmlTemplate)
	buf := bytes.Buffer{}
	status := "Internal Server Error"
	msg.Title = fmt.Sprintf("%d %s", 500, status)
	msg.Status = status
	msg.Message = "Okay, you know what? This one is on me."
	if err := tmpl.Execute(&buf, msg); err != nil {
		fmt.Printf("Bad template execution")
		return
	}
	w.WriteStatusLine(500)
	h := response.GetDefaultHeaders(buf.Len())
	h["Content-Type"] = "text/html"
	w.WriteHeaders(h)
	w.WriteBody(buf.Bytes())
}

func handleHTTPBin(w *response.Writer, _ *request.Request, paths []string) {
	url := fmt.Sprintf("https://httpbin.org/%s", paths[1])
	fmt.Println("URL", url)
	bResp, err := http.Get(url)
	if err != nil {
		fmt.Printf("bad request to: ", url)
		w.WriteStatusLine(500)
		return
	}
	defer bResp.Body.Close()
	w.WriteStatusLine(200)

	h := map[string]string{
		"connection":        "close",
		"Transfer-Encoding": "chunked",
		"Trailer":           "X-Content-SHA256, X-Content-Length",
	}
	w.WriteHeaders(h)

	totalB := make([]byte, 0, 1024)
	b := make([]byte, 1024)
	for {
		n, err := bResp.Body.Read(b)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Println("error while reading body :", err)
			return
		}
		w.WriteChunkedBody(b[:n])
		totalB = append(totalB, b[:n]...)
	}
	w.WriteChunkedBodyDone()

	sha := sha256.New()
	sha.Write(totalB)
	sum := sha.Sum(nil)

	trailers := headers.Headers{
		"X-Content-Length": fmt.Sprintf("%d", len(totalB)),
		"X-Content-SHA256": fmt.Sprintf("%x", sum),
	}
	w.WriteTrailers(trailers)
}

func serveVideo(w *response.Writer, _ *request.Request) {
	file, err := os.ReadFile("./assets/vim.mp4")
	if err != nil {
		fmt.Println("couldn't open file: ", err)
		return
	}
	h := headers.Headers{
		"connection":     "close",
		"Content-Type":   "video/mp4",
		"Content-Length": fmt.Sprintf("%d", len(file)),
	}
	w.WriteStatusLine(200)
	if err := w.WriteHeaders(h); err != nil {
		fmt.Println("bad writing to headers", err)
		return
	}
	if _, err := w.WriteBody(file); err != nil {
		fmt.Println("couldn't write body: ", err)
		return
	}
}

func handleDefault(w *response.Writer, _ *request.Request) {
	msg := rtnMsg{}
	tmpl := template.New("template")
	tmpl.Parse(htmlTemplate)
	buf := bytes.Buffer{}
	msg.Title = fmt.Sprintf("%d OK", 200)
	msg.Status = "Success!"
	msg.Message = "Your request was an absolute banger."
	if err := tmpl.Execute(&buf, msg); err != nil {
		fmt.Printf("Bad template execution")
		return
	}
	w.WriteStatusLine(200)
	h := response.GetDefaultHeaders(buf.Len())
	h["Content-Type"] = "text/html"
	w.WriteHeaders(h)
	w.WriteBody(buf.Bytes())
}
