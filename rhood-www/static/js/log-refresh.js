function websocketLogRefresh($) {

    function initMessageConsumer() {
        var isLogAutoScrollEnabled = true;


        function scrollLogToBottom() {
            var textarea = document.getElementById('logArea');
            textarea.scrollTop = textarea.scrollHeight;
        }


        function showFlashConnectionError(message) {
            var $flash = $('#flashDiv')
            $flash.html(message);
            $flash.show()
        }

        function appendLogText(text) {
            var $element = $('#logArea');
            $element.each(function() {
                var originalText = $(this).text();
                $(this).text(originalText + text + '\n').html();
            });
        }


        function initWebsocket() {
            if (typeof WebSocket == 'undefined') {
                showFlashConnectionError('Websocket streaming is unavailable. You must refresh the page to get new data')
                return;
            }

            // globalControlBoxPublicAddress is set in the server-template file
            var ws = new WebSocket('ws://' + globalControlBoxPublicAddress +  '/admin/ws');






            ws.onopen = function(){
                ws.send(JSON.stringify('Hello!'))
            };
            ws.onmessage = function(message){
                var logData = JSON.parse(message.data);

                if (typeof logData.RawStringMessage == 'undefined') return;

                // 1. Place text to log window
                appendLogText(logData.RawStringMessage)

                if (isLogAutoScrollEnabled) {
                    scrollLogToBottom();
                }

                var $statusPre = $('#downloadStatus');

                var messageType = logData.MessageType;

                // 2. Set text of the download-info area
                if (messageType == 'DOWNLOAD_FINISHED') {
                    $statusPre.text(logData.RawStringMessage).html();
                } else if (messageType == 'DOWNLOAD_STATUS') {
                    $statusPre.text('Download status: ' + logData.RawStringMessage).html();
                }
            }

            ws.onclose = function() {
                showFlashConnectionError('Server streaming is disabled. You must refresh the page to get new data');
            }
        }


        function initAutoscroll() {
            $('#logArea').scroll(function(e) {
                var logAreaText = this;

                // When user viewing messages in the middle of text area, do not autoscroll on new elements.
                // If he

                if (logAreaText.offsetHeight + logAreaText.scrollTop >= logAreaText.scrollHeight) {
                    isLogAutoScrollEnabled = true
                } else {
                    isLogAutoScrollEnabled = false
                }
             });
        }



        scrollLogToBottom()
        initAutoscroll();
        initWebsocket();
    }


    $(document).ready(function() {
        initMessageConsumer();
    });
}

websocketLogRefresh(jQuery);