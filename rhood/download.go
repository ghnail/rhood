package rhood

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
)

//=============================================================================
//		                  Interface functions
//=============================================================================

type DownloadRequest struct {
	URL     string
	Quality string
}

var newFilesToDownloadChan = make(chan DownloadRequest)
var cancelCurrentDownloadChan = make(chan string)

// TODO: download status: active/finished
// TODO: file to URL

func StartDownloadService() {
	downloadLoop()
}

func RequestDownload(url string) {
	go requestDownload(url, "")
}

func RequestDownloadWithQuality(url string, quality string) {
	go requestDownload(url, quality)
}

func CancelCurrentDownload() {
	go cancelCurrentDownload()
}

func requestDownload(url string, quality string) {
	req := DownloadRequest{URL: url, Quality: quality}
	newFilesToDownloadChan <- req
}

func cancelCurrentDownload() {
	timeoutSend(cancelCurrentDownloadChan, "cancel")
}

func downloadLoop() {
	for url := range newFilesToDownloadChan {
		doDownloadActual(url)
	}
}

//=============================================================================
//		              Download list of the available formats
//=============================================================================

func getFormOfAvailableFormats(youtubeUrl string) map[string]map[string]string {
	youtubeDownloader := GetConfVal("youtubeDownloader")

	cmd := exec.Command(youtubeDownloader, "--no-playlist", "-F", youtubeUrl)
	execCmd := NewExecCommand(cmd)

	execCmd.execCommandWithCancel()

	w := parseStatOutput(execCmd.totalStdout, execCmd.totalStderr, execCmd.err)

	selects := make(map[string]map[string]string)

	isChecked := true
	for _, row := range w {
		descr := toVideoDescription(row)

		name := descr[0]

		v := make(map[string]string)
		if isChecked {
			v["isChecked"] = "true"
		}
		v["format"] = descr[1]
		v["value"] = descr[2]
		selects[name] = v

		isChecked = false
	}

	return selects
}

func parseStatOutput(stdout string, stderr string, err error) []string {
	anchor := "format code extension resolution  note"

	if strings.Contains(stdout, anchor) {
		//		[info] Available formats for jTe0-uAdRdU:
		//		format code extension resolution  note
		//		171         webm      audio only  DASH audio , audio@ 48k (worst)
		//		160         mp4       144p        DASH video , video only

		tableWithFormats := regexp.
			MustCompile("(?s)^.*format code extension resolution  note").
			ReplaceAllString(stdout, "")

		logDebug(tableWithFormats)
		mp4Options := FilterString(strings.Split(tableWithFormats, "\n"), func(s string) bool {
			return strings.Contains(strings.ToLower(s), "mp4") &&
				!strings.Contains(strings.ToLower(s), "video only")
		})

		return mp4Options
	}

	if len(stderr) > 0 {
		logErr("%s", stderr)
	} else {
		logErr("Unknown error. Stdout: %s. Stderr: %s", stdout, stderr)
	}

	if err != nil {
		logErr("GO ERROR: %s", err.Error())
	}

	return []string{}
}

func toVideoDescription(s string) []string {
	res := make([]string, 3)

	splitted := regexp.MustCompile("\\s+").Split(s, -1)

	if len(splitted) < 3 {
		logDebug("Ignoring string %s", s)
		return make([]string, 0)
	}

	res[0], splitted = popHead(splitted)
	res[1], splitted = popHead(splitted)

	s = strings.Replace(s, res[0], "", -1)
	s = strings.Replace(s, res[1], "", -1)
	s = strings.TrimSpace(s)

	res[2] = s

	return res
}

func popHead(a []string) (string, []string) {
	return a[0], a[1:]
}

func FilterString(s []string, fn func(string) bool) []string {
	var p []string // == nil
	for _, v := range s {
		if fn(v) {
			p = append(p, v)
		}
	}
	return p
}

//=============================================================================
//						Download data from server
//=============================================================================

func getDownloaderCmd(url string, format string) *exec.Cmd {
	youtubeDownloader := GetConfVal("youtubeDownloader")
	dirStoreVideo := GetConfVal("dirStoreVideo")

	cmd := exec.Command(youtubeDownloader, "--no-playlist", "--newline", "--id", "-f", format, url)

	cmd.Dir = dirStoreVideo

	return cmd
}

func doDownloadActual(downloadRequest DownloadRequest) {

	url := downloadRequest.URL

	quality := downloadRequest.Quality
	if quality == "" {
		quality = "mp4"
	}

	cmd := getDownloaderCmd(url, quality)

	ecmd := NewExecCommand(cmd)
	ecmd.execCommandWithCancel()
}

//-----------------------------------------------------------------------------
//									Save HTML page

func htmlNameOfUrlShort(url string) string {
	fname := strings.Replace(url, "/", "_", -1)
	return fname
}

// The parameter is videoId, not the full url!
func videoFileFullName(videoId string) string {
	dirStoreVideo := GetConfVal("dirStoreVideo")
	videoFileName := dirStoreVideo + videoId + ".mp4"
	return videoFileName
}

func htmlNameOfUrlFull(url string) string {
	dirStoreHtml := GetConfVal("dirStoreHtml")
	filename := dirStoreHtml + "/" + htmlNameOfUrlShort(url) + ".html"

	return filename
}

func saveHtmlPage(youtubeUrl string) {
	// 1. Check youtube host, and video id

	urlMy, err := url.Parse(youtubeUrl)

	videoIdVals := urlMy.Query()["v"]

	isYoutubeUrl := regexp.MustCompile("^(www\\.)?youtube.com").MatchString(urlMy.Host)

	if !isYoutubeUrl || len(videoIdVals) < 0 {
		logErr("Can't determine videoId for url %s" + youtubeUrl)
		return
	}

	// 2. Download corresponding html page
	filename := htmlNameOfUrlFull(youtubeUrl)

	// 3. Try to read last user-viewed page to avoid 'choose language' message
	cached, isFromCache := globalCacheHtmlPage.Get(youtubeUrl)

	if isFromCache {
		logDebug("Cached video %s", youtubeUrl)
		buf := bytes.NewBufferString(cached)

		if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
			logErr("Can't save html page %s. Error: %s", youtubeUrl, err.Error())
		}
		return
	}

	logDebug("Not cached %s", youtubeUrl)

	// 4. Download corresponding html page

	resp, err := timeoutGet(youtubeUrl)

	if err != nil {
		logErr("Can't download html page %s. Error: %s", youtubeUrl, err.Error())
		return
	}

	// 5. Save page to disk

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	if err := ioutil.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		logErr("Can't save html page %s. Error: %s", youtubeUrl, err.Error())
	}
}

//=============================================================================
//							Messages to UI log
//=============================================================================

func jsonToString(jsonData []byte, err error) string {
	if err != nil {
		logErr("Error on serialization %s", err.Error())
		return "{}"
	}

	buf := bytes.NewBuffer(jsonData)

	return buf.String()
}

//=============================================================================
//								Exec utils
//=============================================================================

type ExecCommand struct {
	cmd *exec.Cmd

	err         error
	totalStdout string
	totalStderr string

	cancelChan  chan (string)
	isCancelled bool
}

func NewExecCommand(cmd *exec.Cmd) *ExecCommand {
	return &ExecCommand{cmd, nil, "", "", cancelCurrentDownloadChan, false}
}

func (execCommand *ExecCommand) sendToWebUI(s string) {
	if globalWSServer != nil {
		globalWSServer.Send(s)
	}

	logDebug("Message to UI: %s", s)
}

func (execCommand *ExecCommand) execCommandWithCancel() {
	// 1. Create process
	cmd := execCommand.cmd

	finished := executeAsync(cmd, execCommand.sendToWebUI)

	select {
	case res := <-finished:
		err := res["error"]
		stdout := res["stdout"]
		stderr := res["stderr"]

		if err == "" {
			logDebug("Process is finished correctly")
		} else {
			logErr("Process has failed. Error: %s", err)
		}

		execCommand.err = errors.New(err)
		execCommand.totalStdout = stdout
		execCommand.totalStderr = stderr

	case <-cancelCurrentDownloadChan:
		execCommand.isCancelled = true
		err := cmd.Process.Kill()
		if err == nil {
			logDebug("Process is cancelled")
		} else {
			logErr("Error cancelling process")
		}

		execCommand.totalStdout = ""
		execCommand.totalStderr = ""
		execCommand.err = errors.New("process is cancelled")
	}
}

//-----------------------------------------------------------------------------
//								Common methods

func readToString(reader io.Reader, eachLineCallback func(string)) string {
	res := bytes.NewBuffer(make([]byte, 0, 0))

	scanner := bufio.NewScanner(reader)
	isFirstRun := true
	for scanner.Scan() {
		s := scanner.Text()

		eachLineCallback(s)

		if isFirstRun {
			isFirstRun = false
		} else {
			s = s + "\n"
		}
		res.WriteString(s)
	}

	if err := scanner.Err(); err != nil {
		logErr("Error reading standard input: %s", err.Error())
	}

	return res.String()
}

// Returns map with keys:
// "error": error.Error() of from golang methods, if any
// "stdout": stdout messages as one string
// "stderr": stderr messages as one string
func executeAsync(cmd *exec.Cmd, eachLineCallback func(string)) (finished chan map[string]string) {
	// error, stdout, stderr
	finished = make(chan map[string]string, 1)
	result := map[string]string{"error": "", "stdout": "", "stderr": ""}

	// 1. Get stdout pipe
	stdoutPipe, err := cmd.StdoutPipe()

	if err != nil {
		result["error"] = err.Error()
		finished <- result
		return
	}

	// 2. Get stderr pipe
	stderrPipe, err := cmd.StderrPipe()

	if err != nil {
		result["error"] = err.Error()
		finished <- result
		return
	}

	// 3. Launch subprocess

	err = cmd.Start()

	if err != nil {
		result["error"] = err.Error()
		finished <- result
		return
	}

	// 4. Listen to command output in a separate goroutine

	go func() {
		stdout := readToString(stdoutPipe, eachLineCallback)
		stderr := readToString(stderrPipe, eachLineCallback)

		err := cmd.Wait()

		msg := ""
		if err != nil {
			msg = err.Error()
		}
		result["error"] = msg
		result["stdout"] = stdout
		result["stderr"] = stderr

		finished <- result

	}()

	return
}
