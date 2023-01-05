module github.com/lixin9311/sms-bot

go 1.18

require (
	entgo.io/ent v0.9.1
	github.com/godbus/dbus/v5 v5.0.5
	github.com/lixin9311/backoff/v2 v2.0.0
	github.com/lixin9311/jobcan v0.0.0-20220114073326-da9c37f34160
	github.com/maltegrosse/go-modemmanager v0.1.0
	github.com/mattn/go-sqlite3 v1.14.8
	gopkg.in/tucnak/telebot.v2 v2.3.5
)

require (
	github.com/PuerkitoBio/goquery v1.8.0 // indirect
	github.com/andybalholm/cascadia v1.3.1 // indirect
	github.com/go-resty/resty/v2 v2.7.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/juju/go4 v0.0.0-20160222163258-40d72ab9641a // indirect
	github.com/juju/persistent-cookiejar v0.0.0-20171026135701-d5e5a8405ef9 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.0.0-20220114011407-0dd24b26b47d // indirect
	gopkg.in/errgo.v1 v1.0.1 // indirect
	gopkg.in/retry.v1 v1.0.3 // indirect
)

replace github.com/maltegrosse/go-modemmanager => github.com/lixin9311/go-modemmanager v0.0.0-20211014064810-def1e6f906dd
