#!/bin/bash

cd ..
./build.sh
cd delivery

# Copy binary
mkdir -p rhood/cmd/rhood

mkdir -p rhood/data/
mkdir -p rhood/data/nginx
mkdir -p rhood/data/templates
mkdir -p rhood/data/youtube-dl

mkdir -p rhood/rhood-www/static/
mkdir -p rhood/rhood-www/static/cache/html
mkdir -p rhood/rhood-www/static/cache/video

# 1. Binary
cp ../cmd/rhood/rhood rhood/cmd/rhood
# 2. Templates
cp -r ../data/templates rhood/data
# 3. Youtube-dl
cp ../data/youtube-dl/youtube-dl rhood/data/youtube-dl/youtube-dl
# 4. nginx.conf
cp ../data/conf/examples/nginx.conf rhood/data/nginx/nginx.conf


# 5. Static http resources

cp -r ../rhood-www/static/css rhood/rhood-www/static
cp -r ../rhood-www/static/js rhood/rhood-www/static
cp -r ../rhood-www/static/video-js rhood/rhood-www/static

# 6. Run.sh

cp ../run.sh rhood

# 7. Readme

cp readme.txt rhood/readme.txt

# 8. Version
cd ..
git log -n 1 --format="%H" > delivery/rhood/version.txt
cd delivery


# 8. Pack the archive

zip -r rhood.zip rhood









