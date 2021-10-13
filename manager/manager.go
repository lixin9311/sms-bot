package manager

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/lixin9311/backoff/v2"
	"github.com/lixin9311/sms-bot/ent"
	"github.com/lixin9311/sms-bot/ent/sms"
	"github.com/maltegrosse/go-modemmanager"
)

type EventType string

const (
	SMSEvent         EventType = "SMS"
	PhoneCallEvent   EventType = "PhoneCall"
	ErrorEvent       EventType = "Error"
	ModemAddEvent    EventType = "ModemAdded"
	ModemRemoveEvent EventType = "ModemRemoved"
)

type PhoneCall struct {
	ModemID string
	Caller  string
}

type Event struct {
	Type    EventType
	Payload interface{} // ent.SMS, PhoneCall, error, string
}

// Manager ...
type Manager struct {
	ctx     context.Context
	mapLock sync.RWMutex
	modems  map[string]modemmanager.Modem
	cancels map[string]context.CancelFunc
	eventCh chan Event
	client  *ent.Client
}

// NewManager ...
func NewManager(ctx context.Context, client *ent.Client) (*Manager, error) {
	man := &Manager{
		ctx:     ctx,
		client:  client,
		modems:  map[string]modemmanager.Modem{},
		cancels: map[string]context.CancelFunc{},
		eventCh: make(chan Event, 10),
	}
	err := man.init(ctx)
	return man, err
}

func (m *Manager) init(ctx context.Context) error {
	mmgr, err := newMMGR(ctx)
	if err != nil {
		return err
	}
	mmgr.SetLogging(modemmanager.MMLoggingLevelDebug)
	modems, err := m.scanModems(mmgr)
	if err != nil {
		return fmt.Errorf("unable to scan modems: %v", err)
	}
	for _, modem := range modems {
		if err := m.addModem(modem); err != nil {
			log.Printf("unable to add modem: %v", err)
			m.dispatchErr(fmt.Errorf("unable to add modem: %v", err))
		}
	}
	ch := mmgr.SubscribePropertiesChanged()
	go func() {
		lastScan := time.Now()
		tick := time.NewTicker(time.Minute)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				now := time.Now()
				if now.Sub(lastScan) < 30*time.Second {
					continue
				}
				lastScan = now
				if modems, err := m.scanModems(mmgr); err != nil {
					m.dispatchErr(fmt.Errorf("unable to scan modem: %v", err))
				} else {
					for _, modem := range modems {
						if err := m.addModem(modem); err != nil {
							log.Printf("unable to add modem: %v", err)
							m.dispatchErr(fmt.Errorf("unable to add modem: %v", err))
						}
					}
				}
			case <-ch:
				now := time.Now()
				if now.Sub(lastScan) < 30*time.Second {
					continue
				}
				lastScan = now
				if modems, err := m.scanModems(mmgr); err != nil {
					m.dispatchErr(fmt.Errorf("unable to scan modem: %v", err))
				} else {
					for _, modem := range modems {
						if err := m.addModem(modem); err != nil {
							log.Printf("unable to add modem: %v", err)
							m.dispatchErr(fmt.Errorf("unable to add modem: %v", err))
						}
					}
				}
			}
		}
	}()
	return nil
}

func newMMGR(ctx context.Context) (mmgr modemmanager.ModemManager, err error) {
	bo := backoff.Backoff{}
	for i := 0; i < 10; i++ {
		mmgr, err = modemmanager.NewModemManager()
		if err == nil {
			return mmgr, err
		}
		if err := backoff.Sleep(ctx, bo.Backoff(i)); err != nil {
			return nil, err
		}
	}
	return
}

func (m *Manager) scanModems(mmgr modemmanager.ModemManager) (modems []modemmanager.Modem, err error) {
	bo := backoff.Backoff{
		BaseDelay:  time.Second,
		Multiplier: 1.6,
		Jitter:     0.2,
	}
	for i := 0; 1 < 10; i++ {
		err = mmgr.ScanDevices()
		if err != nil {
			err = fmt.Errorf("failed to scan devices: %w", err)
			log.Println(err)
			time.Sleep(bo.Backoff(i))
			continue
		}
		modems, err = mmgr.GetModems()
		if err != nil {
			err = fmt.Errorf("failed to get modems: %w", err)
			log.Println(err)
			time.Sleep(bo.Backoff(i))
			continue
		}
		return modems, nil
	}
	return
}

func (m *Manager) cancel(id string) {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()

	if cancel := m.cancels[id]; cancel != nil {
		cancel()
	}
}

func (m *Manager) exists(id string) bool {
	m.mapLock.RLock()
	defer m.mapLock.RUnlock()
	_, ok := m.modems[id]
	return ok
}

func (m *Manager) addModem(modem modemmanager.Modem) error {
	id, err := getModemID(modem)
	if err != nil {
		m.dispatchErr(err)
		return err
	}
	if m.exists(id) {
		return nil
	}

	m.mapLock.Lock()
	defer m.mapLock.Unlock()

	m.modems[id] = modem
	ctx, cancel := context.WithCancel(m.ctx)
	m.cancels[id] = cancel
	go m.subscribeModem(ctx, id, modem)
	m.eventCh <- Event{
		Type:    ModemAddEvent,
		Payload: id,
	}
	if smss, err := m.LoadSMS(ctx, id, modem); err != nil {
		err := fmt.Errorf("modem[%s] unable to load existing smss: %w", id, err)
		log.Println(err)
		m.dispatchErr(err)
	} else {
		for _, sms := range smss {
			m.eventCh <- Event{
				Type:    SMSEvent,
				Payload: sms,
			}
		}
	}
	return nil
}

func (m *Manager) ListModems() []string {
	m.mapLock.RLock()
	defer m.mapLock.RUnlock()
	result := make([]string, 0, len(m.modems))
	for k := range m.modems {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

func (m *Manager) GetModemInfo(id string) *ModemInfo {
	m.mapLock.RLock()
	defer m.mapLock.RUnlock()
	modem, ok := m.modems[id]
	if !ok {
		return nil
	}
	return modemInfo(modem)
}

func (m *Manager) subscribeModem(ctx context.Context, id string, modem modemmanager.Modem) {
	defer m.DelModem(id)
	msg, err := modem.GetMessaging()
	if err != nil {
		log.Printf("modem[%s] failed to get messaging endpoint: %v", id, err)
		m.dispatchErr(fmt.Errorf("modem[%s] failed to get messaging endpoint: %w", id, err))
		return
	}
	smsch := msg.SubscribeAdded()
	defer msg.Unsubscribe()
	var voiceCh <-chan *dbus.Signal
	voice, err := modem.GetVoice()
	if err != nil {
		log.Printf("modem[%s] failed to get voice endpoint: %v", id, err)
		m.dispatchErr(fmt.Errorf("modem[%s] failed to get voice endpoint: %w", id, err))
	} else {
		voiceCh = voice.SubscribeCallAdded()
		defer voice.Unsubscribe()
	}
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// check modem state
			state, err := modem.GetState()
			if err != nil {
				err = fmt.Errorf("modem[%s] failed to get state, it may disconnected: %w", id, err)
				log.Println(err)
				m.dispatchErr(err)
				return
			}
			if state == modemmanager.MmModemStateUnknown || state == modemmanager.MmModemStateFailed {
				err = fmt.Errorf("modem[%s] unknown state, it may disconnected: %w", id, err)
				log.Println(err)
				m.dispatchErr(err)
				return
			}
		case <-ctx.Done():
			return
		case sig := <-voiceCh:
			if sig == nil {
				return
			}
			call, err := voice.ParseCallAdded(sig)
			if err != nil {
				// just ignore
				// err = fmt.Errorf("modem[%s] failed to handle phone call added event: %w", id, err)
				// log.Println(err)
				// m.dispatchErr(err)
				continue
			}
			direction, err := call.GetDirection()
			if err != nil {
				err = fmt.Errorf("modem[%s] failed to get call direction: %w", id, err)
				log.Println(err)
				m.dispatchErr(err)
				continue
			}
			if direction == modemmanager.MmCallDirectionIncoming {
				caller, _ := call.GetNumber()
				if caller == "" {
					caller = "📞 Unknown caller"
				}
				m.eventCh <- Event{
					Type: PhoneCallEvent,
					Payload: &PhoneCall{
						ModemID: id,
						Caller:  caller,
					},
				}
			}
		case sig := <-smsch:
			if sig == nil {
				return
			}
			sms, received, err := msg.ParseAdded(sig)
			if err != nil {
				// just ignore
				// err = fmt.Errorf("modem[%s] failed to handle sms added event: %w", id, err)
				// log.Println(err)
				// m.dispatchErr(err)
				continue
			}
			if !received {
				continue
			}
			state, err := sms.GetState()
			if err != nil {
				err = fmt.Errorf("modem[%s] failed to get sms state: %w", id, err)
				log.Print(err)
				m.dispatchErr(err)
				continue
			}
			if state == modemmanager.MmSmsStateReceived {
				entSMS, err := m.saveSMS(id, sms)
				if err != nil {
					m.dispatchErr(err)
				} else {
					m.eventCh <- Event{
						Type:    SMSEvent,
						Payload: entSMS,
					}
				}
			} else if state == modemmanager.MmSmsStateReceiving {
				go m.handleAsyncSMS(id, msg, sms)
			}
		}
	}
}

func (m *Manager) DelModem(id string) {
	m.mapLock.Lock()
	defer m.mapLock.Unlock()
	delete(m.modems, id)
	m.eventCh <- Event{
		Type:    ModemRemoveEvent,
		Payload: id,
	}
}

func getModemID(modem modemmanager.Modem) (string, error) {
	if number, err := getNumber(modem); err == nil {
		return number, nil
	} else if imsi, err := getIMSI(modem); err == nil {
		return imsi, nil
	} else if imei, err := getIMEI(modem); err == nil {
		return imei, nil
	}
	return "", fmt.Errorf("unable to determin modem id")
}

func getNumber(modem modemmanager.Modem) (string, error) {
	number, err := modem.GetOwnNumbers()
	if err != nil {
		return "", err
	} else if len(number) == 0 || number[0] == "" {
		return "", fmt.Errorf("unknown number")
	}
	return number[0], nil
}

func getIMSI(modem modemmanager.Modem) (string, error) {
	sim, err := modem.GetSim()
	if err != nil {
		return "", nil
	}
	imsi, err := sim.GetImsi()
	if err != nil {
		return "", nil
	} else if imsi == "" {
		return "", fmt.Errorf("unknown imsi")
	}
	return imsi, nil
}

func getIMEI(modem modemmanager.Modem) (string, error) {
	gpp, err := modem.Get3gpp()
	if err != nil {
		return "", nil
	}
	imei, err := gpp.GetImei()
	if err != nil {
		return "", err
	} else if imei == "" {
		return "", fmt.Errorf("unknown imei")
	}
	return imei, err
}

func (m *Manager) ListSMSModemIDs(ctx context.Context) ([]string, error) {
	return m.client.SMS.Query().GroupBy(sms.FieldModemID).Strings(ctx)
}

// List ...
func (m *Manager) ListSMS(ctx context.Context, modemid string, offset, limit int) ([]*ent.SMS, int, error) {
	if limit == 0 {
		limit = 100
	}
	q := m.client.SMS.Query().Order(ent.Desc(sms.FieldDischargeTimestamp))
	if modemid != "*" && modemid != "" {
		q.Where(sms.ModemIDEQ(modemid))
	}
	countQ := q.Clone()
	results, err := q.Limit(limit).Offset(offset).All(ctx)
	total := countQ.CountX(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list sms from db: %w", err)
	}
	return results, total, nil
}

// LoadSMS ...
func (m *Manager) LoadSMS(ctx context.Context, id string, modem modemmanager.Modem) ([]*ent.SMS, error) {
	msg, err := modem.GetMessaging()
	if err != nil {
		return nil, err
	}
	smss, err := msg.List()
	if err != nil {
		return nil, fmt.Errorf("failed to list smss: %w", err)
	}
	results := make([]*ent.SMS, 0, len(smss))
	for _, v := range smss {
		state, err := v.GetState()
		if err != nil {
			return nil, fmt.Errorf("failed to get sms state: %w", err)
		}
		if state == modemmanager.MmSmsStateReceiving {
			go m.handleAsyncSMS(id, msg, v)
			continue
		}
		if state == modemmanager.MmSmsStateReceived {
			entSMS, err := m.saveSMS(id, v)
			if err != nil {
				return nil, err
			}
			msg.Delete(v)
			results = append(results, entSMS)
		}
	}
	return results, nil
}

func (m *Manager) handleAsyncSMS(modemID string, msg modemmanager.ModemMessaging, sms modemmanager.Sms) {
	maxRetries := 10
	succ := false
	bo := backoff.Backoff{
		BaseDelay: 10 * time.Second,
		Jitter:    0.1,
	}
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
		dur := bo.Backoff(retries)
		time.Sleep(dur)
	}
	if succ {
		entSMS, err := m.saveSMS(modemID, sms)
		if err != nil {
			log.Printf("failed to save sms: %v", err)
			m.dispatchErr(fmt.Errorf("failed to save sms: %w", err))
			return
		}
		if err := msg.Delete(sms); err != nil {
			log.Printf("failed to delete sms: %v", err)
		}
		m.eventCh <- Event{
			Type:    SMSEvent,
			Payload: entSMS,
		}
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
		m.dispatchErr(err)

		log.Println("deleting the sms")
		if err := msg.Delete(sms); err != nil {
			log.Printf("failed to delete sms: %v", err)
		}
	}
}

func (m *Manager) dispatchErr(err error) {
	m.eventCh <- Event{
		Type:    ErrorEvent,
		Payload: err,
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
func modemInfo(modem modemmanager.Modem) *ModemInfo {
	var num string
	var access string
	var opID string
	var imei string
	var opCode string
	var opName string
	var strState string
	state, _ := modem.GetState()
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
	default:
		strState = "⁉️ UNKNOWN"
	}
	manual, _ := modem.GetManufacturer()
	model, _ := modem.GetModel()
	sigPercent, _, _ := modem.GetSignalQuality()
	nums, _ := modem.GetOwnNumbers()
	if len(nums) > 0 {
		num = nums[0]
	}
	accesses, _ := modem.GetAccessTechnologies()
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
	if sim, err := modem.GetSim(); err == nil {
		opID, _ = sim.GetOperatorIdentifier()
	}
	if gpp, err := modem.Get3gpp(); err == nil {
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
		return "⁉️ Unknown"
	}
	return in
}

func (m *Manager) saveSMS(modemID string, msg modemmanager.Sms) (entSMS *ent.SMS, err error) {
	ctx := context.Background()
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
		sms.ModemIDEQ(modemID),
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
			SetModemID(modemID).
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

func (m *Manager) Subscribe() <-chan Event {
	if m == nil {
		return make(chan Event)
	}
	return m.eventCh
}
