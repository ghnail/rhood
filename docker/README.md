Same project as https://registry.hub.docker.com/u/ghnail/rhood_proxy/ ,
but with the automated build; so the readme is the same.

## Project

Proxy server to save Youtube videos, https://github.com/ghnail/rhood

## Setup

If the video save dir is /var/rhood

mkdir -p /var/rhood/video

mkdir -p /var/rhood/html

docker run -it -p 8090:8090 -p 8081:8081 -v /var/rhood:/data/rhood/cache ghnail/rhood