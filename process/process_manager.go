package process

import (
	"daemon_demo/config"
	"fmt"
	"github.com/CodisLabs/codis/pkg/utils/errors"
	"sync"
	"syscall"
)

type Manager struct {
	sync.RWMutex
	procs map[string]*Process
}

func New() *Manager {
	return &Manager{
		procs: make(map[string]*Process),
	}
}

func (p *Manager) ProcessCreate(name string, config *config.ProcessConfig) *Process {
	p.Lock()
	defer p.Unlock()
	proc, ok := p.procs[name]
	if !ok {
		proc = newProcess(name, config)
		p.procs[name] = proc
		go proc.Run()
	}
	return proc
}

func (p *Manager) ProcessRemove(name string) {
	p.Lock()
	defer p.Unlock()
	proc, ok := p.procs[name]
	if ok {
		delete(p.procs, name)
		proc.Stop()
	}
}

// FIXME 管理进程退出，是否所有管理进程应该退出？
func (p *Manager) Stop() error {
	p.Lock()
	defer p.Unlock()
	for _, proc := range p.procs {
		proc.Stop()
	}
	p.procs = make(map[string]*Process)
	return nil
}

func (p *Manager) ProcessStart(name string) {
	p.Lock()
	defer p.Unlock()
	proc, ok := p.procs[name]
	if ok && proc.IsUserStopped() {
		proc.Up()
	}
}

func (p *Manager) ProcessStop(name string) {
	p.Lock()
	defer p.Unlock()
	proc, ok := p.procs[name]
	if ok {
		proc.Down()
	}
}

func (p *Manager) ProcessStatus() map[string]string {
	rt := make(map[string]string)
	p.RLock()
	defer p.RUnlock()
	for pname, proc := range p.procs {
		rt[pname] = proc.GetStatus().String()
	}
	return rt
}

func (p *Manager) FwdSignal(name string, sig syscall.Signal) error {
	p.RLock()
	defer p.RUnlock()
	proc, ok := p.procs[name]
	if ok {
		return proc.Signal(sig)
	}
	return errors.New(fmt.Sprint("not found process:", name))
}
