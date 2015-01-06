package rhood

import (
	"fmt"
	logging "github.com/op/go-logging"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

var log *logging.Logger

//var format = "%{color}%{time:15:04:05} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}"
var format = "%{color}%{time:15:04:05} %{level:.4s} %{id:03x}%{color:reset} %{shortfile} ▶  %{message}"

// If you want to log the 'file:line' info too, use prev format line.
// But you will have to replace two calls of 'backend.Log' in the github.com/op/go-logging/logger.go file.
// Replace
// l.backend.Log(lvl, 2, record)
// with
// l.backend.Log(lvl, 3, record)
//
// Same for defaultBackend.Log(lvl, 2, record)

var _ = initLogging()

func initLogging() bool {
	var logBackend = logging.NewLogBackend(os.Stdout, "", 0)
	logging.SetBackend(logBackend)
	logging.SetFormatter(logging.MustStringFormatter(format))

	log = logging.MustGetLogger("example")

	return true
}

//=============================================================================
//							Log methods
//=============================================================================

func logDebug(format string, args ...interface{}) {
	log.Debug(format, args...)
}

func logInfo(format string, args ...interface{}) {
	log.Info(format, args...)
}
func logErr(format string, args ...interface{}) {
	log.Error(format, args...)
}

func logFatal(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args)
	log.Fatal(message)
}

func println(a ...interface{}) {
	//	_, file, line, _ := runtime.Caller(1)
	//	fmt.Printf("(%s:%d) %s\n", file, line, a)
	runtime.Caller(1)
	fmt.Println(a)
}

//=============================================================================
//							Other methods
//=============================================================================

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func timeoutSend(channel chan string, message string) bool {
	return timeoutSendT(channel, message, 1000)
}
func timeoutSendT(channel chan string, message string, timeout int64) bool {
	select {
	case channel <- message:
		return true
	case <-time.After(time.Duration(timeout) * time.Millisecond):
		return false
	}
}

// ordered version; same map always has same string
func mapToString(m map[string]string) string {
	allEntries := make([]string, 0, len(m))

	for key, value := range m {
		allEntries = append(allEntries, key+":"+value)
	}

	sort.Strings(allEntries)

	result := "map[" + strings.Join(allEntries, " ") + "]"

	return result
}

func dialTimeout(network, addr string) (net.Conn, error) {
	var timeout = time.Duration(15 * time.Second)
	return net.DialTimeout(network, addr, timeout)
}

func timeoutGet(url string) (resp *http.Response, err error) {
	transport := http.Transport{
		Dial: dialTimeout,
	}

	client := http.Client{
		Transport: &transport,
	}

	return client.Get(url)
}

func getCurrentFileDir() string {

	_, callerSourceFile, _, _ := runtime.Caller(1)
	dir := filepath.Dir(callerSourceFile) + string(filepath.Separator)

	return dir
}
