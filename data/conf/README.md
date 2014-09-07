All files have 2 versions: *.conf, with example data, and *.conf.tpl, to be
processed by ansible-playbook template engine.

files:
- drhood.conf — upstart config to demonize rhood
- nginx.conf — nginx config to serve static files, and to proxy requests to
web engine control and entire Internet
- rhood.conf — parameters of web server, which modifies youtube pages
and listens to 'download video' command
-rhood.test.conf — conf is used in unittesting
-playbook.yaml — steps to deploy the application

