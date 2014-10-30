package rhood

import (
	"github.com/elazarl/goproxy"
	"github.com/elazarl/goproxy/ext/html"
	"regexp"

	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

func getVideoTag(videoUrl string) string {
	video :=
		`<div id="player_replaced" style="overflow:hidden;" class="player-width player-height off-screen-target player-api">
    <video
            id="example_video_1"
            class="video-js vjs-default-skin"
            controls autoplay loop preload="auto" width="100%" height="100%"
            poster="{{ img_url }}"
            data-setup='{ "techOrder": ["flash", "html5"] }'>

        <source src="{{ video_url }}" type='video/mp4' />

    </video>
</div>`

	controlBoxPublicAddress := GetConfVal("controlBoxPublicAddress")

	imgUrl := "http://" + controlBoxPublicAddress + "/static/video-js/oceans-clip.png"

	video = strings.Replace(video, "{{ video_url }}", videoUrl, -1)
	video = strings.Replace(video, "{{ img_url }}", imgUrl, -1)

	return video
}

// We can cut out the entire original player div while replacing it with
// the LAN player. This will 100% stop it, but will also raise JS
// exception, therefore comments will be unavailable.
// Or we can hide this div and corrupt the html5player http link,
// so it will be unable to play; but if another player is used,
// it will be hidden and still playing/requesting external data.

var isDisabledEntirePlayerDiv = false;

func videoFromLan(html string, videoId string, cacheBoxHttpAddress string) string {
	// 1. Replace video player tag

	mp4Url := fmt.Sprintf("%s/static/cache/video/%s.mp4", cacheBoxHttpAddress, videoId)

	videoTag := getVideoTag(mp4Url)

	if (isDisabledEntirePlayerDiv) {
		html = regexp.MustCompile(`(?s)<div id="player-mole-container.*<div class="clear"`).
		ReplaceAllString(html, videoTag+`<div class="clear"`)
	} else {
		// Can't simple remove #player-api, without it comments are not loaded.
		// JS catches NPE, and doesn't reach comments part.
		// But, with removed reference to video player, this div doesn't do anythin bad.
		html = regexp.MustCompile(`html5player`).
			ReplaceAllString(html, "replaced-stub-for-html-5-player")

		// Embed the new player
		html = regexp.MustCompile(`<div class="clear"`).
			ReplaceAllString(html, videoTag+`<div class="clear"`)

		html = regexp.MustCompile(`(<div id="player-api")([^>]*)(></div>)`).
			ReplaceAllString(html, `$1$2 style="display:none"$3`)
	}


	// 2. Add scripts/styles to support player
	playerHeader := `
<script src="{{ cache_box_address }}/static/video-js/video.js"></script>
<link href="{{ cache_box_address }}/static/video-js/video-js.css" rel="stylesheet">
<script>videojs.options.flash.swf = "{{ cache_box_address }}/static/video-js/video-js.swf"</script>

<script src="{{ cache_box_address }}/static/js/init-playlist-player.js"></script>
`
	playerHeader = strings.Replace(playerHeader, "{{ cache_box_address }}", cacheBoxHttpAddress, -1)

	anchor := `</head>`
	html = strings.Replace(html, anchor, playerHeader+anchor, 1)

	// 3. Label to notify about cached video
	doneMessage := `<span style="color:green;">Video is cached</span>`
	anchor = `<div id="yt-masthead-content">`
	html = strings.Replace(html, anchor, anchor+doneMessage, 1)

	return html
}

func addCacheButton(html string, fullVideoUrl string, cacheBoxHttpAddress string) string {
	actionLocation := fmt.Sprintf("%s/admin/download", cacheBoxHttpAddress)

	postForm := `<a href="%s/%s" target="_blank">Cache the video</a>`

	downloadAhref := fmt.Sprintf(postForm, actionLocation, fullVideoUrl)

	anchor := "<div id=\"yt-masthead-content\">"
	html = strings.Replace(html, anchor, anchor+downloadAhref, -1)

	return html
}

func saveHtmlIfRequired(videoFileName string, htmlFileName string, html string, reqUrl string) {
	if fileExists(videoFileName) && !fileExists(htmlFileName) {
		logDebug("Saving html page for url %s", reqUrl)
		// Write not-updated version of the page
		buf := bytes.NewBufferString(html)

		if err := ioutil.WriteFile(htmlFileName, buf.Bytes(), 0644); err != nil {
			logErr("Can't save html page %s. Error: %s", reqUrl, err.Error())
		}
	}
}

func updateYoutubeVideoPage(videoId string, html string, htmlOriginal string, reqUrl string) string {
	logDebug("Video id is '%s'", videoId)

	videoFileName := videoFileFullName(videoId)
	htmlFileName := htmlNameOfUrlFull(reqUrl)

	saveHtmlIfRequired(videoFileName, htmlFileName, htmlOriginal, reqUrl)

	logDebug("Video file is %s", videoFileName)

	cacheBoxHttpAddress := "http://" + GetConfVal("controlBoxPublicAddress")

	if fileExists(videoFileName) {
		html = videoFromLan(html, videoId, cacheBoxHttpAddress)
	} else {
		html = addCacheButton(html, reqUrl, cacheBoxHttpAddress)
	}

	return html
}

func preventYoutubeAjax(html string, controlBoxPublicAddressHttp string) string {
	scripts := `
<script src="{{ cache_box_address }}/static/js/vendor/jquery-1.11.0.min.js" type="text/javascript" name="jquery"></script>
<script src="{{ cache_box_address }}/static/js/disable-ajax-video.js"></script>`

	scripts = strings.Replace(scripts, "{{ cache_box_address }}", controlBoxPublicAddressHttp, -1)

	anchor := "</head>"

	html = strings.Replace(html, anchor, scripts+anchor, 1)

	return html
}

func logNetworkParams() {
	logInfo("Proxy port:\t\t\t\t" + GetConfVal("goappProxyPort"))
	logInfo("Admin address:\t\t\t" + GetConfVal("controlBoxListenAddress"))
	logInfo("Admin public address:\t" + GetConfVal("controlBoxPublicAddress"))
}

func Proxy() {
	var youtubeRegex = "^(www\\.)?youtube.com"

	controlBoxPublicAddressHttp := "http://" + GetConfVal("controlBoxPublicAddress")

	proxy := goproxy.NewProxyHttpServer()

	proxy.OnResponse(goproxy_html.IsHtml).Do(goproxy_html.HandleString(
		func(s string, ctx *goproxy.ProxyCtx) string {
			sOriginal := s

			// For youtube video page check out storage for *.mp4 file.
			// 1. If it was there, then update page to refer this file, not
			// the youtube original.
			// 2. If storage was empty, then add button 'store the video'
			// to the page.

			reqUrl := ctx.Req.URL
			videoId := reqUrl.Query().Get("v")

			isYoutube := regexp.MustCompile(youtubeRegex).MatchString(reqUrl.Host)

			if !isYoutube {
				return s
			}

			// 1. Handle enabled/disabled proxy

			if globalCutoutState.isSiteDisabled() {
				return "Site is disabled on the proxy"
			}

			if globalCutoutState.isCacheDisabled() {
				return s
			}

			// 2. Cache initial, not-updated version of the page.
			// We need video pages only, do not save nothing else.
			if videoId != "" {
				globalCacheHtmlPage.Put(reqUrl.String(), sOriginal)
			}

			// 3. Disable ajax-load

			// Append JS preventing AJAX youtube
			// For example, search page hasn't videos, nothing to update there.
			// But it loads data by ajax, and, when player appears on the screen,
			// we can't replace it with cached version.
			// So we need to update each youtube page; now all pages are loaded
			// through our proxy, and all cached videos are loaded from the LAN.
			s = preventYoutubeAjax(s, controlBoxPublicAddressHttp)

			if videoId == "" {
				return s
			}

			// 4. Update web page cache

			s = updateYoutubeVideoPage(videoId, s, sOriginal, reqUrl.String())

			return s
		}))

	proxy.
		OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(youtubeRegex))).
		DoFunc(

		func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			// Read HTML from disk, if it was saved
			// Mainly for the case when the video is deleted, but we have it in the cache
			videoId := req.URL.Query().Get("v")

			if videoId != "" {
				//				filename := dirStoreHtml + videoId + ".html"
				filename := htmlNameOfUrlFull(req.URL.String())

				if fileExists(filename) {
					logDebug("Read file %s", filename)
					html, err := ioutil.ReadFile(filename)
					if err == nil {
						logDebug("Found file for the url %s", req.URL.String())
						return req, goproxy.NewResponse(req, "text/html; charset=UTF-8", http.StatusOK,
							string(html))
					} else {
						logErr("Error reading file: %s", err.Error())
					}
				}
			}

			return req, nil
		})

	logInfo("Ready to start")

	goappProxyPort := GetConfVal("goappProxyPort")

	proxy.Verbose = true

	globalproxy = proxy

	logNetworkParams()


	log.Fatal(http.ListenAndServe(":"+goappProxyPort, proxy))
}
