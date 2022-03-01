# path: /usr/lib/systemd/system/healthreport.service
# Please replace `curly braces` with the actual values
[Unit]
Description=healthreport daemon
After=network-online.target

[Service]
# set http proxy
# Environment="HTTP_PROXY=http://localhost:1080"
# Environment="HTTPS_PROXY=http://localhost:1080"
ExecStart={dir to exec}/healthreport -u {username} -p {password} -t {punch time}
ExecReload=/bin/kill -HUP $MAINPID
Type=notify

[Install]
WantedBy=multi-user.target
