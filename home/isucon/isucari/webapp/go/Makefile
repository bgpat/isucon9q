export GO111MODULE=on

all: isucari rotate.access_log rotate.slow_log

isucari: *.go
	go build -o isucari

rotate.access_log:
	cp /var/log/nginx/access.log /var/log/nginx/access.log.backup
	cat /var/log/nginx/access.log.backup | kataribe > kataribe.out
	slackcat -c kataribe -i 😏 -u 'kataribeくん' kataribe.out
	echo  > /var/log/nginx/access.log

rotate.slow_log:
	mysqldumpslow -s t /var/log/mysql/mysql-slow.sql > /var/log/mysql/mysql-slow.sql.backup
	head -35 /var/log/mysql/mysql-slow.sql.backup > slow_log.out
	slackcat -c slowlog -i 😓  -u 'slowlogくん' slow_log.out
	echo  > /var/log/mysql/mysql-slow.sql
