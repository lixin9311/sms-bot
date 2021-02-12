package manager

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/lixin9311/backoff"
	"github.com/lixin9311/sms-bot/ent"
	"github.com/lixin9311/sms-bot/ent/sms"
	"github.com/maltegrosse/go-modemmanager"
)

// Manager ...
type Manager struct {
	modem  modemmanager.Modem
	msg    modemmanager.ModemMessaging
	client *ent.Client

	smsCH         chan *ent.SMS
	smsErrCH      chan error
	closeCHSMS    chan struct{}
	smsSubscribed bool

	phoneCH         chan string
	phoneErrCH      chan error
	closeCHPhone    chan struct{}
	phoneSubscribed bool
}

// NewManager ...
func NewManager(modem modemmanager.Modem, msg modemmanager.ModemMessaging, client *ent.Client) *Manager {
	return &Manager{modem: modem, msg: msg, client: client}
}

// List ...
func (m *Manager) List(ctx context.Context, limit int) ([]*ent.SMS, error) {
	if limit == 0 {
		limit = 100
	}
	results, err := m.client.SMS.Query().Order(ent.Desc(sms.FieldDischargeTimestamp)).Limit(limit).All(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sms from db: %w", err)
	}
	return results, nil
}

// LoadSMS ...
func (m *Manager) LoadSMS(ctx context.Context) ([]*ent.SMS, error) {
	smss, err := m.msg.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list smss: %w", err)
	}
	log.Printf("got %d sms(s)\n", len(smss))
	results := make([]*ent.SMS, 0, len(smss))
	for _, v := range smss {
		bytes, _ := v.MarshalJSON()
		log.Printf("%s", bytes)
		state, err := v.GetState()
		if err != nil {
			return nil, fmt.Errorf("failed to get sms state: %w", err)
		}
		if state == modemmanager.MmSmsStateReceiving {
			go m.handleAsyncSMS(v)
			continue
		}
		if state == modemmanager.MmSmsStateReceived {
			entSMS, err := m.saveAndDeleteSMS(ctx, v)
			if err != nil {
				return nil, err
			}
			results = append(results, entSMS)
		}
	}
	return results, nil
}

// SubscribeSMS ...
func (m *Manager) SubscribeSMS(ctx context.Context) (<-chan *ent.SMS, <-chan error, error) {
	if m.smsSubscribed {
		return nil, nil, errors.New("already listening, unsubscribe first")
	}
	m.smsSubscribed = true

	m.smsCH = make(chan *ent.SMS)
	m.smsErrCH = make(chan error)
	m.closeCHSMS = make(chan struct{})

	go func() {
		ch := m.msg.SubscribeAdded()
		defer func() {
			m.msg.Unsubscribe()
			m.smsSubscribed = false
		}()
		for {
			select {
			case <-m.closeCHSMS:
				return
			case <-ctx.Done():
				return
			case sig := <-ch:
				fmt.Printf("dbus signal: %+v\n", *sig)
				sms, received, err := m.msg.ParseAdded(sig)
				if err != nil {
					log.Println(fmt.Errorf("failed to parse sms added: %w", err))
					continue
				}
				if !received {
					continue
				}
				state, err := sms.GetState()
				if err != nil {
					err = fmt.Errorf("failed to get sms state: %w", err)
					log.Print(err)
					m.smsErrCH <- err
					continue
				}
				if state == modemmanager.MmSmsStateReceived {
					log.Println("sms recved")
					entSMS, err := m.saveAndDeleteSMS(ctx, sms)
					if err != nil {
						m.smsErrCH <- err
						continue
					}
					m.smsCH <- entSMS
				} else if state == modemmanager.MmSmsStateReceiving {
					log.Println("sms still receiving")
					m.handleAsyncSMS(sms)
				}
			}
		}
	}()
	return m.smsCH, m.smsErrCH, nil
}

// UnsubscribeSMS ...
func (m *Manager) UnsubscribeSMS() {
	if m.smsSubscribed && m.closeCHSMS != nil {
		close(m.closeCHSMS)
	}
}

// UnsubscribeCall ...
func (m *Manager) UnsubscribeCall() {
	if m.phoneSubscribed && m.closeCHPhone != nil {
		close(m.closeCHPhone)
	}
}

// SubscribeCall will subscribe to incoming phone calls.
// It will only return caller's number, can do nothing else
func (m *Manager) SubscribeCall(ctx context.Context) (<-chan string, <-chan error, error) {
	if m.phoneSubscribed {
		return nil, nil, errors.New("already listening, unsubscribe first")
	}
	m.phoneSubscribed = true

	m.phoneCH = make(chan string)
	m.phoneErrCH = make(chan error)
	m.closeCHPhone = make(chan struct{})
	voice, err := m.modem.GetVoice()
	if err != nil {
		m.phoneSubscribed = false
		return nil, nil, fmt.Errorf("failed to get voice: %w", err)
	}
	go func() {
		defer func() {
			voice.Unsubscribe()
			m.phoneSubscribed = false
		}()

		ch := voice.SubscribeCallAdded()

		for {
			select {
			case <-m.closeCHPhone:
				return
			case <-ctx.Done():
				return
			case sig := <-ch:
				call, err := voice.ParseCallAdded(sig)
				if err != nil {
					log.Println(fmt.Errorf("failed to parse call added: %w", err))
					continue
				}
				direction, err := call.GetDirection()
				if err != nil {
					m.phoneErrCH <- fmt.Errorf("failed to get call direction: %w", err)
					continue
				}
				if direction == modemmanager.MmCallDirectionIncoming {
					log.Println("phone call recved")
					number, err := call.GetNumber()
					if err != nil {
						m.phoneErrCH <- fmt.Errorf("failed to get call number: %w", err)
						continue
					}
					if number == "" {
						number = "Unknown caller"
					}
					m.phoneCH <- number
				}
			}
		}
	}()
	return m.phoneCH, m.phoneErrCH, nil
}

func (m *Manager) handleAsyncSMS(sms modemmanager.Sms) {
	maxRetries := 10
	succ := false
	exp := backoff.NewExponential(&backoff.Config{
		BaseDelay: time.Second,
		MaxDelay:  10 * time.Second,
	})
	for retries := 0; retries < maxRetries; retries++ {
		state, err := sms.GetState()
		if err != nil {
			log.Printf("failed to fetch state for sms: %v", err)
			continue
		}
		if state == modemmanager.MmSmsStateReceived {
			succ = true
			break
		}
		dur := exp.Backoff(retries)
		log.Printf("no luck, sleep for %v", dur)
		time.Sleep(dur)
	}
	if succ {
		entSMS, err := m.saveAndDeleteSMS(context.Background(), sms)
		if err != nil {
			log.Printf("failed to save sms: %v", err)
			return
		}
		m.smsCH <- entSMS
	} else {
		log.Printf("unable to download sms after %d retries, drop it", maxRetries)
		number, _ := sms.GetNumber()
		if number == "" {
			number = "unknown"
		}
		ts := mustGetSMSTimestamp(sms)
		err := fmt.Errorf("Got a sms from %s at %s, unable to download sms after %d retries, drop it",
			number,
			ts.Local().Format("2006/01/02 15:04:05"),
			maxRetries,
		)
		log.Println("notifying via error ch")
		m.smsErrCH <- err
		log.Println("deleting the sms")
		if err := m.msg.Delete(sms); err != nil {
			log.Printf("failed to delete sms: %v", err)
		}
	}
}

// ModemInfo ...
type ModemInfo struct {
	Manufacturer        string
	Model               string
	State               string
	SignalQuality       int
	PhoneNumber         string
	AccessTech          string
	SimOperatorCode     string
	IMEI                string
	NetworkOperatorCode string
	NetworkOperatorName string
}

// Info will get the information of the modem
func (m *Manager) Info() *ModemInfo {
	var num string
	var access string
	var opID string
	var imei string
	var opCode string
	var opName string
	var strState string
	state, _ := m.modem.GetState()
	switch state {
	case modemmanager.MmModemStateFailed:
		strState = "⚠️ FAILED"
	case modemmanager.MmModemStateUnknown:
		strState = "❔ UNKNOWN"
	case modemmanager.MmModemStateInitializing:
		strState = "📃 INITIALIZING"
	case modemmanager.MmModemStateLocked:
		strState = "🔒 LOCKED"
	case modemmanager.MmModemStateDisabled:
		strState = "❌ DISABLED"
	case modemmanager.MmModemStateDisabling:
		strState = "⏳ DISABLING"
	case modemmanager.MmModemStateEnabling:
		strState = "⏳ ENABLING"
	case modemmanager.MmModemStateEnabled:
		strState = "✔️ ENABLED"
	case modemmanager.MmModemStateSearching:
		strState = "🔍 SEARCHING"
	case modemmanager.MmModemStateRegistered:
		strState = "✅ REGISTERED"
	case modemmanager.MmModemStateDisconnecting:
		strState = "📴 DISCONNECTING"
	case modemmanager.MmModemStateConnecting:
		strState = "⏳ CONNECTING"
	case modemmanager.MmModemStateConnected:
		strState = "🔗 CONNECTED"
	}
	manual, _ := m.modem.GetManufacturer()
	model, _ := m.modem.GetModel()
	sigPercent, _, _ := m.modem.GetSignalQuality()
	nums, _ := m.modem.GetOwnNumbers()
	if len(nums) > 0 {
		num = nums[0]
	}
	accesses, _ := m.modem.GetAccessTechnologies()
	if len(accesses) > 0 {
		switch accesses[0] {
		case modemmanager.MmModemAccessTechnologyUnknown:
			access = "🛸 UNKNOWN"
		case modemmanager.MmModemAccessTechnologyPots:
			access = "🐌 POTS"
		case modemmanager.MmModemAccessTechnologyGsm:
			access = "🐌 GSM"
		case modemmanager.MmModemAccessTechnologyGsmCompact:
			access = "🐌 GSMCOMPACT"
		case modemmanager.MmModemAccessTechnologyGprs:
			access = "🐢 GPRS"
		case modemmanager.MmModemAccessTechnologyEdge:
			access = "🚲 EDGE"
		case modemmanager.MmModemAccessTechnologyUmts:
			access = "🛵 UMTS"
		case modemmanager.MmModemAccessTechnologyHsdpa:
			access = "🚘 HSDPA"
		case modemmanager.MmModemAccessTechnologyHsupa:
			access = "🚘 HSUPA"
		case modemmanager.MmModemAccessTechnologyHspa:
			access = "🚄 HSPA"
		case modemmanager.MmModemAccessTechnologyHspaPlus:
			access = "✈️ HSPAPLUS"
		case modemmanager.MmModemAccessTechnology1xrtt:
			access = "🛵 1XRTT"
		case modemmanager.MmModemAccessTechnologyEvdo0:
			access = "🚘 EVDO0"
		case modemmanager.MmModemAccessTechnologyEvdoa:
			access = "🚘 EVDOA"
		case modemmanager.MmModemAccessTechnologyEvdob:
			access = "🚘 EVDOB"
		case modemmanager.MmModemAccessTechnologyLte:
			access = "🚀 LTE"
		}
	}
	if sim, err := m.modem.GetSim(); err == nil {
		opID, _ = sim.GetOperatorIdentifier()
	}
	if gpp, err := m.modem.Get3gpp(); err == nil {
		imei, _ = gpp.GetImei()
		opCode, _ = gpp.GetOperatorCode()
		opName, _ = gpp.GetOperatorName()
	}

	return &ModemInfo{
		Manufacturer:        emptyUnknown(manual),
		Model:               emptyUnknown(model),
		State:               emptyUnknown(strState),
		SignalQuality:       int(sigPercent),
		PhoneNumber:         emptyUnknown(num),
		AccessTech:          emptyUnknown(access),
		SimOperatorCode:     emptyUnknown(opID),
		IMEI:                emptyUnknown(imei),
		NetworkOperatorCode: emptyUnknown(opCode),
		NetworkOperatorName: emptyUnknown(opName),
	}
}

func emptyUnknown(in string) string {
	if in == "" {
		return "Unknown"
	}
	return in
}

func (m *Manager) saveAndDeleteSMS(ctx context.Context, msg modemmanager.Sms) (entSMS *ent.SMS, err error) {
	number, err := msg.GetNumber()
	if err != nil {
		return nil, fmt.Errorf("failed to get sms number: %w", err)
	}
	data, err := msg.GetData()
	if err != nil {
		return nil, fmt.Errorf("failed to get sms data: %w", err)
	}
	text, err := msg.GetText()
	if err != nil {
		return nil, fmt.Errorf("failed to get sms text: %w", err)
	}
	ts := mustGetSMSTimestamp(msg)
	tx, err := m.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to start trasaction: %w", err)
	}
	n, err := tx.SMS.Update().Where(
		sms.NumberEQ(number),
		sms.DataEQ(data),
		sms.TextEQ(text),
		sms.DischargeTimestampEQ(ts),
		sms.DeliveryStateEQ(sms.DeliveryStateRECEIVED),
	).SetUpdatedAt(time.Now()).Save(ctx)
	if err != nil {
		err = fmt.Errorf("failed to perform pessimistic lock: %w", err)
		return nil, rollback(ctx, tx, err)
	}

	if n != 0 {
		log.Printf("a same sms already exists in the db")
		entSMS, err = tx.SMS.Query().Where(
			sms.NumberEQ(number),
			sms.DataEQ(data),
			sms.TextEQ(text),
			sms.DischargeTimestampEQ(ts),
			sms.DeliveryStateEQ(sms.DeliveryStateRECEIVED),
		).Only(ctx)
		if err != nil {
			err = fmt.Errorf("failed to query the sms from db: %w", err)
			return nil, rollback(ctx, tx, err)
		}
	} else {
		entSMS, err = tx.SMS.Create().
			SetNumber(number).
			SetData(data).
			SetText(text).
			SetDischargeTimestamp(ts).
			SetDeliveryState(sms.DeliveryStateRECEIVED).
			Save(ctx)
		if err != nil {
			err = fmt.Errorf("failed to save sms into db: %w", err)
			return nil, rollback(ctx, tx, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to save sms into the db: %w", err)
	}

	if err := m.msg.Delete(msg); err != nil {
		return entSMS, fmt.Errorf("sms saved, but failed to delete from device storage: %w", err)
	}
	return entSMS, nil
}

func rollback(ctx context.Context, tx *ent.Tx, err error) error {
	if rerr := tx.Rollback(); rerr != nil {
		return fmt.Errorf("failed to rollback after error(%v): %w", err, rerr)
	}
	return err
}

func mustGetSMSTimestamp(msg modemmanager.Sms) time.Time {
	ts, err := msg.GetDischargeTimestamp()
	if err != nil {
		ts, err = msg.GetTimestamp()
		if err != nil {
			ts = time.Now()
		}
	}

	if ts.Equal(time.Unix(0, 0)) || ts.IsZero() {
		ts = time.Now()
	}
	return ts
}
