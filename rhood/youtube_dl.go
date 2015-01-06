package rhood

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"io/ioutil"
	"net/http"
	"io"
	"os"
	"time"
	"encoding/hex"
	"crypto/sha256"
	"errors"
	"strconv"
	"strings"
)

//=============================================================================
//				Downloading and updating of youtube-dl itself
//=============================================================================


//=============================================================================
//								File hash
//=============================================================================

func fileHash(filename string) (string, error) {
	//	"d03bb6c1e354566b4926f9115fdb45d7c41a4c58dc34d392cfe80578aafdbfc9"

	buf, err := ioutil.ReadFile(filename)
	if err != nil {return "", err}

	hash := sha256.Sum256(buf)

	result := hex.EncodeToString(hash[:])

	return result, nil
}




//=============================================================================
//				File transfer with progress notification
//=============================================================================



type ReaderWithProgressBar struct {
	io.Reader
	totalRead int64
	estimatedSize int64
	ProgressChan chan string
	// Stop, if 'received data' > 'estimatedSize'
	IsFallOnDataOverflow bool
}

func (rdr *ReaderWithProgressBar) GetPercentage() float64 {
	percentage := 100 * (float64(rdr.totalRead) / float64(rdr.estimatedSize))
	return percentage
}

func (rdr *ReaderWithProgressBar) FormatPercentage() string {
	res := ""
	if rdr.estimatedSize > 0 {
		percentage := rdr.GetPercentage()
		res = fmt.Sprintf("%.02f%%", percentage)
	} else {
		res = "???"
	}


	return res
}

func (rdr *ReaderWithProgressBar) Read(p []byte) (int, error) {
	n, err := rdr.Reader.Read(p)
	rdr.totalRead += int64(n)

	msg := ""

	if err == nil || err == io.EOF {
		// 1. Make notification about new data
		msg = fmt.Sprintf("Read %d of %d, %s",
			rdr.totalRead, rdr.estimatedSize, rdr.FormatPercentage())


		// 2. If receivedData > expectedData, and this error handler is turned
		// on, then terminate the process
		isDataOverflow := (rdr.totalRead > rdr.estimatedSize) && rdr.IsFallOnDataOverflow
		if isDataOverflow {
			msg = fmt.Sprintf("Data transmission is stopped. Expected: %d bytes, received: %d bytes",
				rdr.estimatedSize, rdr.totalRead)
			err = errors.New(msg)
		}
	}

	// 3. Change notification message to error, if required
	if err != nil && err != io.EOF {
		msg = "Error:" + err.Error()
	}

	// 4. Do the notification
	select {
	case rdr.ProgressChan <- msg:
		break
	case <-time.After(1 * time.Millisecond):
		//		default:
		// Single percent will cause log substitution problems
		msg = strings.Replace(msg, "%", "%%", -1)
		logErr("Can't send progress notification " + msg)
	}


	return n, err
}

func NewReaderWithProgressBar(originalReader io.Reader, size int64 ) *ReaderWithProgressBar {
	chn := make(chan string, 10)
	//	res := &Progress{Reader:originalReader, estimatedSize:size, }
	res := &ReaderWithProgressBar{originalReader, 0, size, chn, false}

	return res
}

func copyWithProgressBar(dst io.Writer, src io.Reader, size int64, notification chan string) error{
	// TODO: rename
	srcWithProgress := NewReaderWithProgressBar(src, size)

	go func() {
		// 'ProgressChan' is buffered chan, so it works as proxy
		for c := range srcWithProgress.ProgressChan {
			//			println("Notify ", c)
			notification <- c
		}
	} ()

	defer close(srcWithProgress.ProgressChan)

	_, err := io.Copy(dst, srcWithProgress); if err != nil {return err}

	//	<-time.After(1 * time.Second)


	return nil
}


//=============================================================================
//						Download/update logic
//=============================================================================



var _timeoutHttpClient = http.Client{Timeout: 30 * time.Second}


func downloadYoutubeDLIfRequiredWithConsoleOutput(youtubeDlFile string) error {
	ch := make(chan string)
//	chFinish := make(chan string)
	go func(){
		var isLogging bool
		for c := range ch {
			// Normally this message is not interesting
			isLogging =  c != "Finish. youtube-dl is in place"

			if isLogging {
				logDebug("Youtube-dl downloading progress: " + strings.Replace(c, "%", "%%", -1))
			}


			if strings.Contains(c, "100.00%") || strings.HasPrefix(c, "Error:") || strings.HasPrefix(c, "Finish."){
				break
			}
		}

		if isLogging {
			logDebug("Finished youtube-dl script downloading.")
		}

	}()

	err := downloadYoutubeDLIfRequired(youtubeDlFile, ch)

	return err
}

func downloadYoutubeDLIfRequired(youtubeDlFile string, progressChan chan string) error {
	if _, err := os.Stat(youtubeDlFile); os.IsNotExist(err) {
		logDebug("The downloader %s is missing. Let's try to download it.", youtubeDlFile)
		return downloadYoutubeDl(youtubeDlFile, progressChan)
	} else {
		go func() {
			progressChan <- "Finish. youtube-dl is in place"
		} ()

	}

	return nil
}





func downloadYoutubeDl(youtubeDlFile string, progressChan chan string) error {
//	sourceLink := "http://localhost:8000/youtube-dl"
	urlOfBinary := "https://yt-dl.org/downloads/2015.01.05.1/youtube-dl"
	expectedHash :=
	"f2275634b229d1443a25d836fb04751e999651deec4d3c84c42ed1c2dbb15c26"


	dir := filepath.Dir(youtubeDlFile)

	// 1, Prepare temp file at the target dir
	tmpFile, err := ioutil.TempFile(dir, "ytdl")
	defer func() {
		// If all is OK, the function closes and renames temp file.
		// But if the error occurred, and function doesn't finish it
		// ending, this defer will work and clean the situation.
		// The cleaning errors are ignored.
		tmpFile.Close()
		err = os.Remove(tmpFile.Name())
	}()

	if err != nil {return err}

	// 2. Launch download to temp file
	// TODO: implement 'wget -c' behaviour
	client := _timeoutHttpClient
	resp, err := client.Get(urlOfBinary)
	defer resp.Body.Close()
	//	println("Content length is " + resp.Header.Get("Content-Length"))
	if (err != nil) {return err}


	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	err = copyWithProgressBar(tmpFile, resp.Body, size, progressChan)

	// 3. Make sure we've downloaded right file
	err = tmpFile.Close(); if err != nil {return err}

	hash, err := fileHash(tmpFile.Name()); if err != nil {return err}



	if hash != expectedHash {
		return errors.New(fmt.Sprintf("Downloaded file hash is %s, expected %s",
			hash, expectedHash))
	}

	// 4. Make file executable, and move it from tmp-name to real-name

	err = os.Chmod(tmpFile.Name(), 0755); if err !=nil {return err}

	err = os.Rename(tmpFile.Name(), youtubeDlFile); if err != nil {return err}

	return nil
}

func updateYoutubeDl(youtubeDlFile string) {
	cmd := exec.Command(youtubeDlFile, "-U")

	execCmd := NewExecCommand(cmd)

	execCmd.execCommandWithCancel()

	w := parseStatOutput(execCmd.totalStdout, execCmd.totalStderr, execCmd.err)

	println(w)
}
