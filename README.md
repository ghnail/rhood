# 1. Description:

This is proxy server for Youtube. It extends all video pages with a new button to cache the data:

[Cache button](https://raw.githubusercontent.com/ghnail/rhood/master/doc/img/cache-button.png)

When you press it, the new dialog will appear to select quality of the video to download.

[Download options](https://raw.githubusercontent.com/ghnail/rhood/master/doc/img/cache-button.png)

And the next time you visit this page, you will get video traffic from the local server:

[Cached video](https://raw.githubusercontent.com/ghnail/rhood/master/doc/img/cached-video.png)

It will appear as a green text "Video is cached", and you also will have the different video player.

# 2. How it works

There is a program, youtube-dl, which can download Youtube video to local file.
Now it can be played without internet connection, which is great.

But it's also good to see video info, comments and all other youtube stuff, so here is an idea:
we download video, and the next time user touch same video page, we replace link to remote video
with the link to our LAN server.

Youtube has complicated player, which works with a lot of video chunks instead of single video file,
and that is why it's not that easy to cache the media. Because of that we need to replace the entire player
to use a video from LAN. We use videoJs for this purpose.

The big minus of this replacement is that we lose all features of youtube player, like annotations or subtitles.

As a final word, the application has two main parts: the proxy server (which modifies youtube pages), and the admin
interface, which downloads/hosts video files (and also has few pages for settings).

# 3. How to build

You need a number of software tools:
- golang environment
- git
- youtube-dl

and if you want to set up separate box:
- lxc
- ansible

## 3.1. Golang

We need golang 1.3. Follow the official docs [https://golang.org/doc/install](https://golang.org/doc/install) to install it.

The shell command:
```bash
go version
```

must result something like:
> go version go1.3 linux/amd64

## 3.2. CVS

The go code requires git for the most of projects, but it's also good to see mercurial in
the system for the bitbucket/google.code libraries. Rsync is required for deployment tasks,
so it's optional dependency.
```bash
sudo apt-get install git mercurial rsync
```

Now the call:
```bash
git --version
```
will return something like that:
> git version 1.8.1.2

## 3.3. Youtube-dl

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

The command
```bash
/home/venv/rhood/bin/youtube-dl --version
```

must result something like that:

> 2014.09.04.3

## 3.4. Build the app
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

## 3.5. Test the app

Open admin page

http://localhost:2000/admin/status

Or, for example, try to cache this video:

http://localhost:2000/youtube/http://www.youtube.com/watch?v=UU5wFUqoBbk

If everything work fine, you can use proxy localhost:8081, and open direct youtube link

http://www.youtube.com/watch?v=UU5wFUqoBbk

You can also edit file
rhood/conf.go, and replace line

`"controlBoxPublicAddress": "localhost:2000",`

with your IP address:

`"controlBoxPublicAddress": "192.168.1.189:2000",`

And try it from your LAN.

## 3.6. Ansible delivery

You may want to deploy application on the separate box. You can build application,
copy binary, templates, configs and web static files to the box, and configure Nginx
by yourself. But it can be done automatically with Ansible scripting.

The script will prepare environment, get github sources, build the application,
and configure daemon launch with the Nginx support.

In this section we will setup LXC container as a server box, and run ansible to do all
other stuff.

### 3.6.1. Install software

```bash
sudo apt-get install lxc lxc-templates
sudo apt-get install ansible
```
### 3.6.2. Prepare the box

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

# 4. IDE setup

The Intellij IDEA has a good unofficial plugin for the Go language, and it has a nice quality.

# 5. Brief description

The main code parts are:

- youtube_dl, the python program to download videos from youtube
- Go application itself (proxy server + admin pages)
- app config (ports and fs paths)
- html templates (they are in the separate dir, and not embedded in the app)
- satic files (js, css, videojs, jquery) and downloaded html pages/video files

For produuction there are few additional things:

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
