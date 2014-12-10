package rhood

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"regexp"
	"testing"
	"time"
)

func TestAsyncExecSimple(t *testing.T) {
	cmd := exec.Command("echo", "-ne", "Hello, world!")
	finish := executeAsync(cmd, func(s string) {})

	done := <-finish
	expected := "Hello, world!"
	actual := done["stdout"]

	assert.Equal(t, expected, actual)
}

func TestAsyncExecCancel(t *testing.T) {
	cmd := exec.Command("sleep", "1")

	startTime := time.Now()

	finish := executeAsync(cmd, func(s string) {})
	cmd.Process.Kill()

	done := <-finish
	stopTime := time.Now()
	expected := "signal: killed"
	actual := done["error"]

	// valid message
	assert.Equal(t, expected, actual)
	// and process is finished before cmd has slept full time
	assert.WithinDuration(t, stopTime, startTime, time.Millisecond*100)
}

func TestIntDownloadStartAndCancel(t *testing.T) {
	loadTestConfig()
	go StartDownloadService()
	go func() {
		time.Sleep(1 * time.Second)

		logDebug("Download is requested")
		requestDownload("http://youtube.com/watch?v=cuq_y8Ugf5g", "")
		//		newFilesToDownloadChan <- "http://youtube.com/watch?v=5-2ThFddglk"

		logDebug("Message is sent")
	}()

	go func() {
		time.Sleep(2 * time.Second)

		cancelCurrentDownloadChan <- "cancel"

		logDebug("Cancelled")
	}()

	go downloadLoop()

	expected := `[Download is requested]
[Message is sent]
Message to UI: [youtube] Setting language
[Cancelled]
[Process is cancelled]

`

	logDebug("There should be messages\n" + expected)

	time.Sleep(3 * time.Second)
}

func TestVideoFormatHtmlFormRender(t *testing.T) {
	// we need dirTemplate conf argument
	loadTestConfig()

	//	map[18:map[isChecked:true format:mp4 value:640x360] 22:map[format:mp4 value:1280x720    (best)]]
	selects := make(map[string]map[string]string)
	selects["18"] = map[string]string{"isChecked": "true", "format": "mp4", "value": "640x360"}

	params := make(map[string]interface{})
	params["downloadOptions"] = selects

	expected := `<div id="form-div" style="">



<form class="pure-form pure-form-aligned"  method="post" action="/admin/download/">
    Please, select the required video quality

        <label for="18" class="pure-radio">
            <input id="18" type="radio" name="video_quality" value="18" checked>
            640x360
        </label>

    <input type="hidden" name="video_url" value="">
    <button type="submit" class="pure-button form-button">Download</button>
</form>
</div>

`

	//	globalTemplatesAll.ExecuteTemplate(os.Stdout, "admin/download-get.html", params)
	buf := new(bytes.Buffer)
	initTemplates()
	globalTemplatesAll.ExecuteTemplate(buf, "download-form-format.html", params)

	actual := buf.String()

	// Otherwise we have whitespace problems
	expected = regexp.MustCompile("\\s+").ReplaceAllString(expected, "")
	actual = regexp.MustCompile("\\s+").ReplaceAllString(actual, "")

	assert.Equal(t, expected, actual)
}

func TestHtmlDownload(t *testing.T) {
	// "http://youtube.com/watch?v=cuq_y8Ugf5g"
	//
	//	saveHtmlPage(youtubeUrl)
}

func TestStringParsingToMessage(t *testing.T) {
	progressString := "[download]   8.1% of 8.71MiB at 230.42KiB/s ETA 00:35"

	expected := `{"MessageType":"DOWNLOAD_STATUS","SequenceNumber":0,"DownloadPercent":"8.1%","DownloadSize":"8.71MiB","DownloadSpeed":"230.42KiB/s","ElapsedTime":"00:35","DownloadUrl":"","RawStringMessage":"[download]   8.1% of 8.71MiB at 230.42KiB/s ETA 00:35"}`

	message := NewMessageWebUIFromString(progressString)
	actual := jsonToString(message.SerializeJson())

	assert.Equal(t, expected, actual)
}

func TestStringParsingFinishToMessage(t *testing.T) {
	progressString := "[download] 100% of 3.18MiB in 00:30 youtube_url www.example.com"

	expected := `{"MessageType":"DOWNLOAD_FINISHED","SequenceNumber":0,"DownloadPercent":"","DownloadSize":"","DownloadSpeed":"","ElapsedTime":"","DownloadUrl":"www.example.com","RawStringMessage":"[download] 100% of 3.18MiB in 00:30 youtube_url www.example.com"}`

	message := NewMessageWebUIFromString(progressString)
	actual := jsonToString(message.SerializeJson())

	assert.Equal(t, expected, actual)
}
