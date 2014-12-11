Robin Y Hood proxy

# 1. Description:

This is proxy server for Youtube. It adds to all video pages a new button to cache the data:

![Cache button](https://raw.githubusercontent.com/ghnail/rhood/master/doc/img/cache-button.png)

When you press it, the new dialog will appear to select video quality.

![Download options](https://raw.githubusercontent.com/ghnail/rhood/master/doc/img/download-options.png)

And the next time you visit video page, you will get media traffic from the local server:

![Cached video](https://raw.githubusercontent.com/ghnail/rhood/master/doc/img/cached-video.png)

The video player is changed, and we see a green text: "Video is cached".

# 2. How it works

Sometimes (hello, slow networks!) it may be nice to work with a regular youtube
interface (http address, user comments, playlists, recommendations), but with a
video downloaded from the LAN server.

The first idea: cache the media data on Squid. But it doesn't work.
Youtube has complicated player, which works with a lot of video chunks instead of single video file.
And each of them can have dynamic name for the same content, so Squid can't handle them.

But there is a program, youtube-dl, which can download Youtube video to local file.
Now it can be played with any offline media player, like VLC or mplayer.
We can also save it on server, and feed browser with it.

Original Youtube player can't request LAN server, so we need to replace
it with another software, in this app it is VideoJS.

The main disadvantage is that we have lost annotations and subtitles,
but for the most of videos it's not that huge loss.

# 3 Install

You can download the distribution, use Docker image, or build
the app by yourself.

## 3.1. Distribution

The downloader requires Python installed on your system.
If it's missing, please, install it.

```bash
sudi apt-get install python
```


Then you can download the last *.zip file from the Releases tab
(for example, https://github.com/ghnail/rhood/releases/tag/0.02 ),
unzip it and launch ./run.sh script. Be careful, you must be in
the rhood root directory; the script is not path-independent for now.

Visit http://localhost:8090 to see, if the page is opened.
If it's OK, you can visit some video

http://localhost:8090/youtube/watch?v=UU5wFUqoBbk

Or set up the HTTP proxy localhost:8081,
and bisit

http://youtube.com/watch?v=UU5wFUqoBbk

## 3.2. Docker setup

It's possible to run the app from the Docker container.

First of all, install the Docker: https://docs.docker.com/installation/ubuntulinux/

```bash
sudo apt-get install docker-io
```

Now pull the app image
```bash
docker pull ghnail/rhood_proxy
```

Prepare the directory, where the videos will be saved

```bash
mkdir -p /var/rhood_proxy/cache
```

And launch the container

```bash
docker run -it -p 8090:8090 -p 8081:8081 -v /var/rhood_proxy/cache:/data/rhood/cache ghnail/rhood_proxy
```
The proxy is 8081, the web interface address is 8090.

To test it with the actions from section 3.1.

# 4. How to build

You need a number of software tools:
- golang environment
- git
- youtube-dl

and if you want to set up separate box:
- lxc
- ansible

## 4.1. Golang

We need golang 1.3. Follow the official docs [https://golang.org/doc/install](https://golang.org/doc/install) to install it.

The shell command:
```bash
go version
```

must output something like:
> go version go1.3 linux/amd64

## 4.2. VCS

The go code requires git for the most of projects, but it's also good to see mercurial
in the system for the bitbucket/google.code libraries. Rsync is required for deployment tasks,
so it's optional dependency.
```bash
sudo apt-get install git mercurial rsync
```

Now the call:
```bash
git --version
```
will say something like that:
> git version 1.8.1.2

## 4.3. Youtube-dl

The best way is to use virtualenv+pip.

sudo apt-get install python-virtualenv
sudo apt-get install python-pip

Make shure you are able to write directory `/home/venv/rhood` and run command
```bash
virtualenv /home/venv/rhood
```

And now install youtube_dl from this environment
```bash
source /home/venv/rhood/bin/activate
pip install -I youtube_dl==2014.09.04.3
```

The request
```bash
/home/venv/rhood/bin/youtube-dl --version
```

must result something like that:

> 2014.09.04.3

## 4.4. Build the app
Download the project

```bash
go get github.com/ghnail/rhood
```
With dependencies: go to the project dir (for example, /home/user/gocode/src/github.com/ghnail/rhood )
and run `go get ./...`

```bash
cd /home/user/gocode/src/github.com/ghnail/rhood
go get ./...
```

And, finally, build and run the project:
```bash
cd cmd/rhood
go build

./rhood
```

You can also run auto tests:
```bash
cd /home/user/gocode/src/github.com/ghnail/rhood
cd rhood
go test -v
```

## 4.5. Test the app

Open admin page

http://localhost:2000/admin/status

Or, for example, try to cache this video:

http://localhost:2000/youtube/http://www.youtube.com/watch?v=UU5wFUqoBbk

If everything work fine, you can use proxy localhost:8081, and open direct youtube link

http://www.youtube.com/watch?v=UU5wFUqoBbk

## 4.6. Ansible delivery

You may want to deploy application on the separate box. You can build application,
copy binary, templates and web static files to the box, and configure Nginx
by yourself. But it can be done automatically with Ansible scripting.

The script will prepare environment, get github sources, build the application,
and configure daemon launch with the Nginx support.

In this section we will setup LXC container as a server box, and run ansible to do all
other stuff.

### 4.6.1. Install software

```bash
sudo apt-get install lxc lxc-templates
sudo apt-get install ansible
```
### 4.6.2. Prepare the box

```bash
sudo lxc-create -n rhoodbox -t ubuntu
sudo lxc-start -n rhoodbox
sudo lxc-list
```

The last command will show you the IP address of the rhoodbox. For example, 192.168.1.130.

> The ansible will download project and golang distribution from the web.
If you want to use local versions (good for slow network/dev purposes),
you can edit `playbook.yaml` lines `go_download_location: http://golang.org/dl/{{ go_tarball }}`
and `git: repo=git...` to target your local URLs. There might be comments with the working LAN examples.
And, if you are new to LXC, you may want to use local apt cache, for example, with
this instruction: https://bugs.launchpad.net/ubuntu/+source/lxc/+bug/1081786

Now go to the project root dir, and then go to dir
```bash
cd data/ansible
```
edit file `ansible_hosts` to target your box IP,



run script `download-roles.sh`

Now set up ssh key auth: http://askubuntu.com/questions/61557/how-do-i-set-up-ssh-authentication-keys

Then ssh to the rhood box, and install python:
> default user/password: ubuntu/ubuntu

```bash
ssh ubuntu@$RHOOD_BOX_IP
ubuntu@rhoodbox:~$ sudo apt-get update
ubuntu@rhoodbox:~$ sudo apt-get install python
```

Now, on the your dev box, in the `project_dir/data/ansible` dir, run script `run-ansible.sh`

This will take a while to download, build and configure everything.

Now visit `http://$RHOOD_BOX_IP:90` and see the admin page.

You can use proxy in two versions: nginx-optimized proxy with port 85, or use the 'raw' go application on the port 8081.

# 5. IDE setup

The Intellij IDEA has a good unofficial plugin for the Go language, and it has a nice quality.

# 6. Brief structure description

The main application parts are:

- youtube_dl, the python program to download videos from youtube
- Go application itself (proxy server + admin pages)
- html templates (they are in the separate dir, and not embedded in the app)
- satic files (js, css, videojs, jquery) and downloaded html pages/video files

For production there are few additional things:

- upstart script to launch/log app as daemon
- nginx as second proxy to serve static files


Directory structure:

- cmd/rhood/rhood - code to launch the app, 'main' function
- rhood - most of the golang code of the app
- data/ansible - ansible scripts how to build and deploy app on external server
- data/conf - all config files for the app
- data/conf/examples - examples of conf templates with substituted values
- data/templates - html templates
- www-rhood - all front-end, and saved html pages/video files

Ansible-target box, the build and deploy separate server:

- /home/ubuntu - home dir
- /home/ubuntu/gocode - go root
- /home/ubuntu/virtualenv - venv for the youtube_dl
- /home/ubuntu/rhood - dir of the deployed app
- /etc/nginx/nginx.conf - nginx
- /etc/init/drhood.conf - upstart daemon config
- /var/log/nginx - nginx logs
- /var/log/upstart/drhood.log - app daemon log
