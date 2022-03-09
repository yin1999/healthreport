if ! command -v systemd >/dev/null 2>&1; then
	echo "systemd not exists, stopping..."
	exit 1
fi

serviceFile=/usr/lib/systemd/system/healthreport.service
if [ "$1" = "uninstall" ]; then
	if [ -f $serviceFile ]; then
		echo "uninstalling healthreport service..."
		systemctl disable healthreport.service
		systemctl stop healthreport.service
		rm -f ${serviceFile}
		systemctl daemon-reload
		echo "uninstall success"
	else
		echo "service not exists, stoping..."
	fi
	exit 0
fi

if [ "$1" != "install" ]; then
	echo "usage: $0 install|uninstall"
	exit 1
fi

parentDir=$(dirname $(dirname $(readlink -f "$0")))
program=${parentDir}/healthreport
if [ $# -lt 3 ]; then
	if [ ! -f "${parentDir}/account.json" ]; then
		echo "Usage: $0 install <username> <password> [punchTime]"
		echo "Example:"
		echo "case#1    $0 install 1862410000 123456"
		echo "case#2    $0 install 1862410000 123456 08:00"
		echo
		echo "或者你可以先创建文件 '${parentDir}/account.json'"
		echo "Example: create account.json by Command Line:"
		echo "    cd ${parentDir} && ./healthreport -u 1862410000 -p 123456 -save"
		echo "然后运行命令:"
		echo "    $0 install [punchTime]"
		echo "Example:"
		echo "case#1    $0 install"
		echo "case#2    $0 install 08:00"
		exit 1
	else
		execStart="${program} -account \${CREDENTIALS_DIRECTORY}/account.json"
		if [ $# -eq 2 ]; then
			execStart="${execStart} -t $2"
		fi
	fi
else
	execStart="${program} -u $2 -p $3"
	if [ $# -eq 4 ]; then
		execStart="${execStart} -t $4"
	fi
fi

cat > ${serviceFile} <<- EOF
[Unit]
Description=healthreport daemon
After=network-online.target

[Service]
# set http proxy
# Environment="HTTP_PROXY=http://localhost:1080"
# Environment="HTTPS_PROXY=http://localhost:1080"
DynamicUser=yes
LoadCredential=account.json:${parentDir}/account.json
ExecStart=${execStart}
ExecReload=/bin/kill -HUP \$MAINPID
Type=notify

[Install]
WantedBy=multi-user.target
EOF

# reload systemd
systemctl daemon-reload

echo "install healthreport.service success"
