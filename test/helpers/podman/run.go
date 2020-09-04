package podman

import (
	"encoding/json"
	"strings"

	"github.com/openshift/cluster-logging-operator/pkg/logger"
)

type ErrorGetter interface {
	Error() error
}

type PodState string

const (
	PodUnkown  PodState = "Unknown"
	PodRunning PodState = "Running"
	PodStopped PodState = "Stopped"
)

type PodCommand interface {
	Run() PodCommand
	WithImage(image string) PodCommand
	AddVolume(hostPath, containerPath string) PodCommand

	Create() PodCommand
	AddContainer(image string, args ...string) PodCommand
	Remove() ErrorGetter

	State() PodState
}

func Pod(name string) PodCommand {
	p := &pod{
		name:         name,
		volumes:      map[string][]string{},
		containerIDs: []string{},
	}
	return p
}

type pod struct {
	name         string
	podID        string
	containerIDs []string
	err          error
	image        string
	volumes      map[string][]string
}

func (p *pod) State() PodState {
	out := p.run("pod", "inspect", p.image)
	inspect := map[string]interface{}{}
	if p.err = json.Unmarshal([]byte(out), &inspect); p.err == nil {
		if state, found := inspect["State"]; found {
			return state.(PodState)
		}
	}
	return PodUnkown
}

func (p *pod) Error() error {
	return p.err
}
func (p *pod) WithImage(image string) PodCommand {
	p.image = image
	return p
}
func (p *pod) AddVolume(hostPath, containerPath string) PodCommand {
	if _, ok := p.volumes[hostPath]; !ok {
		p.volumes[hostPath] = []string{}
	}
	p.volumes[hostPath] = append(p.volumes[hostPath], containerPath)
	return p
}
func (p *pod) Run() PodCommand {
	args := []string{}
	for host, targets := range p.volumes {
		for _, target := range targets {
			args = append(args, "-v", host+":"+target)
		}
	}
	args = append(args, "--pod", "new:"+p.name, p.image)
	p.containerIDs = append(p.containerIDs, p.run("run", "-d", args...))
	return p
}

func (p *pod) AddContainer(image string, args ...string) PodCommand {
	args = append([]string{"--pod", p.podID, image}, args...)
	p.containerIDs = append(p.containerIDs, p.run("run", "-d", args...))
	return p
}
func (p *pod) Create() PodCommand {
	p.podID = p.run("pod", "create")
	return p
}
func (p *pod) Remove() ErrorGetter {
	p.run("pod", "rm", p.name, "--force")
	return p
}

func (p *pod) run(cmd, subcmd string, args ...string) string {

	r := runner{
		collectArgsFunc: func() []string {
			args := append([]string{cmd, subcmd}, args...)
			return sanitizeArgs(strings.Join(args, " "))
		},
	}
	logger.Infof("run args: %v", r.collectArgsFunc())
	out, err := r.Run()
	p.err = err
	return out
}
