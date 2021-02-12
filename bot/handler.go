package bot

import (
	"context"
	"fmt"
	"log"

	tb "gopkg.in/tucnak/telebot.v2"
)

func (b *Bot) onStart(m *tb.Message) {
	if !m.Private() {
		return
	}
	b.bot.Send(m.Sender, "Hello!", menu)
}

func (b *Bot) onInfo(m *tb.Message) {
	info := b.man.Info()
	str, err := formatModemInfo(info)
	if err != nil {
		b.onError(err)
		return
	}
	if _, err := b.bot.Send(m.Sender, str, tb.ModeHTML); err != nil {
		log.Println(err)
	}
}

func (b *Bot) onListSMS(m *tb.Message) {
	ctx := context.Background()
	smss, err := b.man.List(ctx, 10)
	if err != nil {
		b.onError(err)
		return
	}
	text, err := formatMultipleSMS(smss)
	if err != nil {
		b.onError(err)
		return
	}
	b.bot.Send(m.Sender, text, tb.ModeHTML)
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
	case "inline":
		b.bot.Send(m.Sender, "*some* inline", selector, tb.ModeMarkdownV2)
	}
	return
}

func (b *Bot) help(m *tb.Message) {
	b.bot.Send(m.Sender, "help")
}

func (b *Bot) prev(c *tb.Callback) {
	b.bot.Respond(c, &tb.CallbackResponse{})
}
