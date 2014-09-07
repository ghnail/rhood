user www-data;
worker_processes 4;
pid /run/nginx.pid;

events {
	worker_connections 768;
}

http {
    resolver 127.0.0.1;
    #resolver 8.8.8.8;

    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    include    mime.types;

    # websocket
    map $http_upgrade $connection_upgrade {
        default upgrade;
        ''      close;
    }



    server {
        listen {{ control_box_public_address }};

        # websocket section
        location /admin/ws {
            proxy_pass http://{{ control_box_listen_address }};
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection $connection_upgrade;
        }



        location /static {
            #root /home/caching_proxy/static-data/;
            root {{ dir_rem_webserver }}/;
        }

        location / {

            #proxy_pass http://10.0.3.130:8090;
            proxy_pass http://{{ control_box_listen_address }};
        }


    }

    server {
        #listen 10.0.3.130:85;
        listen {{ nginx_proxy_address }} ;

        root {{ dir_rem_webserver }}/;

        # '=' is to exclude path '/watch_fragments_ajax?v=...'
        location = /watch {
            set $hst $host;

            if ($host ~* "^(www.)?youtube\.com$") {
                #set $hst 10.0.3.130:8090 ;
                set $hst {{ control_box_public_address }};
            }

            proxy_pass http://$hst;
        }
        location / {
            proxy_pass http://$http_host$request_uri;
        }
    }
}
