package rhood

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Root dir, /home/user/ , is build-dependent. It might be different in your
	// environment, so the test is ignored.
	t.Skip()
	loadTestConfig()

	configMap := GetFullConfMap()
	result := mapToString(configMap)

	expected := "map[controlBoxBindAddress:0.0.0.0:2000 controlBoxPublicAddress:localhost:2000 controlBoxPublicAddressWebsocket:192.168.1.189:2000 dirRoot:/home/user/gocode/src/github.com/ghnail/rhood dirStatic:/home/user/gocode/src/github.com/ghnail/rhood/rhood-www/static dirStoreHtml:/home/user/gocode/src/github.com/ghnail/rhood/rhood-www/static/cache/html/ dirStoreVideo:/home/user/gocode/src/github.com/ghnail/rhood/rhood-www/static/cache/video/ dirTemplates:/home/user/gocode/src/github.com/ghnail/rhood/data/templates/ goappProxyBindAddress:0.0.0.0:8081 youtubeDownloader:/home/venv/rhood/bin/youtube-dl]"

	assert.Equal(t, expected, result)
}
