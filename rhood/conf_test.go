package rhood

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	configMap := loadTestConfig()
	result := mapToString(configMap)

	expected := "map[controlBoxListenAddress:192.168.1.1:80 controlBoxPublicAddress:192.168.1.1:8080 controlBoxPublicAddressWebsocket:localhost:2000 dirRoot:/var/www/rhood dirStatic:/var/www/rhood/rhood-www/static dirStoreHtml:/var/www/rhood/rhood-www/static/cache/html/ dirStoreVideo:/var/www/rhood/rhood-www/static/cache/video/ dirTemplates:/var/www/rhood/data/templates/ goappProxyPort:8082 youtubeDownloader:/usr/bin/youtube-dl]"

	assert.Equal(t, expected, result)
}
