package bot

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/lixin9311/sms-bot/ent"
	"github.com/lixin9311/sms-bot/manager"
	tb "gopkg.in/tucnak/telebot.v2"
)

// Bot ...
type Bot struct {
	man      *manager.Manager
	myUserID int
	bot      *tb.Bot
}

type userID int

func (u userID) Recipient() string {
	return strconv.Itoa(int(u))
}

var (
	// myID          = 117609741
	funcMap = template.FuncMap{
		"add1": func(i int) int {
			return i + 1
		},
	}
	incomingCallTemplate = template.Must(template.New("incomingCall").Parse(incomingCallTemplateStr))
	modemInfoTemplate    = template.Must(template.New("info").Parse(modemInfoTemplateStr))
	singleSMSTemplate    = template.Must(template.New("singleSMS").Parse(singleSMSTemplateStr))
	multipleSMSTemplate  = template.Must(template.New("multiSMS").Funcs(funcMap).Parse(multipleSMSTemplateStr))
	errorTemplate        = template.Must(template.New("error").Parse(errorTemplateStr))
	extractSender        = regexp.MustCompile(`^From (\d+)`)
	// Universal markup builders.
	menu     = &tb.ReplyMarkup{ResizeReplyKeyboard: true}
	selector = &tb.ReplyMarkup{}

	// Reply buttons.
	btnHelp     = menu.Text("ℹ Help")
	btnInfo     = menu.Text("ℹ Info")
	btnSettings = menu.Text("⚙ Settings")
	btnSMS      = menu.Text("✉ List SMS")

	// Inline buttons.
	//
	// Pressing it will cause the client to
	// send the bot a callback.
	//
	// Make sure Unique stays unique as per button kind,
	// as it has to be for callback routing to work.
	//
	btnPrev  = selector.Data("⬅", "prev")
	btnNext  = selector.Data("➡", "next")
	btnReply = selector.Data("↩️", "reply")
	commands = []tb.Command{
		{
			Text:        "list",
			Description: "list 10 recent SMSs",
		},
		{
			Text:        "me",
			Description: "print user information",
		},
	}
)

func init() {
	menu.Reply(
		menu.Row(btnSMS, btnInfo),
		menu.Row(btnHelp, btnSettings),
	)

	selector.Inline(
		selector.Row(btnPrev, btnNext),
	)
}

func userIDFilter(userID int) func(*tb.Update) bool {
	return func(up *tb.Update) bool {
		if up.Message != nil && up.Message.Sender != nil && up.Message.Sender.ID == userID {
			return true
		} else if up.EditedMessage != nil && up.EditedMessage.Sender != nil && up.EditedMessage.Sender.ID == userID {
			return true
		} else if up.Callback != nil && up.Callback.Sender != nil && up.Callback.Sender.ID == userID {
			return true
		} else if up.Query != nil && up.Query.From.ID == userID {
			return true
		} else if up.ChosenInlineResult != nil && up.ChosenInlineResult.From.ID == userID {
			return true
		} else if up.PollAnswer != nil && up.PollAnswer.User.ID == userID {
			return true
		}
		return false
	}
}

func usernameFilter(username string) func(*tb.Update) bool {
	return func(up *tb.Update) bool {
		if up.Message != nil && up.Message.Sender != nil && up.Message.Sender.Username == username {
			return true
		} else if up.EditedMessage != nil && up.EditedMessage.Sender != nil && up.EditedMessage.Sender.Username == username {
			return true
		} else if up.Callback != nil && up.Callback.Sender != nil && up.Callback.Sender.Username == username {
			return true
		} else if up.Query != nil && up.Query.From.Username == username {
			return true
		} else if up.ChosenInlineResult != nil && up.ChosenInlineResult.From.Username == username {
			return true
		} else if up.PollAnswer != nil && up.PollAnswer.User.Username == username {
			return true
		}
		return false
	}
}

// NewBot will init a tg bot
func NewBot(token string, myUserID int, man *manager.Manager) (*Bot, error) {
	b, err := tb.NewBot(tb.Settings{
		Token: token,
		Poller: &tb.MiddlewarePoller{
			Poller: &tb.LongPoller{Timeout: 10 * time.Second},
			Filter: userIDFilter(myUserID),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create tg bot: %w", err)
	}
	if err := b.SetCommands(commands); err != nil {
		return nil, fmt.Errorf("failed to reset commands: %w", err)
	}
	bot := &Bot{bot: b, myUserID: myUserID, man: man}
	bot.init()
	return bot, nil
}

func (b *Bot) init() {
	b.bot.Handle("/start", b.onStart)
	b.bot.Handle("/me", b.onMe)
	b.bot.Handle(tb.OnText, b.onText)
	b.bot.Handle(&btnSMS, b.onListSMS)
	b.bot.Handle(&btnInfo, b.onInfo)
	b.bot.Handle(&btnHelp, b.help)
	b.bot.Handle(&btnPrev, b.prev)
}

func (b *Bot) onError(in error) {
	text, err := formatError(in)
	if err != nil {
		log.Println(err)
		return
	}
	b.bot.Send(userID(b.myUserID), text, tb.ModeHTML)
}

func (b *Bot) onSMS(sms *ent.SMS) {
	text, err := formatSingleSMS(sms)
	if err != nil {
		b.onError(err)
		return
	}
	b.bot.Send(userID(b.myUserID), text, tb.ModeHTML)
}

func (b *Bot) onCall(number string) {
	text, err := formatIncomingCall(number)
	if err != nil {
		b.onError(err)
		return
	}
	b.bot.Send(userID(b.myUserID), text, tb.ModeHTML)
}

// Start will start the bot
func (b *Bot) Start(ctx context.Context) error {
	smsch, errch, err := b.man.SubscribeSMS(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe from sms: %w", err)
	}

	phonech, perrch, err := b.man.SubscribeCall(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe from phone call: %w", err)
	}

	go func() {
		for {
			select {
			case sms := <-smsch:
				b.onSMS(sms)
			case err := <-errch:
				b.onError(err)
			case err := <-perrch:
				b.onError(err)
			case phone := <-phonech:
				b.onCall(phone)
			case <-ctx.Done():
				b.bot.Stop()
			}
		}
	}()

	b.bot.Start()
	return nil
}

func formatIncomingCall(number string) (string, error) {
	buf := new(bytes.Buffer)
	err := incomingCallTemplate.Execute(buf, number)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatSingleSMS(sms *ent.SMS) (string, error) {
	buf := new(bytes.Buffer)
	err := singleSMSTemplate.Execute(buf, sms)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatMultipleSMS(smss []*ent.SMS) (string, error) {
	buf := new(bytes.Buffer)
	err := multipleSMSTemplate.Execute(buf, smss)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatError(in error) (string, error) {
	buf := new(bytes.Buffer)
	err := errorTemplate.Execute(buf, in)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatUser(user *tb.User) string {
	return fmt.Sprintf("Username: %s\nID: %d\nFirst: %s\nLast: %s\n", user.Username, user.ID, user.FirstName, user.LastName)
}

func formatModemInfo(in *manager.ModemInfo) (string, error) {
	buf := new(bytes.Buffer)
	if err := modemInfoTemplate.Execute(buf, in); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var incomingCallTemplateStr = `<b>Incoming call from:</b> {{.}}`

var errorTemplateStr = `❌ <code>{{.Error}}</code>`

var multipleSMSTemplateStr = `
{{range $i, $sms := .}}
<b>[{{add1 $i}}/{{len $}}] From {{$sms.Number}} {{$sms.DischargeTimestamp.Format "2006/01/02 15:04:05"}}:</b>
{{if $sms.Text}} {{$sms.Text}} {{else}} <pre>{{$sms.String}}</pre> {{end}}
{{end}}
`

var singleSMSTemplateStr = `<b>From {{.Number}} {{.DischargeTimestamp.Format "2006/01/02 15:04:05"}}:</b>
{{if .Text}} {{.Text}} {{else}} <pre>{{.String}}</pre> {{end}}
`

var modemInfoTemplateStr = `<b>Manufacturer:</b> 🏭 <code>{{.Manufacturer}}</code>
<b>Model:</b> 🚗 <code>{{.Model}}</code>
<b>Signal:</b> 📶 {{.SignalQuality}}%
<b>State:</b> {{.State}}
<b>Phone Number:</b> 📱 {{.PhoneNumber}}
<b>Access:</b> {{.AccessTech}}
<b>SIM Operator Code:</b> {{.SimOperatorCode}}
<b>IMEI:</b> {{.IMEI}}
<b>Network Operator:</b> {{.NetworkOperatorName}} {{.NetworkOperatorCode}}
`
