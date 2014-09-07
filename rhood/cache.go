package rhood

import (
	"encoding/json"
	"github.com/golang/groupcache/lru"
	"html"
	"regexp"
	"strings"
	"sync"
)

//=============================================================================

// Enable/disable proxy
type CutoutState struct {
	isNoCache  bool
	isNoAccess bool
	mtx        sync.Mutex
}

func (c *CutoutState) isCacheDisabled() bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	return c.isNoCache
}

func (c *CutoutState) setCacheDisabled(state bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.isNoCache = state
	c.isNoAccess = false
}

func (c *CutoutState) isSiteDisabled() bool {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	return c.isNoAccess
}

func (c *CutoutState) setSiteDisabled(state bool) {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.isNoAccess = state

}

//=============================================================================

// Cache of youtube HTML pages
type RingCache struct {
	cache *lru.Cache
	mutex sync.Mutex
}

func (rc *RingCache) Put(url string, val string) {
	rc.mutex.Lock()

	rc.cache.Add(url, val)

	rc.mutex.Unlock()
}

func (rc *RingCache) Get(url string) (val string, isOk bool) {
	rc.mutex.Lock()

	valFromCache, isFound := rc.cache.Get(url)

	rc.mutex.Unlock()

	val, isString := valFromCache.(string)

	isValidData := isFound && isString

	if !isValidData {
		val = ""
	}

	return val, isValidData
}

func NewRingCache(size int) *RingCache {
	cache := lru.New(size)
	mutex := sync.Mutex{}

	return &RingCache{cache, mutex}
}

//=============================================================================

// Cache of the last log messages, at the 'status' page
type LastMessages struct {
	// seqNum => MessageWebUI
	cache *lru.Cache

	cacheSize int

	lastSeqNum int

	mutex *sync.Mutex
}

func NewLastMessages(size int) *LastMessages {
	res := LastMessages{}

	res.cacheSize = size
	res.cache = lru.New(size)

	res.lastSeqNum = 0
	res.mutex = &sync.Mutex{}

	return &res
}

func (lm *LastMessages) AddMessage(message *MessageWebUI) {
	lm.mutex.Lock()
	defer func() { lm.mutex.Unlock() }()

	message.SequenceNumber = lm.lastSeqNum
	lm.cache.Add(lm.lastSeqNum, message)

	lm.lastSeqNum += 1
}

func (lm *LastMessages) GetLastMessages() []*MessageWebUI {
	lm.mutex.Lock()
	defer func() { lm.mutex.Unlock() }()

	result := make([]*MessageWebUI, 0, 0)

	lastNum := lm.lastSeqNum
	startNum := lm.lastSeqNum - (lm.cacheSize - 1)

	for i := startNum; i != lastNum+1; i++ {
		dataFromCache, ok := lm.cache.Get(i)
		if !ok {
			continue
		} // no message; when cache is not filled yet

		if msg, ok := dataFromCache.(*MessageWebUI); ok {
			result = append(result, msg)
		} else {
			logErr("Unknown type of message %s", msg)
		}
	}

	return result
}

type MessageWebUI struct {
	// 	DOWNLOAD_STATUS DOWNLOAD_FINISHED RAW_STRING
	MessageType string

	SequenceNumber int

	DownloadPercent string
	DownloadSize    string
	DownloadSpeed   string
	ElapsedTime     string

	DownloadUrl string

	RawStringMessage string
}

func (message *MessageWebUI) SerializeJson() ([]byte, error) {
	val, err := json.Marshal(message)

	if err != nil {
		logErr("Can't serialize message %s. Error: %s", message, err.Error())
	}

	return val, err
}

// progressString := "[download]   8.1% of 8.71MiB at 230.42KiB/s ETA 00:35"
// progressString := "[download]   8.1% of 8.71MiB at 230.42KiB/s ETA 00:35 url www.youtube.com/watch?v=test"
func NewMessageWebUIFromString(status string) *MessageWebUI {
	status = html.EscapeString(status)

	message := &MessageWebUI{}
	message.MessageType = "RAW_STRING"
	message.RawStringMessage = status

	if !strings.HasPrefix(status, "[download]") {
		logDebug("Not a status info %s", status)

		return message
	}

	if strings.Contains(status, "has already been downloaded") {
		logDebug("Not a status but finish info %s", status)
		message.MessageType = "DOWNLOAD_FINISHED"

		components := regexp.MustCompile("\\s+").Split(status, -1)
		if len(components) >= 8 {
			message.DownloadUrl = components[7]
		}

		return message
	}

	// Process 'finish' message
	// [download] 100% of 3.18MiB in 00:30 youtube_url www.example.com
	isProgressFinish := regexp.MustCompile(" in [\\d:]+($| youtube_url.*$)").MatchString(status)
	if isProgressFinish {
		components := regexp.MustCompile("\\s+").Split(status, -1)

		if len(components) >= 8 {
			message.DownloadUrl = components[7]
		}

		message.MessageType = "DOWNLOAD_FINISHED"
		return message
	}

	components := regexp.MustCompile("\\s+").Split(status, -1)

	if len(components) < 8 {
		logErr("Can't parse string %s", status)
		return message
	}

	message.MessageType = "DOWNLOAD_STATUS"
	message.DownloadPercent = components[1]
	message.DownloadSize = components[3]
	message.DownloadSpeed = components[5]
	message.ElapsedTime = components[7]

	if len(components) >= 10 {
		message.DownloadUrl = components[9]
	}

	return message
}
