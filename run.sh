#!/usr/bin/env /bin/bash

# Set up default values, if var is empty
YOUTUBE_DL="${YOUTUBE_DL:-__APP_ROOT__/youtube-dl/youtube-dl/youtube-dl}"
BIND_PROXY="${BIND_PROXY:-0.0.0.0:8081}"
BIND_WEB="${BIND_WEB:-0.0.0.0:8090}"
PUBLIC_ADDRESS="${PUBLIC_ADDRESS:-localhost:8090}"
PUBLIC_ADDRESS_WS="${PUBLIC_ADDRESS_WS:-localhost:8090}"


# Set bin file
RHOOD_BIN="$PWD/cmd/rhood/rhood"

"$RHOOD_BIN" \
    --youtube-dl="$YOUTUBE_DL" \
    --bind-proxy="$BIND_PROXY" \
    --bind-web="$BIND_WEB" \
    --public-address="$PUBLIC_ADDRESS" \
    --public-address-ws="$PUBLIC_ADDRESS_WS"