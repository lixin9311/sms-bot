package bot

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"regexp"
	"strconv"
	"strings"
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
	modemAddedTemplate   = template.Must(template.New("info").Parse(modemAddedTemplateStr))
	modemRemovedTemplate = template.Must(template.New("info").Parse(modemRemovedTemplateStr))
	singleSMSTemplate    = template.Must(template.New("singleSMS").Parse(singleSMSTemplateStr))
	multipleSMSTemplate  = template.Must(template.New("multiSMS").Funcs(funcMap).Parse(multipleSMSTemplateStr))
	errorTemplate        = template.Must(template.New("error").Parse(errorTemplateStr))
	extractSender        = regexp.MustCompile(`^From (\d+)`)
	// Universal markup builders.
	menu = &tb.ReplyMarkup{ResizeReplyKeyboard: true}

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
}

func userIDFilter(userID int) func(*tb.Update) bool {
	if userID == 0 {
		return func(up *tb.Update) bool {
			return true
		}
	}
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
	b.bot.Handle(&btnSMS, b.onListSMSOption)
	b.bot.Handle(&btnInfo, b.onInfo)
	b.bot.Handle(&btnHelp, b.help)
	b.bot.Handle(tb.OnCallback, func(cb *tb.Callback) {
		parts := strings.Split(strings.TrimSpace(cb.Data), "|")
		if len(parts) != 2 {
			return
		}
		action := parts[0]
		data := parts[1]
		switch action {
		case "modem_info":
			b.onModemInfo(cb, data)
		case "sms_page":
			b.onSMSListPage(cb, data)
		default:
			b.bot.Respond(cb, &tb.CallbackResponse{
				CallbackID: cb.ID,
			})
		}
	})
}

func (b *Bot) onModemInfo(cb *tb.Callback, data string) {
	b.bot.Respond(cb, &tb.CallbackResponse{
		CallbackID: cb.ID,
	})
	info := b.man.GetModemInfo(data)
	if info == nil {
		if _, err := b.bot.Send(cb.Sender, "Modem info unavailable", tb.ModeHTML); err != nil {
			log.Println(err)
		}
	} else {
		str, err := formatModemInfo(info)
		if err != nil {
			b.onError(err)
			return
		}

		if _, err := b.bot.Send(cb.Sender, str, tb.ModeHTML); err != nil {
			log.Println(err)
		}
	}
}

func (b *Bot) onSMSListPage(cb *tb.Callback, data string) {
	var msg string
	defer b.bot.Respond(cb, &tb.CallbackResponse{
		CallbackID: cb.ID,
		Text:       msg,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	parts := strings.Split(data, ":")
	if len(parts) != 2 {
		return
	}
	id := parts[0]
	offset, err := strconv.Atoi(parts[1])
	if err != nil {
		return
	}
	if offset < 0 {
		return
	}

	smss, total, err := b.man.ListSMS(ctx, id, offset, 10)
	if err != nil {
		b.onError(err)
		return
	}
	text, err := formatMultipleSMS(smss)
	if err != nil {
		b.onError(err)
		return
	}
	var prevOffset, nextOffset int
	prevOffset = offset - 10
	if len(smss) < 10 {
		nextOffset = -1
	} else {
		nextOffset = offset + 10
	}
	currentPage := offset/10 + 1
	totalPages := total / 10
	if total%10 != 0 {
		totalPages++
	}
	selector := &tb.ReplyMarkup{}
	btnPrev := selector.Data("⬅", "sms_page", id+":"+strconv.Itoa(prevOffset))
	btnPage := selector.Data(fmt.Sprintf("%d/%d", currentPage, totalPages), "noop")
	btnNext := selector.Data("➡", "sms_page", id+":"+strconv.Itoa(nextOffset))
	selector.Inline(
		selector.Row(btnPrev, btnPage, btnNext),
	)
	if cb.Message == nil {
		if _, err := b.bot.Send(cb.Sender, text, selector, tb.ModeHTML); err != nil {
			log.Println(err)
		}
	} else {
		if _, err := b.bot.Edit(cb.Message, text, selector, tb.ModeHTML); err != nil {
			log.Println(err)
		}
	}
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

func (b *Bot) onCall(call *manager.PhoneCall) {
	text, err := formatIncomingCall(call)
	if err != nil {
		b.onError(err)
		return
	}
	b.bot.Send(userID(b.myUserID), text, tb.ModeHTML)
}

func (b *Bot) onModemAdded(id string) {
	text, err := formatModemAdded(id)
	if err != nil {
		b.onError(err)
		return
	}
	b.bot.Send(userID(b.myUserID), text, tb.ModeHTML)
}

func (b *Bot) onModemRemoved(id string) {
	text, err := formatModemRemoved(id)
	if err != nil {
		b.onError(err)
		return
	}
	b.bot.Send(userID(b.myUserID), text, tb.ModeHTML)
}

// Start will start the bot
func (b *Bot) Start(ctx context.Context) error {
	ch := b.man.Subscribe()

	go func() {
		for {
			select {
			case event := <-ch:
				if b.myUserID == 0 {
					continue
				}
				switch event.Type {
				case manager.ErrorEvent:
					b.onError(event.Payload.(error))
				case manager.SMSEvent:
					b.onSMS(event.Payload.(*ent.SMS))
				case manager.PhoneCallEvent:
					b.onCall(event.Payload.(*manager.PhoneCall))
				case manager.ModemAddEvent:
					b.onModemAdded(event.Payload.(string))
				case manager.ModemRemoveEvent:
					b.onModemRemoved(event.Payload.(string))
				}
			case <-ctx.Done():
				b.bot.Stop()
			}
		}
	}()

	b.bot.Start()
	return nil
}

func formatModemAdded(id string) (string, error) {
	buf := new(bytes.Buffer)
	err := modemAddedTemplate.Execute(buf, id)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatModemRemoved(id string) (string, error) {
	buf := new(bytes.Buffer)
	err := modemRemovedTemplate.Execute(buf, id)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func formatIncomingCall(call *manager.PhoneCall) (string, error) {
	buf := new(bytes.Buffer)
	err := incomingCallTemplate.Execute(buf, call)
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

var incomingCallTemplateStr = `<b>Incoming call to {{.ModemID}} from:</b> {{.Caller}}`
var modemAddedTemplateStr = `<b>Modem connected:</b> {{.}}`
var modemRemovedTemplateStr = `<b>Modem disconnected:</b> {{.}}`
var errorTemplateStr = `❌ <code>{{.Error}}</code>`

var multipleSMSTemplateStr = `
{{range $i, $sms := .}}
<b>[{{add1 $i}}/{{len $}}] From {{$sms.Number}} To {{$sms.ModemID}} {{$sms.DischargeTimestamp.Format "2006/01/02 15:04:05"}}:</b>
{{if $sms.Text}} {{$sms.Text}} {{else}} <pre>{{$sms.String}}</pre> {{end}}
{{end}}
`

var singleSMSTemplateStr = `<b>From {{.Number}} To {{.ModemID}} {{.DischargeTimestamp.Format "2006/01/02 15:04:05"}}:</b>
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
