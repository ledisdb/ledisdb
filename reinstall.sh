#/bin/bash
len=$(git status |grep modified |wc | awk '{print $1}')
if [ "$len" -gt 0 ]; then
    printf "\nYou have local modified files\n"
    exit 1
fi

git pull --rebase
ps -ef |grep -v grep |grep ledis| awk '{print $2}'|xargs kill -9       

go install ./...

source ./dev.sh
nohup ledis-server & 
day=$(ps aux|grep -v grep |grep ledis-server | awk '{print $9}')
printf "ledis-server 启动于 $day"
