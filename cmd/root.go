package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/coolapso/go-live-server/internal/util"
	"github.com/fsnotify/fsnotify"
	"golang.org/x/net/websocket"

	"github.com/spf13/cobra"
)

var (
	watchDir  string
	port      string
	clients   sync.Map
	broadcast = make(chan bool)
	browser   bool
	file      string
)

var rootCmd = &cobra.Command{
	Use:   "live-server",
	Short: "A simple development server with live reloading",
	Long: `go-live-server is a simple development webserver with live reloading capabilityes. 
It allows you to quickly edit and visualize changes when developing simple html and css files`,

	PreRun: func(cmd *cobra.Command, args []string) {
		preRun()
	},

	Run: func(cmd *cobra.Command, args []string) {
		liveServer()
	},
}

func preRun() {
	if !strings.HasPrefix(port, ":") {
		port = fmt.Sprintf(":%s", port)
	}
}

// dirFinder Receives a watch dir and finds all directories within it to be watched
func getSubDirs(d string) (dirs []string, err error) {
	err = filepath.Walk(d, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Failed to access path %q: %v\n", path, err)
		}

		if info.IsDir() {
			dirs = append(dirs, path)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return dirs, nil
}

// watcher watches directories for changes
func watcher(dirs []string, broadcast chan (bool)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Failed to create new watcher:", err)
	}
	defer watcher.Close()

	for _, dir := range dirs {
		if err := watcher.Add(dir); err != nil {
			log.Fatalf("Failed to add %v to watcher:", dir)
		}
	}
	log.Println("Watching dirs:", watcher.WatchList())

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Remove) {
				log.Println("Modified file:", event.Name)
				broadcast <- true
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

func renderWebhookScript(port string) string {
	return fmt.Sprintf(`<script>
    const ws = new WebSocket("ws://localhost%s/ws");
	ws.onopen = function(event) {
		console.log('Live reload enabled.');
	};

    ws.onmessage = function(event) {
        if (event.data === "reload") {
            window.location.reload();
			ws.close()
        }
	};

	ws.onclose = function(event) {
		console.log('socket  connection terminated.');
	};
</script>
`, port)
}

func injectScript(w http.ResponseWriter, script, html string) {
	if strings.Contains(html, "</head>") {
		injectContent := script + "</head>"
		html = strings.Replace(html, "</head>", injectContent, 1)
	}

	w.Header().Set("Content-Type", "text/html")
	if _, err := w.Write([]byte(html)); err != nil {
		http.Error(w, "Failed to write content to response writer", http.StatusInternalServerError)
		log.Println("Failed to write content to response writer:", err)
	}
}

func fileHandler(w http.ResponseWriter, r *http.Request, wd string) {
	filePath := filepath.Join(wd, r.URL.Path)
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			log.Println("Not found:", filePath)
			return
		}

		log.Println("Failed to get info from:", filePath)
		return
	}

	if info.IsDir() {
		filePath = filepath.Join(filePath, "index.html")
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.NotFound(w, r)
		log.Println("File not found:", filePath)
		return
	}

	switch {
	case strings.HasSuffix(filePath, ".html"):
		html, err := os.ReadFile(filePath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to open file: %v", filePath), http.StatusInternalServerError)
			log.Println("Failed to open file:", filePath)
		}
		injectScript(w, renderWebhookScript(port), string(html))
	default:
		http.ServeFile(w, r, filePath)
	}
}

func handleWebSocketConnections(ws *websocket.Conn, clients *sync.Map, broadcast chan bool) {
	clients.Store(ws, true)
	for {
		<-broadcast
		clients.Range(func(key, value interface{}) bool {
			c := key.(*websocket.Conn)
			err := websocket.Message.Send(c, "reload")
			if err != nil {
				if !errors.Is(err, syscall.EPIPE) && !errors.Is(err, net.ErrClosed) {
					log.Println("Error sending reload message:", err)
				}
			}
			clients.Delete(c)
			c.Close()
			return true
		})
	}
}

func liveServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fileHandler(w, r, watchDir)
	})
	http.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		handleWebSocketConnections(ws, &clients, broadcast)
	}))
	dirs, err := getSubDirs(watchDir)
	if err != nil {
		log.Fatal(err)
	}

	go watcher(dirs, broadcast)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Printf("listening on %v, press CTRC+C to terminate the server\n", port)
		if err := http.ListenAndServe(port, nil); err != nil {
			log.Fatal(err)
		}
	}()

	if browser {
		time.Sleep(50 * time.Millisecond)
		_ = util.OpenURL(fmt.Sprintf("http://localhost%s/%s", port, file))
	}
	wg.Wait()
}

func init() {
	rootCmd.Flags().StringVarP(&watchDir, "watch-dir", "d", "./", "Sets the directory to watch for")
	rootCmd.Flags().StringVarP(&port, "port", "p", ":8080", "The port server is going to listen on")
	rootCmd.Flags().BoolVar(&browser, "browser", true, "Enable or disable automatic opening of the browser")
	rootCmd.Flags().StringVar(&file, "open-file", "", "Specify the relative path to open the browser in the directory being served")
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
