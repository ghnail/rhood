package rhood

import (
	"github.com/elazarl/goproxy"
	"html/template"
	"os"
	"path/filepath"
	"flag"
	"net"
)

var globalproxy *goproxy.ProxyHttpServer
var globalWSServer *WSServer

var globalCacheHtmlPage = NewRingCache(20)
var globalCacheLastMessages = NewLastMessages(20)

// Global variables
var globalCutoutState = CutoutState{}
var globalTemplatesAll = template.New("")

// Values for tests, the app will override them by calling LoadConfig
var _conf = map[string]string{}


// For tests/app info; normally you should use GetConfVal instead.
func GetFullConfMap() map[string] string {
	return _conf;
}

func GetConfVal(name string) string {
	return _conf[name]
}


func LoadConfig() {
	//	_conf := make(map[string]string)
	//
	var youtubeDownloader = flag.String("youtube-dl", "/home/venv/rhood/bin/youtube-dl", "path to youtube-dl executable file")
	var goappProxyBindAddress = flag.String("bind-proxy", "0.0.0.0:8081", "bind address of proxy service")
	var controlBoxBindAddress = flag.String("bind-web", "0.0.0.0:2000", "bind address of web interface and file server")
	var controlBoxPublicAddress = flag.String("public-address", "localhost:2000", "from where web browser will request cached videos")
	var controlBoxPublicAddressWebsocket = flag.String("public-address-ws", "localhost:2000", "websocket addressto access admin streamin interface")

	flag.Parse()

	dirRoot := getRootDir()

	conf := map[string]string{
		"youtubeDownloader":       *youtubeDownloader,
		"controlBoxBindAddress": *controlBoxBindAddress,
		//"controlBoxPublicAddress": "192.168.1.189:2000",
		"controlBoxPublicAddress": *controlBoxPublicAddress,
		"controlBoxPublicAddressWebsocket": *controlBoxPublicAddressWebsocket,
		"goappProxyBindAddress":          *goappProxyBindAddress,

		"dirRoot": dirRoot,
	}

	updateConfig(conf)

	_conf = conf
}

func loadTestConfig() {
	conf := map[string]string{
		"youtubeDownloader":       "/home/venv/rhood/bin/youtube-dl",
		"controlBoxBindAddress": "0.0.0.0:2000",
		"controlBoxPublicAddress": "localhost:2000",
		"controlBoxPublicAddressWebsocket": "192.168.1.189:2000",
		"goappProxyBindAddress":          "0.0.0.0:8081",

		"dirRoot": getRootDir(),
	}

	updateConfig(conf)

	_conf = conf
}

func testIfRHoodRootDir(dir string) bool {
	templateDir := filepath.Join(dir, "data", "templates")

	stat, err := os.Stat(templateDir)
	return err == nil && stat.IsDir()
}

func getRootDir() string {
	rootDir := ""
	isRootDir := false
	invalidRootDirs := ""

	// 1. Look for root, referencing from running binary.

	// 'if' for the code similarity
	if (!isRootDir) {
		// Compiled version. the running binary file is <root>/cmd/rhood/rhood
		// so we need to go 2 levels up
		binaryDir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
		rootDir, _ = filepath.Abs(binaryDir + "/../../")
		isRootDir = testIfRHoodRootDir(rootDir)
	}

	if (!isRootDir) {invalidRootDirs += rootDir}



	// 2. Look for root, referencing from the source code file.
	if (!isRootDir) {
		// Test if development environment. The base file is <root>/rhood/conf.go
		// and we need go one level up.
		thisSourceFileDir := getCurrentFileDir()
		rootDir,_ = filepath.Abs(thisSourceFileDir + "/../")
		isRootDir = testIfRHoodRootDir(rootDir)
	}

	if (!isRootDir) {invalidRootDirs += ":" + rootDir}

	if (!isRootDir) {
		logFatal("Can't find root dir. It must contain 'data/templates' folder. Tested dirs are: %s", invalidRootDirs)
	}

	return rootDir
}


func updateConfig(conf map[string]string) {
	updateProxyPort(conf)
	updateDirs(conf)
}

func updateProxyPort(conf map[string]string) {
	proxyAddress := conf["goappProxyBindAddress"]


	_, port, err := net.SplitHostPort(proxyAddress)
	if (err != nil) {
		logFatal("Can't extract proxy port from %s. Error: %s", proxyAddress, err.Error())
	}


	conf["goappProxyPort"] = port
}


func updateDirs(conf map[string]string) {
	dirRoot := conf["dirRoot"]
	dirStatic := dirRoot + "/rhood-www/static"
//	dirStatic := "/home/venv/v1/www/flask/rhood_youtube/static/"

	conf["dirStatic"] = dirStatic

	conf["dirStoreHtml"] = dirStatic + "/cache/html/"
	conf["dirStoreVideo"] = dirStatic + "/cache/video/"

	conf["dirTemplates"] = dirRoot + "/data/templates/"

}
