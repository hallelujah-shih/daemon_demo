package process

import (
	"daemon_demo/config"
	"errors"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

type ProcessState int

const (
	STOPPED ProcessState = iota
	STARTING
	RUNNING
	EXITED
	FATAL
)

func (p ProcessState) String() string {
	switch p {
	case STOPPED:
		return "STOPPED"
	case STARTING:
		return "STARTING"
	case RUNNING:
		return "RUNNING"
	case EXITED:
		return "EXITED"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

type Process struct {
	sync.RWMutex
	name          string
	config        *config.ProcessConfig
	cmd           *exec.Cmd
	state         ProcessState
	startTimeUnix int64
	stopTimeUnix  int64
	userStop      bool
	exit          int32
}

func newProcess(name string, config *config.ProcessConfig) *Process {
	return &Process{
		name:   name,
		config: config,
	}
}

func (p *Process) Run() {
	for atomic.LoadInt32(&p.exit) == 0 {
		if p.IsRunning() {
			time.Sleep(1 * time.Second)
			continue
		}
		if p.IsUserStopped() {
			time.Sleep(2 * time.Second)
			continue
		}
		p.start()
	}
	p.Signal(syscall.SIGKILL)
}

func (p *Process) start() {
	p.Lock()
	switch p.state {
	case STOPPED:
		fallthrough
	case EXITED:
		fallthrough
	case FATAL:
		p.state = STARTING
	default:
		p.Unlock()
		return
	}
	p.Unlock()

	go func() (err error) {
		var (
			errLog, outLog io.Writer
		)
		errLog, err = os.OpenFile(p.config.StderrPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			logrus.Warnln("process:", p.name, "stderr log error:", err)
			errLog = ioutil.Discard
		}
		outLog, err = os.OpenFile(p.config.StdoutPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			logrus.Warnln("process:", p.name, "stdout log error:", err)
			outLog = ioutil.Discard
		}

		defer func() {
			if f, ok := errLog.(*os.File); ok {
				f.Close()
			}
			if f, ok := outLog.(*os.File); ok {
				f.Close()
			}
			logrus.Debugln("process:", p.name, "start:", p.startTimeUnix, "stop:", p.stopTimeUnix, "reason:", err)
		}()

		p.cmd = exec.Command(p.config.CommandName, p.config.CommandArgs...)
		// set log TODO:可以增加 multiwriter用于查看状态的时候显示最近的日志
		p.cmd.Stdout = outLog
		p.cmd.Stderr = errLog

		// setuidgid
		if p.config.UidGid != "" {
			if uid, gid, err := p.getUidGid(p.config.UidGid); err != nil {
				logrus.Errorln("process:", p.name, "get uid gid:", p.config.UidGid, "error:", err)
			} else {
				logrus.Debugln("process:", p.name, "user:", p.config.UidGid, "uid:", uid, "gid:", gid)
				p.cmd.SysProcAttr = &syscall.SysProcAttr{}
				p.cmd.SysProcAttr.Credential = &syscall.Credential{Uid: uint32(uid), Gid: uint32(gid)}
			}
		}
		// env
		p.cmd.Env = p.mergeUserEnvs(p.config.Envs)

		// TODO chdir /
		// TODO 需要实现rlimit操作

		if err = p.cmd.Start(); err != nil {
			logrus.Errorln("process:", p.name, "start fail:", err)
			time.Sleep(5 * time.Second)
			p.Lock()
			p.state = FATAL
			p.Unlock()
			return
		}

		logrus.Debugln("process:", p.name, "running")
		p.Lock()
		p.state = RUNNING
		p.startTimeUnix = time.Now().Unix()
		p.Unlock()

		err = p.cmd.Wait()
		p.Lock()
		if p.cmd.Process != nil && p.cmd.ProcessState != nil {
			p.state = EXITED
		} else {
			p.state = FATAL
		}
		p.stopTimeUnix = time.Now().Unix()
		p.Unlock()
		return
	}()
}

func (p *Process) Stop() {
	atomic.StoreInt32(&p.exit, 1)
}

func (p *Process) Up() {
	p.Lock()
	defer p.Unlock()
	if p.userStop {
		p.userStop = false
	}
	logrus.Debugln("process:", p.name, "up")
}

func (p *Process) Down() {
	p.Lock()
	ustop := p.userStop
	p.userStop = true
	p.Unlock()

	if ustop {
		return
	}

	errTerm := p.Signal(syscall.SIGTERM)
	errCond := p.Signal(syscall.SIGCONT)
	logrus.Debugln("process:", p.name, "down:", errTerm, errCond)
}

func (p *Process) Signal(sig syscall.Signal) error {
	p.Lock()
	defer p.Unlock()
	if p.cmd != nil && p.cmd.Process != nil {
		logrus.Debugln("fwd signal:", sig, "to process:", p.name)
		return p.cmd.Process.Signal(sig)
	} else {
		return errors.New("process not running")
	}
}

func (p Process) getEnvs(envs []string) map[string]string {
	rt := make(map[string]string)
	for _, env := range envs {
		if index := strings.Index(env, "="); index == -1 {
			continue
		} else {
			rt[env[:index]] = env[index+1:]
		}
	}
	return rt
}

func (p *Process) mergeUserEnvs(envs []string) []string {
	rt := []string{}
	sysEnvs := p.getEnvs(os.Environ())
	userEnvs := p.getEnvs(envs)
	for key, value := range userEnvs {
		sysEnvs[key] = value
	}
	for key, value := range sysEnvs {
		rt = append(rt, strings.Join([]string{key, value}, "="))
	}
	return rt
}

func (p Process) getUidGid(username string) (uid int, gid int, err error) {
	u, err := user.Lookup(username)
	if err != nil {
		return
	}
	uid, err = strconv.Atoi(u.Uid)
	if err != nil {
		return
	}
	gid, err = strconv.Atoi(u.Gid)
	if err != nil {
		return
	}
	return
}

func (p *Process) IsRunning() bool {
	p.RLock()
	defer p.RUnlock()
	if p.cmd != nil && p.cmd.Process != nil {
		return p.cmd.Process.Signal(syscall.Signal(0)) == nil
	}
	return false
}

func (p *Process) IsUserStopped() bool {
	p.RLock()
	defer p.RUnlock()
	return p.userStop
}

func (p Process) GetStatus() ProcessState {
	p.RLock()
	defer p.RUnlock()
	return p.state
}

func (p Process) GetStartTime() int64 {
	p.RLock()
	defer p.RUnlock()
	return p.startTimeUnix
}

func (p Process) GetStopTime() int64 {
	p.RLock()
	defer p.RUnlock()
	return p.stopTimeUnix
}
