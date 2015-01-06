package rhood

import (
	"fmt"
	//	"os"
	//	"net/http"
	//	"io"
	"os/exec"
	//	"io/ioutil"
	"path/filepath"
	"io/ioutil"
	"net/http"
	"io"
	//	"os"
	"os"
	"time"
	//	"crypto/sha512"
	"encoding/hex"
	//	"reflect"
	"crypto/sha256"
	"errors"
	"strconv"
	"strings"
)
//func cryptoTesting() {
//	fl := "/home/z/gocode/src/github.com/ghnail/rhood/data/youtube-dl/youtube-dl"
//
//	println(fileHash(fl))
////	// TODO: http get, limit number of bytes
//}




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



type Progress struct {
	io.Reader
	totalRead int64
	estimatedSize int64
	ProgressChan chan string
	// Stop, if received data > estimatedSize
	IsLimitData bool
}

func (pb *Progress) GetPercentage() float64 {
	percentage := 100 * (float64(pb.totalRead) / float64(pb.estimatedSize))
	return percentage
}

func (pb *Progress) FormatPercentage() string {
	res := ""
	if pb.estimatedSize > 0 {
		percentage := pb.GetPercentage()
		res = fmt.Sprintf("%.02f%%", percentage)
	} else {
		res = "???"
	}


	return res
}

func (pb *Progress) Read(p []byte) (int, error) {
	n, err := pb.Reader.Read(p)
	pb.totalRead += int64(n)

	msg := ""

	if err == nil || err == io.EOF {
		// 1. Make notification about new data
		msg = fmt.Sprintf("Read %d of %d, %s",
			pb.totalRead, pb.estimatedSize, pb.FormatPercentage())


		// 2. If receivedData > expectedData, and this error is turned
		// on, then terminate the process
		isDataOverflow := pb.totalRead > pb.estimatedSize && pb.IsLimitData
		if isDataOverflow {
			msg = fmt.Sprintf("Data transmission is stopped. Expected: %d bytes, received: %d bytes",
				pb.estimatedSize, pb.totalRead)
			err = errors.New(msg)
		}
	}

	// 3. Change notification message to error, if required
	if err != nil && err != io.EOF {
		msg = "Error:" + err.Error()
	}

	// 4. Do the notification
	select {
	case pb.ProgressChan <- msg:
		break
	case <-time.After(1 * time.Millisecond):
		//		default:
		// Single percent will cause log substitution problems
		msg = strings.Replace(msg, "%", "%%", -1)
		logErr("Can't send progress notification " + msg)
	}


	return n, err
}

func NewProgressReader(originalReader io.Reader, size int64 ) *Progress{
	chn := make(chan string, 10)
	//	res := &Progress{Reader:originalReader, estimatedSize:size, }
	res := &Progress{originalReader, 0, size, chn, false}

	return res
}

func copyWithProgressBar(dst io.Writer, src io.Reader, size int64, notification chan string) error{
	// TODO: rename
	srcProgress := NewProgressReader(src, size)

	go func() {
		// 'ProgressChan' is buffered chan, so it works as proxy
		for c := range srcProgress.ProgressChan {
			//			println("Notify ", c)
			notification <- c
		}
	} ()

	defer close(srcProgress.ProgressChan)

	_, err := io.Copy(dst, srcProgress); if err != nil {return err}

	//	<-time.After(1 * time.Second)


	return nil

}





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

//	println("Downloading")
//	<- chFinish
//	println("Donw")

	return err
//	if _, err := os.Stat(youtubeDlFile); os.IsNotExist(err) {
//		logDebug("Youtube-dl is not in place. Let's try to download it.")
//		return downloadYoutubeDl(youtubeDlFile, progressChan)
//	}
//
//	return nil
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



var sourceLink = "http://localhost:8000/youtube-dl"





//// PassThru wraps an existing io.Reader.
////
//// It simply forwards the Read() call, while displaying
//// the results from individual calls to it.
//type PassThru struct {
//	io.Reader
//	total int64 // Total # of bytes transferred
//	EstimatedSize int64
//	ProgressChan chan string
//}
//
//// Read 'overrides' the underlying io.Reader's Read method.
//// This is the one that will be called by io.Copy(). We simply
//// use it to keep track of byte counts and then forward the call.
//func (pt *PassThru) Read(p []byte) (int, error) {
//	n, err := pt.Reader.Read(p)
//	pt.total += int64(n)
//
//	if err == nil {
//		fmt.Println("Read", n, "bytes for a total of", pt.total)
//	}
//
//	return n, err
//}
//
//
//func testBufferedChannel() {
//
//}




//func copyWithProgressBarExample() {
//	var w bytes.Buffer
//	content := strings.Repeat("123456789", 1000)
//	r := bytes.NewBufferString(content)
//
//	ch := make(chan string)
//	chFinish := make(chan string)
//
//	go func() {
//		for c := range ch {
//			println("Chan", c)
//			println(strings.Contains(c, "100.00%"))
//			if strings.Contains(c, "100.00%") {
//				break
//			}
//		}
//
//		chFinish <- "done"
//
//	} ()
//
//	println("Copy is started")
//
//	err := copyWithProgressBar(&w, r, int64(len(content)), ch)
//
//	println("Copy is finished", err)
//
//	<-chFinish
//	println("Done")
////	<-time.After(100 * time.Millisecond)
//}




func downloadFileWithProgressBar() {
	//	copyWithProgressBarExample()
	//	return
	client := _timeoutHttpClient
	resp, err := client.Get(sourceLink)
	defer resp.Body.Close()

	if err != nil {return}

	println("Content length is " + resp.Header.Get("Content-Length"))


	//	src := &PassThru{Reader: resp.Body}
	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)

	os.Remove("/tmp/ytdl.py")



	tmpFile, err := os.OpenFile("/tmp/ytdl.py", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	defer tmpFile.Close()

	src := resp.Body

	ch := make(chan string)

	go func(){
		for c:= range ch {
			println("CH", c)
		}
	}()

	copyWithProgressBar(tmpFile, src, size, ch)

	time.After(1 * time.Millisecond)

}


//
//func copyWithProgressBar123() {
//
//
//
//
//	if (true) {return}
////	testBufferedChannel()
////	if (true) {return}
//	client := timeoutHttpClient
//	resp, err := client.Get(sourceLink)
//	defer resp.Body.Close()
//
//	if err != nil {return}
//
//	println("Content length is " + resp.Header.Get("Content-Length"))
//
//
////	src := &PassThru{Reader: resp.Body}
//	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
//	src := NewProgressReader(resp.Body, size)
////	size = size
////	src := NewProgressReader(resp.Body, 0)
//
//
//
////	f, err = os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
//
//	go func() {
//		for c := range src.ProgressChan {
//			println("From chan", c)
//		}
//	} ()
//
//	os.Remove("/tmp/ytdl.py")
//
//
//
//
//	tmpFile, err := os.OpenFile("/tmp/ytdl.py", os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
//	defer tmpFile.Close()
//
//	_, err = io.Copy(tmpFile, src); if err != nil {return}
//
//	println("Total read ", src.totalRead)
//
//
////	<-time.After(1 * time.Second)
//
//}











func downloadYoutubeDl(youtubeDlFile string, progressChan chan string) error {
	sourceLink := "http://localhost:8000/youtube-dl"
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
	resp, err := client.Get(sourceLink)
	defer resp.Body.Close()
	//	println("Content length is " + resp.Header.Get("Content-Length"))
	if (err != nil) {return err}


	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	err = copyWithProgressBar(tmpFile, resp.Body, size, progressChan)

	// 3. Make sure we've downloaded right file
	err = tmpFile.Close(); if err != nil {return err}

	hash, err := fileHash(tmpFile.Name()); if err != nil {return err}


	expectedHash :=
	"d03bb6c1e354566b4926f9115fdb45d7c41a4c58dc34d392cfe80578aafdbfc9"

	if hash != expectedHash {
		return errors.New(fmt.Sprintf("Downloaded file hash is %s, expected %s",
			hash, expectedHash))
	}

	// 4. Make file executable, and move it from tmp-name to real-name

	err = os.Chmod(tmpFile.Name(), 0755); if err !=nil {return err}

	err = os.Rename(tmpFile.Name(), youtubeDlFile); if err != nil {return err}

	return nil
}



//func downloadYoutubeDl(youtubeDlFile string) error {
//	sourceLink := "http://localhost:8000/youtube-dl"
//	dir := filepath.Dir(youtubeDlFile)
//
//	// 1, Prepare temp file at the target dir
//	tmpFile, err := ioutil.TempFile(dir, "ytdl")
//	defer func() {
//		// If all is OK, the function closes and renames temp file.
//		// But if the error occurred, and function doesn't finish it
//		// ending, this defer will work and clean the situation.
//		// The cleaning errors are ignored.
//		tmpFile.Close()
//		err = os.Remove(tmpFile.Name())
//	}()
//
//	if err != nil {return err}
//
//	// 2. Launch download to temp file
//	// TODO: implement 'wget -c' behaviour
//	client := _timeoutHttpClient
//	resp, err := client.Get(sourceLink)
//	defer resp.Body.Close()
////	println("Content length is " + resp.Header.Get("Content-Length"))
//	if (err != nil) {return err}
//
//	_, err = io.Copy(tmpFile, resp.Body); if err != nil {return err}
//
//	// 3. Make sure we've downloaded right file
//	err = tmpFile.Close(); if err != nil {return err}
//	hash, err := fileHash(tmpFile.Name()); if err != nil {return err}
//
//
//	expectedHash :=
//		"d03bb6c1e354566b4926f9115fdb45d7c41a4c58dc34d392cfe80578aafdbfc9"
//
//	if hash != expectedHash {
//		return errors.New(fmt.Sprintf("Downloaded file hash is %s, expected %s",
//			hash, expectedHash))
//	}
//
//	// 4. Make file executable, and move it from tmp-name to real-name
//
//	err = os.Chmod(tmpFile.Name(), 0755); if err !=nil {return err}
//
//	err = os.Rename(tmpFile.Name(), youtubeDlFile); if err != nil {return err}
//
//	return nil
//}

func updateYoutubeDl(youtubeDlFile string) {
	cmd := exec.Command(youtubeDlFile, "-U")

	execCmd := NewExecCommand(cmd)

	execCmd.execCommandWithCancel()

	w := parseStatOutput(execCmd.totalStdout, execCmd.totalStderr, execCmd.err)

	println(w)
}
