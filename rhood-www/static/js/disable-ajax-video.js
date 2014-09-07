function disableAjaxVideoLinks($) {
    function changeLinkListeners() {

        $('.related-video, .playlist-video, .yt-uix-sessionlink, .videowall-still, .next-playlist-list-item, .prev-playlist-list-item').each(function(e){
            $(this).unbind();
            $(this).on('click', function(e){
                if (e.which !== 1) return; // middle click guard
                window.location.href= $(this).attr('href');
                e.stopImmediatePropagation();
                e.preventDefault();
            })
        })
    }

    function setDomChangeListener() {
        MutationObserver = window.MutationObserver || window.WebKitMutationObserver;

        if (typeof MutationObserver !== 'undefined') {
            var observer = new MutationObserver(function(mutations, observer) {
                changeLinkListeners();
            });
            $('.branded-page-v2-primary-col, #watch7-sidebar, #watch7-sidebar-modules, #results, #page').each(
                function() {
                    observer.observe(this, {
                        subtree: true,
                        childList: true
                    });
                 })
        } else {
            setInterval(changeLinkListeners, 5000);
        }
    }
    // Ajax playlist is disabled, now page is reloaded
    // TODO: handle Flash backend

    function playNextVideo() {
    	var nextVideo = $('.next-playlist-list-item').attr('href');

    	if (nextVideo != 'undefined') {
    		window.location.href = nextVideo;
    	}
    }

    function setOriginalVideoElementHookNextTrack() {
    	$('.video-stream').each(function() {
    		$(this).unbind('ended');
    		$(this).on('ended', function() {
    			playNextVideo();
    		});
    	})
    }


    // Do the init


    $(document).ready(function() {
        setDomChangeListener();
        changeLinkListeners();
        setOriginalVideoElementHookNextTrack();
    })

}

disableAjaxVideoLinks(jQuery);