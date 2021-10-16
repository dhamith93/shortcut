package main

import (
	"bytes"
	"context"
	"encoding/json"
	"html"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type meta struct {
	Url         string
	MaxFileSize string
}

type clipboardItem struct {
	DeviceName string
	Content    string
}

type message struct {
	MsgType string
	Msg     interface{}
}

type handler struct {
	server         http.Server
	upgrader       websocket.Upgrader
	conn           *websocket.Conn
	hub            *hub
	config         config
	connectedIPs   []string
	fileList       []FileList
	clipboardItems []clipboardItem
}

func (h *handler) handleRequests() {
	router := mux.NewRouter().StrictSlash(true)
	router.Path("/meta").Methods("GET").HandlerFunc(h.sendMeta)
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		h.handleWS(h.hub, w, r)
	})
	router.Path("/upload").Methods("POST").Handler(h.checkRequest(h.handleFile))
	router.Path("/clipboard").Methods("POST").Handler(h.checkRequest(h.handleClipboardItem))
	router.Path("/clipboard").Methods("GET").Handler(h.checkRequest(h.getClipboardItems))
	router.Path("/files").Methods("GET").Handler(h.checkRequest(h.getFiles))
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./public/")))

	h.server.Addr = h.config.Port
	h.server.Handler = handlers.CompressHandler(router)
	h.server.SetKeepAlivesEnabled(false)

	Log("info", "Shortcut started on http://"+getOutboundIP()+h.config.Port)
	go openBrowser("http://" + getOutboundIP() + h.config.Port)
	log.Fatal(h.server.ListenAndServe())
}

func (h *handler) shutdown(ctx context.Context) error {
	return h.server.Shutdown(ctx)
}

func (h *handler) checkRequest(endpoint func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !contains(h.connectedIPs, strings.Split(r.RemoteAddr, ":")[0]) {
			Log("info", "Bad request - device count exceeded - "+r.RemoteAddr)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode("BadRequest")
			return
		}

		endpoint(w, r)
	})
}

func (h *handler) sendMeta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(&meta{
		Url:         "http://" + getOutboundIP() + h.server.Addr,
		MaxFileSize: h.config.MaxFileSize,
	})
}

func (h *handler) getFiles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&h.fileList)
}

func (h *handler) handleFile(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(32 << 20)
	var buf bytes.Buffer
	defer buf.Reset()
	file, header, err := r.FormFile("file")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	defer r.MultipartForm.RemoveAll()
	device := r.Form.Get("device-name")
	fileName := header.Filename

	msg, valid := h.validateFile(fileName, device, header.Size)

	if !valid {
		Log("error", msg)
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(msg)
		return
	}

	h.fileList, err = handleFile(file, device, fileName)
	if err != nil {
		Log("error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err.Error())
		return
	}

	go h.sendUpdate(message{"fileList", h.fileList})
	h.getFiles(w, r)
}

func (h *handler) handleClipboardItem(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	body, _ := ioutil.ReadAll(r.Body)
	clipboardItem := clipboardItem{}
	err := json.Unmarshal(body, &clipboardItem)
	if err != nil {
		Log("error", err.Error())
		w.WriteHeader(http.StatusBadRequest)
	}
	clipboardItem.Content = html.EscapeString(clipboardItem.Content)
	h.clipboardItems = append(h.clipboardItems, clipboardItem)

	go h.sendUpdate(message{"clipboardItems", h.clipboardItems})
	json.NewEncoder(w).Encode(h.clipboardItems)
}

func (h *handler) getClipboardItems(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.clipboardItems)
}

func (h *handler) handleWS(hub *hub, w http.ResponseWriter, r *http.Request) {
	if len(h.hub.clients) >= h.config.MaxDeviceCount {
		Log("info", "Allowed device count exceeded - rejected: "+r.RemoteAddr)
		http.Error(w, "Allowed device count exceeded", http.StatusForbidden)
		return
	}

	var err error
	h.upgrader = websocket.Upgrader{}
	h.conn, err = h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	client := &client{hub: hub, conn: h.conn, send: make(chan []byte, 256)}
	client.hub.register <- client
	h.connectedIPs = append(h.connectedIPs, strings.Split(r.RemoteAddr, ":")[0])

	go client.writePump()
	data, _ := json.Marshal([]message{{"fileList", h.fileList}, {"clipboardItems", h.clipboardItems}})
	client.send <- data
}

func (h *handler) sendUpdate(v ...interface{}) {
	data, _ := json.Marshal(v)

	for c := range h.hub.clients {
		select {
		case c.send <- data:
		default:
			delete(h.hub.clients, c)
			close(c.send)
		}
	}
}

func (h *handler) validateFile(fileName string, deviceName string, size int64) (string, bool) {
	if !validFileName(deviceName) || !validFileName(fileName) {
		return "invalid file name and/or device name " + deviceName + " > " + fileName, false
	}

	allowedSize, err := strconv.Atoi(strings.TrimSpace(strings.ReplaceAll(h.config.MaxFileSize, "MB", "")))
	if err != nil {
		log.Fatal(err)
	}
	allowedSize = allowedSize * 1024 * 1024

	if size > int64(allowedSize) {
		return "file size exceed max allowed size " + strconv.FormatInt(size, 10) + " > " + strconv.Itoa(allowedSize), false
	}

	return "", true
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func validFileName(fileName string) bool {
	return !strings.Contains(fileName, "..") && !strings.Contains(fileName, "\\") && !strings.Contains(fileName, "/")
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}
