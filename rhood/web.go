package rhood

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"html/template"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

//=============================================================================
//								Admin methods
//=============================================================================

func cutoutActionGorillaRoute(w http.ResponseWriter, r *http.Request) {
	var err error = nil

	// 1. Save POST data

	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			logErr("Error on cutout. Error: %s", err.Error())
		}

		state := r.PostForm.Get("proxyState")

		switch state {
		case "cutout_pass":
			globalCutoutState.setCacheDisabled(true)
		case "cutout_cache":
			globalCutoutState.setCacheDisabled(false)
		case "cutout_all_off":
			globalCutoutState.setSiteDisabled(true)
		default:
			err = errors.New("Unknown post param: '" + state + "'")
		}
	}

	// 2.1. Prepare template

	// 2.2. With parameters
	params := make(map[string]interface{})
	params["page_title"] = "Proxy cutout"

	isCacheDisabled := globalCutoutState.isCacheDisabled()
	isSiteDisabled := globalCutoutState.isSiteDisabled()

	params["cutout_pass"] = isCacheDisabled && !isSiteDisabled
	params["cutout_cache"] = !isCacheDisabled && !isSiteDisabled
	params["cutout_all_off"] = isSiteDisabled

	if err != nil {
		params["flash_error"] = true
		logErr("Error on submit " + err.Error())
	}

	// 3. Response this template
	execTemplate(w, "cutout.html", params)
}

func downloadRequestActionGorillaRoute(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})
	params["page_title"] = "Download"

	if r.Method == "POST" {
		err := r.ParseForm()

		if err != nil {
			logErr("Error with tasks: " + err.Error())
		}
		logDebug("Request for download %s", r.PostForm.Get("youtube_link"))

		link := r.Form.Get("youtube_link")
		if link != "" {
			// TODO: is it required?
			html.EscapeString(link)

			logDebug("Youtube url to download %s", link)

			http.Redirect(w, r, "http://"+r.Host+"/admin/download/"+link, 303)
			return
		}

		params["flash_message"] = "Please, enter http address of the video"

	}

	execTemplate(w, "download-request.html", params)
}

func execTemplate(w http.ResponseWriter, templateName string, params interface{}) {
	if err := globalTemplatesAll.ExecuteTemplate(w, templateName, params); err != nil {
		logErr("Template error: " + err.Error())
	}
}

func statusActionGorillaRoute(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})
	params["initial_log_messages"] = "hello"

	if r.Method == "POST" {
		r.ParseForm()
		if r.Form.Get("cancel_task") == "cancel_task" {
			cancelCurrentDownload()
			params["flash_message"] = "The task is cancelled"
		}
	}

	messages := globalCacheLastMessages.GetLastMessages()

	logMessages := make([]string, len(messages), len(messages))
	lastDownloadState := "No downloads information"

	for i, msg := range messages {
		logMessages[i] = msg.RawStringMessage
		if msg.MessageType == "DOWNLOAD_FINISHED" || msg.MessageType == "DOWNLOAD_STATUS" {
			lastDownloadState = msg.RawStringMessage
		}

	}

	params["initial_log_messages"] = strings.Join(logMessages, "\n")
	params["download_status"] = lastDownloadState
	params["control_box_public_address"] = GetConfVal("controlBoxPublicAddress")
	params["control_box_public_address_websocket"] = GetConfVal("controlBoxPublicAddressWebsocket")

	execTemplate(w, "status.html", params)
}

func createCounterObject(nameToValue map[string]string) []string {
	counter := make([]string, len(_conf), len(_conf))
	i := 0
	for key, _ := range nameToValue {
		counter[i] = key
		i += 1
	}

	sort.Strings(counter)

	return counter
}

func oddityMap(names []string) map[string]bool {
	params := make(map[string]bool)

	for i, val := range names {
		params[val] = i%2 == 0
	}

	return params
}

func descriptionActionGorillaRoute(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})
	params["page_title"] = "Hello"

	execTemplate(w, "description.html", params)
}

func aboutActionGorillaRoute(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})
	params["config_options"] = _conf
	params["page_title"] = "About"

	params["counter"] = createCounterObject(_conf)
	params["oddity"] = oddityMap(createCounterObject(_conf))

	execTemplate(w, "about.html", params)
}

func inlineActionGorillaRoute(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})
	execTemplate(w, "inline.html", params)
}

func notFoundActionGorillaRoute(w http.ResponseWriter, r *http.Request) {
	params := make(map[string]interface{})
	execTemplate(w, "not-found-404.html", params)
}

func proxiedActionGorillaRoute(w http.ResponseWriter, r *http.Request) {
	youtubeUrl := "http://" + "www.youtube.com" + r.RequestURI

	processUrl(youtubeUrl, w, r)
}

func downloadActionGorillaRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		body, err := getRequestBody(r)
		err = err

		videoUrl := body.Get("video_url")
		videoQuality := body.Get("video_quality")

		params := make(map[string]string)
		params["video_url"] = videoUrl

		if videoUrl == "" {
		} else {
			youtubeUrl := videoUrl
			RequestDownloadWithQuality(youtubeUrl, videoQuality)
		}

		logDebug("Requested download for url %s with quality %s", videoUrl, videoQuality)

		execTemplate(w, "download-registered.html", params)
	} else { // GET
		realYoutubeUrl := strings.Replace(r.RequestURI, "/admin/download/", "", -1)

		logDebug("Preprocessing url %s", realYoutubeUrl)

		URL, _ := url.Parse(realYoutubeUrl)
		youtubeId := URL.Query().Get("v")

		logDebug("Getting formats for video with id %s", youtubeId)

		selects := getFormOfAvailableFormats(realYoutubeUrl)

		params := make(map[string]interface{})
		params["page_title"] = "Video download"

		params["video_url"] = realYoutubeUrl

		params["video_url"] = realYoutubeUrl
		params["downloadOptions"] = selects

		if len(selects) == 0 {
			params["is_not_available"] = true
		}

		execTemplate(w, "download.html", params)
	}
}

const DOWNLOAD_PAGE_PREFIX = "/admin/download/"

type DoubleSlashWorkaroundInterceptHandler struct{ router *mux.Router }

func (handler *DoubleSlashWorkaroundInterceptHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if strings.HasPrefix(request.URL.Path, DOWNLOAD_PAGE_PREFIX) {

		// Without this method the path
		// localhost/test/http://example.com will become
		// localhost/test/http:/example.com
		// it's bad, so we ignore cleanPath logic.
		downloadActionGorillaRoute(writer, request)
		return
	}

	handler.router.ServeHTTP(writer, request)
}

func RunGorillaMux() {
	dirStatic := GetConfVal("dirStatic")
	controlBoxListenAddress := GetConfVal("controlBoxListenAddress")

	r := mux.NewRouter()
	r.StrictSlash(true)
	r.PathPrefix("/static").
		Handler(http.StripPrefix("/static",
		http.FileServer(http.Dir(dirStatic))))

	r.HandleFunc("/", descriptionActionGorillaRoute).Methods("GET")

	r.HandleFunc("/admin/cutout", cutoutActionGorillaRoute).Methods("POST", "GET")
	r.HandleFunc("/admin/status", statusActionGorillaRoute).Methods("GET", "POST")
	r.HandleFunc("/admin/about", aboutActionGorillaRoute).Methods("GET")
	r.HandleFunc("/admin/inline", inlineActionGorillaRoute).Methods("GET")

	r.HandleFunc("/admin/download-request", downloadRequestActionGorillaRoute).Methods("POST", "GET")
	r.HandleFunc(DOWNLOAD_PAGE_PREFIX+"{url:.*}", downloadActionGorillaRoute).Methods("POST", "GET")

	r.HandleFunc("/watch", proxiedActionGorillaRoute).Methods("GET")

	r.HandleFunc("/youtube/{url:.*}", func(w http.ResponseWriter, r *http.Request) {
		youtubeUrl := strings.Replace(r.RequestURI, "/youtube/", "/", -1)
		youtubeUrl = "http://" + "www.youtube.com" + youtubeUrl

		processUrl(youtubeUrl, w, r)
	}).Methods("GET")

	wsServer := NewWSServer()
	globalWSServer = wsServer
	go wsServer.Start()

	r.HandleFunc("/admin/ws", wsServer.Serve)

	r.NotFoundHandler = http.HandlerFunc(notFoundActionGorillaRoute)

	initTemplates()

	http.ListenAndServe(controlBoxListenAddress, &DoubleSlashWorkaroundInterceptHandler{r})
}

type Counter struct {
	isOdd bool
}

func NewCounter() *Counter { return &Counter{false} }
func NextIsOdd(counter Counter) bool {
	counter.isOdd = !counter.isOdd
	fmt.Printf("%p; %s\n", &counter, counter.isOdd)
	return counter.isOdd
}

func initTemplates() {
	var odd_row = func(isOdd bool) bool {
		isOdd = !isOdd
		return isOdd
	}

	globalTemplatesAll = template.Must(template.New("").
		Funcs(template.FuncMap{"odd_row": odd_row}).
		Funcs(template.FuncMap{"NewCounter": NewCounter}).
		Funcs(template.FuncMap{"NextIsOdd": NextIsOdd}).
		ParseGlob(GetConfVal("dirTemplates") + "/*.html"))

	return
}

func getRequestBody(r *http.Request) (url.Values, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)

	postBody := buf.String()

	val, err := url.ParseQuery(postBody)

	return val, err
}

func getHttpClientThroughProxy() (*http.Client, error) {
	goappProxyPort := GetConfVal("goappProxyPort")

	proxyAddress := "http://localhost:" + goappProxyPort
	proxyUrl, err := url.Parse(proxyAddress)
	myClient := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)}}

	return myClient, err
}

func proxyWrite(respFromRemote *http.Response, w http.ResponseWriter) {

	for headerName, headerValues := range respFromRemote.Header {
		val := strings.Join(headerValues, "; ")
		w.Header().Set(headerName, val)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(respFromRemote.Body)

	w.WriteHeader(respFromRemote.StatusCode)
	w.Write(buf.Bytes())
}

//=======================       Direct link for youtube      ===============================

func processUrl(youtubeUrl string, w http.ResponseWriter, r *http.Request) {
	client, err := getHttpClientThroughProxy()

	logDebug("Processing url %s", youtubeUrl)

	if err != nil {
		errorPage(w, "Error with proxied http client.", err)
		return
	}

	// 2. Read response from server

	req, _ := http.NewRequest("GET", youtubeUrl, nil)

	hdr := make(http.Header)

	hdr["Proxy-Connection"] = []string{"keep=alive"}
	for k, v := range r.Header {
		hdr[k] = v
	}

	req.Header = hdr

	for _, cookie := range r.Cookies() {
		req.AddCookie(cookie)
	}

	response, err := client.Do(req)

	if err != nil {
		errorPage(w, "Error with http request. ", err)
		return
	}

	// 3. Copy it to response to user
	proxyWrite(response, w)
}

func errorPage(w http.ResponseWriter, logMessage string, err error) {
	if err != nil {
		logMessage += " " + err.Error()
	}

	w.WriteHeader(500)

	params := make(map[string]interface{})
	params["error_message"] = logMessage

	execTemplate(w, "error.html", params)
}
