#!/bin/bash

sftp -o "ProxyCommand nc -v -X connect -x [proxy:host] %h %p" -oIdentityFile=[ssh key] user@blahblah.com <<EOF 
cd [some path]
lcd [some local path]
put [filename]
EOF
