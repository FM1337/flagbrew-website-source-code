package daemon

import (
	"net"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
)

type SettingDaemon struct {
	enabled bool
	runtime time.Duration
	ip      string
	home    string
}

func NewSettingDaemon(home string) *SettingDaemon {
	return &SettingDaemon{enabled: false, runtime: 1 * time.Hour, home: home}
}

func (d *SettingDaemon) Start() {
	d.enabled = true
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if !d.enabled {
				break
			}
			mutex.Lock()
			d.lookUpAllensIP()
			mutex.Unlock()
			time.Sleep(d.runtime)
		}
	}()
}

func (d *SettingDaemon) lookUpAllensIP() {
	ips, err := net.LookupIP(d.home)
	if err != nil {
		helpers.LogToSentry(err)
	}

	for _, ip := range ips {
		if ipv4 := ip.To4(); ipv4 != nil {
			d.ip = ipv4.String()
		}
	}
}

func (d *SettingDaemon) GetOwnerIP() string {
	return d.ip
}

func (d *SettingDaemon) Stop() {
	d.enabled = false
}
