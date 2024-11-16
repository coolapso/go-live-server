package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"

	// "os"
	"slices"
	"time"

	"golang.org/x/net/websocket"
)

const (
	supressLogs = true
)

func TestMain(m *testing.M) {
	if supressLogs {
		if err := setLogsDevNull(); err != nil {
			fmt.Fprintf(os.Stderr, "Error setting logs to devnull: %v\n", err)
		}
	}

	code := m.Run()
	os.Exit(code)
}

func TestPreRun(t *testing.T) {
	t.Run("test wihtout semicolon", func(t *testing.T) {
		want := ":1080"
		port = "1080"
		preRun()
		if want != port {
			t.Fatalf("Port doesn't match, want %s, got %s", want, port)
		}
	})

	t.Run("test with semicolon", func(t *testing.T) {
		want := ":1080"
		port = ":1080"
		preRun()
		if want != port {
			t.Fatalf("Port doesn't match, want %s, got %s", want, port)
		}
	})
}

func TestDirFinder(t *testing.T) {
	t.Run("Test find dir", func(t *testing.T) {
		want := []string{"../test", "../test/css", "../test/images", "../test/images/testImages"}
		got, err := getSubDirs("../test")
		if err != nil {
			t.Fatal("Got error, didn't expect one:", err)
		}

		if len(want) != len(got) {
			t.Fatalf("Want list with %v got list with %v", len(want), len(got))
		}

		for _, d := range got {
			if !slices.Contains(want, d) {
				t.Fatalf("want %v, got %v", want, got)
			}
		}
	})
	t.Run("Test error", func(t *testing.T) {
		_, err := getSubDirs("fooBarDir")
		if err == nil {
			t.Fatalf("Didn't get any error, expected one")
		}
	})
}

func TestWatcher(t *testing.T) {
	watchDirs := []string{"../test", "../test/css", "../test/images", "../test/images/testImages"}
	testFile := "../test/css/newFile.txt"
	broadcast := make(chan bool)

	go watcher(watchDirs, broadcast)
	time.Sleep(100 * time.Millisecond)

	t.Run("Test file creation watcher", func(t *testing.T) {
		createTestFile(t, testFile)
		defer os.Remove(testFile)

		select {
		case <-broadcast:
		case <-time.After(time.Second):
			t.Fatal("Did not receive filechange event")
		}
	})

	t.Run("Test file deletion watcher", func(t *testing.T) {
		createTestFile(t, testFile)
		if err := os.Remove(testFile); err != nil {
			t.Fatal("Failed to remove test file:", err)
		}
		defer os.Remove(testFile)

		select {
		case <-broadcast:
		case <-time.After(time.Second):
			t.Fatal("Did not receive filechange event")
		}
	})
}

func TestRenderWebhookScript(t *testing.T) {
	port = ":8080"
	want := fmt.Sprintf(`<script>
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
</script>`, port)
	want = strings.Join(strings.Fields(want), " ")

	got := renderWebhookScript(port)

	got = strings.Join(strings.Fields(got), " ")
	if want != got {
		t.Fatalf("Script body doesn't match. Want:\n%s\n\nGot:\n\n%s\n", want, got)
	}
}

func TestInjectScript(t *testing.T) {
	htmlContent := "<html><head></head><body>Hello World</body></html>"
	script := "Some script to be injected"
	want := "<html><head>" + script + "</head><body>Hello World</body></html>"

	t.Run("Test Injection", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		injectScript(recorder, script, htmlContent)
		got := recorder.Body.String()

		if want != got {
			t.Fatalf("Html body doesn't match expected. want:\n%s\ngot:\n%s\n", want, got)
		}
	})
}

func TestFileHandler(t *testing.T) {
	watchDir = "../test"

	t.Run("Test non existing directory", func(t *testing.T) {
		request := &http.Request{
			URL: &url.URL{Path: "./foo/"},
		}

		recorder := httptest.NewRecorder()
		fileHandler(recorder, request, watchDir)
		responseStatus := recorder.Result().StatusCode
		if responseStatus != 404 {
			t.Fatalf("Expected 404, got %v", responseStatus)
		}
	})

	t.Run("Test root dir without html", func(t *testing.T) {
		request := &http.Request{
			URL: &url.URL{Path: "./images"},
		}

		recorder := httptest.NewRecorder()
		fileHandler(recorder, request, watchDir)
		responseStatus := recorder.Result().StatusCode
		if responseStatus != 404 {
			t.Fatalf("Expected 404, got %v", responseStatus)
		}
	})

	t.Run("Test serve html", func(t *testing.T) {
		request := &http.Request{
			URL: &url.URL{Path: "./"},
		}

		want, err := os.ReadFile("../test/injected.html")
		if err != nil {
			t.Fatal("Failed to open test file", err)
		}

		recorder := httptest.NewRecorder()
		fileHandler(recorder, request, watchDir)
		got := recorder.Body.String()
		if string(want) != got {
			t.Fatalf("Response body does not match expected. want:\n%s\ngot:\n%s\n", want, got)
		}
	})

	t.Run("Test serve file", func(t *testing.T) {
		request := &http.Request{
			URL: &url.URL{Path: "./test.txt"},
		}

		want := "hello world!"

		recorder := httptest.NewRecorder()
		fileHandler(recorder, request, watchDir)
		got := strings.TrimSpace(recorder.Body.String())
		if want != got {
			t.Fatalf("Response body does not match expected. want:\n%s\ngot:\n%s\n", want, got)
		}
	})

}

func TestHandleWebSocketConnections(t *testing.T) {
	var clients sync.Map
	reload := make(chan bool)
	server := httptest.NewServer(websocket.Handler(func(ws *websocket.Conn) {
		handleWebSocketConnections(ws, &clients, reload)
	}))
	defer server.Close()

	wsURL := "ws" + server.URL[len("http"):]
	ws, err := websocket.Dial(wsURL, "", "http://localhost/")
	if err != nil {
		t.Fatalf("failed to connect to WebSocket server: %v", err)
	}
	defer ws.Close()

	time.Sleep(50 * time.Millisecond)

	tclients := 0
	clients.Range(func(_, _ interface{}) bool {
		tclients++
		return true
	})
	if tclients != 1 {
		t.Fatalf("Failed, expected 1 client got %v", tclients)
	}

	select {
	case reload <- true:
		var message string
		if err := websocket.Message.Receive(ws, &message); err != nil {
			t.Fatalf("got error receiving message, didn't expect one: %v", err)
		}

		if message != "reload" {
			t.Fatalf("Expected message 'reload', got %s", message)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("test timeout waiting for reload")
	}

	ws.Close()
	time.Sleep(100 * time.Millisecond)

	tclients = 0
	clients.Range(func(_, _ interface{}) bool {
		tclients++
		return true
	})

	if tclients != 0 {
		t.Fatalf("Failed, expected 0 client got %v", tclients)
	}
}

func createTestFile(t testing.TB, testFile string) {
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatal("Failed to write test file:", err)
	}
}

func setLogsDevNull() error {
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		return fmt.Errorf("Failed to open devnull")
	}
	defer devNull.Close()
	log.SetOutput(devNull)

	return nil
}
