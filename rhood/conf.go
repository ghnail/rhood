package rhood

import (
	"github.com/elazarl/goproxy"
	ini "github.com/vaughan0/go-ini"
	"html/template"
	"os"
	"path/filepath"
)

var globalproxy *goproxy.ProxyHttpServer
var globalWSServer *WSServer

var globalCacheHtmlPage = NewRingCache(20)
var globalCacheLastMessages = NewLastMessages(20)

// Global variables
var globalCutoutState = CutoutState{}
var globalTemplatesAll = template.New("")

var _conf = getConfigDefaultValues()

func GetConfVal(name string) string {
	return _conf[name]
}

func getConfigDefaultValues() map[string]string {
	dirRoot, err := filepath.Abs((getCurrentFileDir() + "../"))
	if err != nil {
		logErr("Error with root dir: %s", err)
		dirRoot = "."
	}

	conf := map[string]string{
		"youtubeDownloader":       "/home/venv/rhood/bin/youtube-dl",
		"controlBoxListenAddress": "0.0.0.0:2000",
		//"controlBoxPublicAddress": "192.168.1.189:2000",
		"controlBoxPublicAddress": "localhost:2000",
		"goappProxyPort":          "8081",

		//		"dirStatic": "/home/venv/v1/www/flask/rhood_youtube/static",

		"dirRoot": dirRoot,
	}

	updateDirs(conf)

	return conf
}

func LoadConfig() map[string]string {
	_conf = loadProductionConfig()

	return _conf
}

func loadTestConfig() map[string]string {
	testConfigFile := getCurrentFileDir() + "../data/conf/test/rhood.test.conf"
	return getConfig(testConfigFile)
}

func loadProductionConfig() map[string]string {
	// If rhood.conf is in the launching binary dir, load it,
	// otherwise get default values
	dirOfBinary, err := filepath.Abs(filepath.Dir(os.Args[0]))

	confFile := ""
	if err == nil {
		confFile = dirOfBinary + "/rhood.conf"
	}

	return getConfig(confFile)
}

func getConfig(fileName string) map[string]string {
	resultConf := getConfigDefaultValues()

	if fileExists(fileName) {
		logDebug("Loading config from file %s", fileName)
		fileConf := getConfigFromFile(fileName)

		for key, value := range fileConf {
			resultConf[key] = value
		}
	} else {
		logInfo("No file config with name %s", fileName)
	}

	updateDirs(resultConf)

	return resultConf
}

func getConfigFromFile(fileName string) map[string]string {
	file, err := ini.LoadFile(fileName)

	if err != nil {
		logErr("Error loading file %s. Error: %s", fileName, err.Error())
		return make(map[string]string)
	}

	return file[""]
}

func updateDirs(conf map[string]string) {
	dirRoot := conf["dirRoot"]
	dirStatic := dirRoot + "/rhood-www/static"

	conf["dirStatic"] = dirStatic

	conf["dirStoreHtml"] = dirStatic + "/cache/html/"
	conf["dirStoreVideo"] = dirStatic + "/cache/video/"

	conf["dirTemplates"] = dirRoot + "/data/templates/"

}
