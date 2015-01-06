package rhood

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func getProxyTestHtml() string {
	html := `
<html>
<head>
</head>
<body>

<div id="yt-masthead-content">Here might be toolbar></div>
<div id="player-mole-container" class="somestuff">Hello, world! <div id="player-api" class="player-width player-height off-screen-target player-api"></div></div><div class="clear">Good day</div>

<body>
</html>`

	return html
}

func TestProxyVideoFromLanPlayerReplace(t *testing.T) {
	html := getProxyTestHtml()

	//	actual := updateYoutubeVideoPage("test_id", html, html, "www.youtube.com/watch?v=test_id")

	expected := `
<html>
<head>

<script src="http://localhost:1234/static/video-js/video.js"></script>
<link href="http://localhost:1234/static/video-js/video-js.css" rel="stylesheet">
<script>videojs.options.flash.swf = "http://localhost:1234/static/video-js/video-js.swf"</script>

<script src="http://localhost:1234/static/js/init-playlist-player.js"></script>
</head>
<body>

<div id="yt-masthead-content"><span style="color:green;">Video is cached</span>Here might be toolbar></div>
<div id="player_replaced" style="overflow:hidden;" class="player-width player-height off-screen-target player-api">
    <video
            id="example_video_1"
            class="video-js vjs-default-skin"
            controls autoplay loop preload="auto" width="100%" height="100%"
            poster="http://localhost:2000/static/video-js/oceans-clip.png"
            data-setup='{ "techOrder": ["flash", "html5"] }'>

        <source src="http://localhost:1234/static/cache/video/test_id.mp4" type='video/mp4' />

    </video>
</div><div class="clear">Good day</div>

<body>
</html>`

	isDisabledEntirePlayerDiv = true
	actual := videoFromLan(html, "test_id", "http://localhost:1234")

	html = html

	assert.Equal(t, expected, actual)
}

func TestProxyVideoFromLanPlayerReplaceWithHiding(t *testing.T) {
	html := getProxyTestHtml()

	//	actual := updateYoutubeVideoPage("test_id", html, html, "www.youtube.com/watch?v=test_id")

	expected := `
<html>
<head>

<script src="http://localhost:1234/static/video-js/video.js"></script>
<link href="http://localhost:1234/static/video-js/video-js.css" rel="stylesheet">
<script>videojs.options.flash.swf = "http://localhost:1234/static/video-js/video-js.swf"</script>

<script src="http://localhost:1234/static/js/init-playlist-player.js"></script>
</head>
<body>

<div id="yt-masthead-content"><span style="color:green;">Video is cached</span>Here might be toolbar></div>
<div id="player-mole-container" class="somestuff">Hello, world! <div id="player-api" class="player-width player-height off-screen-target player-api" style="display:none"></div></div><div id="player_replaced" style="overflow:hidden;" class="player-width player-height off-screen-target player-api">
    <video
            id="example_video_1"
            class="video-js vjs-default-skin"
            controls autoplay loop preload="auto" width="100%" height="100%"
            poster="http://localhost:2000/static/video-js/oceans-clip.png"
            data-setup='{ "techOrder": ["flash", "html5"] }'>

        <source src="http://localhost:1234/static/cache/video/test_id.mp4" type='video/mp4' />

    </video>
</div><div class="clear">Good day</div>

<body>
</html>`

	isDisabledEntirePlayerDiv = false
	actual := videoFromLan(html, "test_id", "http://localhost:1234")

	html = html

	assert.Equal(t, expected, actual)
}

func TestProxyVideoFromLanCacheButton(t *testing.T) {
	html := getProxyTestHtml()

	//	actual := updateYoutubeVideoPage("test_id", html, html, "www.youtube.com/watch?v=test_id")

	expected := `
<html>
<head>
</head>
<body>

<div id="yt-masthead-content"><a href="http://localhost:1234/admin/download/www.youtube.com/watch?v=test_id" target="_blank">Cache the video</a>Here might be toolbar></div>
<div id="player-mole-container" class="somestuff">Hello, world! <div id="player-api" class="player-width player-height off-screen-target player-api"></div></div><div class="clear">Good day</div>

<body>
</html>`

	isDisabledEntirePlayerDiv = true
	actual := addCacheButton(html, "www.youtube.com/watch?v=test_id", "http://localhost:1234")

	html = html

	assert.Equal(t, expected, actual)
}

func TestProxyVideoPreventAjax(t *testing.T) {
	html := getProxyTestHtml()

	//	actual := updateYoutubeVideoPage("test_id", html, html, "www.youtube.com/watch?v=test_id")

	expected := `
<html>
<head>

<script src="http://localhost:1234/static/js/vendor/jquery-1.11.0.min.js" type="text/javascript" name="jquery"></script>
<script src="http://localhost:1234/static/js/disable-ajax-video.js"></script></head>
<body>

<div id="yt-masthead-content">Here might be toolbar></div>
<div id="player-mole-container" class="somestuff">Hello, world! <div id="player-api" class="player-width player-height off-screen-target player-api"></div></div><div class="clear">Good day</div>

<body>
</html>`

	actual := preventYoutubeAjax(html, "http://localhost:1234")

	html = html

	assert.Equal(t, expected, actual)
}
