package rhood

import (
	"testing"
	"io/ioutil"
	"os"
	"time"
	"strings"
	"bytes"
	"github.com/stretchr/testify/assert"
)


func TestCopyWithProgress(t *testing.T) {
	// 1. Prepare
	var w bytes.Buffer
	content := strings.Repeat("123456789", 1000)
	r := bytes.NewBufferString(content)

	ch := make(chan string)
	chFinish := make(chan string)

	go func() {
		for c := range ch {
//			println(c)
			if strings.Contains(c, "100.00%") || strings.HasPrefix(c, "Error:"){

				break
			}
		}

		chFinish <- "done"

	} ()



	// 2. Do the work
	err := copyWithProgressBar(&w, r, int64(len(content)), ch)

	// 3. Assert
	if err != nil {
		assert.Fail(t, err.Error())
	}

	select {
	case <-chFinish:
		// all is OK
		break
	case <- time.After(100 * time.Millisecond):
		assert.Fail(t, "Copying is not finished in time")
	}
}


func TestFileHash(t *testing.T) {
	// 1. Prepare
	f,_ := ioutil.TempFile(".", "hash-test")

	defer func(){
		f.Close();
		os.Remove(f.Name())
	}()

	f.WriteString("What is my hash?")
	f.Close()

	// 2. Do the work
	res,_ := fileHash(f.Name())


	// 3. Assert
	expected := "64fbe68343d374df7e7e588b1754271eb2134367fb12d96720a7d7ed1a3fb596"
	assert.Equal(t, expected, res)
}

func TestFirst(t *testing.T) {
//	t.Skip("Test requires Internet connection")
//	ch := make(chan string,1000)
	ch := make(chan string)

	file, _ := ioutil.TempFile(".", "test-youtube-dl")
	file.Close()
	tmpName := file.Name()


	defer func() {
		file.Close()
		os.Remove(tmpName)
	}()


	err := downloadYoutubeDl(tmpName, ch);

	assert.Nil(t, err)

	stat, _ := os.Stat(tmpName)

	// Or use hash value
	assert.Equal(t, 723782, stat.Size())


}
