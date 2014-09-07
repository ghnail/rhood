function initRedirectToNextTrack($) {
	$(document).ready(function() {
		$('#example_video_1').each(function() {
			var myPlayer = videojs('example_video_1');

			var nextVideo = $('.next-playlist-list-item').attr('href');

			if (typeof nextVideo != 'undefined') {
				myPlayer.on('ended', function() {
					window.location.href = nextVideo;
				});
			}
		})
	});
}

initRedirectToNextTrack(jQuery);