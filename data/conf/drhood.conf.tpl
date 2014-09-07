description "rhood daemon"

start on (filesystem)
stop on runlevel [016]

respawn
console log
setuid www-data
setgid www-data

chdir {{ dir_rem_webserver }}

script
    {{ dir_rem_bin }}/rhood
end script