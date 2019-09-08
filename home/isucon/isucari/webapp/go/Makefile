export GO111MODULE=on

all: deploy rotate.access_log rotate.slow_log

isucari: *.go
	go build -o isucari

.PHONY: deploy
deploy:
	ssh isucon9q-1 "systemctl stop isucari.golang.service"
	ssh isucon9q-2 "systemctl stop isucari.golang.service"
	ssh isucon9q-3 "systemctl stop isucari.golang.service"
	ssh isucon9q-1 "cd /home/isucon/isucari/webapp/go && make isucari"
	scp isucon9q-1:/home/isucon/isucari/webapp/go/isucari isucon9q-2:/home/isucon/isucari/webapp/go/isucari
	scp isucon9q-1:/home/isucon/isucari/webapp/go/isucari isucon9q-3:/home/isucon/isucari/webapp/go/isucari
	ssh isucon9q-1 "systemctl start isucari.golang.service"
	ssh isucon9q-2 "systemctl start isucari.golang.service"
	ssh isucon9q-3 "systemctl start isucari.golang.service"

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
