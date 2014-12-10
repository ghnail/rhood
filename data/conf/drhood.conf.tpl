description "rhood daemon"

start on (filesystem)
stop on runlevel [016]

respawn
console log
setuid www-data
setgid www-data

chdir {{ dir_rem_webserver }}

script
    {{ dir_rem_bin }}/rhood --youtube-dl="{{ file_rem_youtube_dl_bin }}"  \
    --bind-proxy="{{ goapp_proxy_bind_address }}" \
    --bind-web="{{ control_box_listen_address }}" \
    --public-address="{{ control_box_public_address }}" \
    --public-address-ws="{{ control_box_public_address_websocket }}"

end script