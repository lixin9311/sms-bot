package bot

import (
	"context"
	"fmt"
	"log"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (b *Bot) onStart(m *tb.Message) {
	if !m.Private() {
		return
	}
	b.bot.Send(m.Sender, "Hello!", menu)
}

func (b *Bot) onInfo(m *tb.Message) {
	modems := b.man.ListModems()

	if len(modems) == 0 {
		if _, err := b.bot.Send(m.Sender, "No modem is available..."); err != nil {
			log.Println(err)
		}
		return
	}

	selector := &tb.ReplyMarkup{}
	btns := make([]tb.Btn, len(modems))
	for i, id := range modems {
		btns[i] = selector.Data(id, "modem_info", id)
	}
	rows := make([]tb.Row, 0, len(btns)/3+1)
	for len(btns) > 0 {
		var row tb.Row
		if len(btns) >= 3 {
			row = selector.Row(btns[:3]...)
			btns = btns[3:]
		} else {
			row = selector.Row(btns...)
			btns = btns[len(btns):]
		}
		rows = append(rows, row)
	}
	selector.Inline(rows...)
	if _, err := b.bot.Send(m.Sender, "Choose a modem:", selector); err != nil {
		log.Println(err)
	}
	// info := b.man.Info()
	// str, err := formatModemInfo(info)
	// if err != nil {
	// 	b.onError(err)
	// 	return
	// }
	// if _, err := b.bot.Send(m.Sender, str, tb.ModeHTML); err != nil {
	// 	log.Println(err)
	// }
}

func (b *Bot) onListSMSOption(m *tb.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	modems, err := b.man.ListSMSModemIDs(ctx)
	if err != nil {
		b.onError(err)
	}
	if len(modems) == 0 {
		if _, err := b.bot.Send(m.Sender, "No modem is available..."); err != nil {
			log.Println(err)
		}
		return
	}
	selector := &tb.ReplyMarkup{}
	btns := make([]tb.Btn, len(modems)+1)
	btns[0] = selector.Data("All", "sms_page", "*:0")
	for i, id := range modems {
		btns[i+1] = selector.Data(id, "sms_page", id+":0")
	}
	rows := make([]tb.Row, 0, len(btns)/3+1)
	for len(btns) > 0 {
		var row tb.Row
		if len(btns) >= 3 {
			row = selector.Row(btns[:3]...)
			btns = btns[3:]
		} else {
			row = selector.Row(btns...)
			btns = btns[len(btns):]
		}
		rows = append(rows, row)
	}
	selector.Inline(rows...)
	if _, err := b.bot.Send(m.Sender, "Choose a modem:", selector); err != nil {
		log.Println(err)
	}
}

func (b *Bot) onMe(m *tb.Message) {
	b.bot.Send(m.Sender, formatUser(m.Sender))
}

func (b *Bot) onText(m *tb.Message) {
	if m.ReplyTo != nil && m.ReplyTo.Sender.ID == b.bot.Me.ID {
		matches := extractSender.FindStringSubmatch(m.ReplyTo.Text)
		if len(matches) == 2 {
			to := matches[1]
			b.bot.Send(m.Sender, fmt.Sprintf("[🚧WIP] sending reply message to %s:\n%s", to, m.Text))
			return
		}
	}
	switch m.Text {
	case "menu":
		b.bot.Send(m.Sender, "some menu", menu)
	}
	return
}

func (b *Bot) help(m *tb.Message) {
	b.bot.Send(m.Sender, "help")
}

func (b *Bot) prev(c *tb.Callback) {
	b.bot.Respond(c, &tb.CallbackResponse{})
}
