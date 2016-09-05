package transform

import (
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

func (cbc *ComposeBuildContext) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(*cbc)
	if err != nil {
		var ctx string
		err = unmarshal(&ctx)
		if err != nil {
			return err
		}
		cbc.Context = ctx
	}
	return nil
}

type ComposeBuildContext struct {
	Context    string            `yaml:"context" `
	Dockerfile string            `yaml:"dockerfile" `
	Args       map[string]string `yaml:"args" `
}

func (c ComposeContainer) IngestBuild() *BuildContext {
	if c.Build != nil {
		bc := &BuildContext{
			Context:    c.Build.Context,
			Dockerfile: c.Build.Dockerfile,
			Args:       c.Build.Args,
		}
		return bc
	}
	return nil
}

func (c *ComposeContainer) EmitBuild(in *BuildContext) {
	if in != nil {
		c.Build = &ComposeBuildContext{
			Context:    in.Context,
			Dockerfile: in.Dockerfile,
			Args:       in.Args,
		}
	}
}

func parseComposePortMapping(line string) *PortMapping {
	pm := &PortMapping{Protocol: "tcp"}
	if strings.HasSuffix(line, "/udp") {
		line = strings.TrimSuffix(line, "/udp")
		pm.Protocol = "udp"
	}
	parts := strings.Split(line, ":")
	if len(parts) == 1 {
		port, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
		pm.ContainerPort = port
	} else if len(parts) == 2 {
		port, err := strconv.Atoi(parts[0])
		if err != nil {
			panic(err)
		}
		pm.HostPort = port
		port, err = strconv.Atoi(parts[1])
		if err != nil {
			panic(err)
		}
		pm.ContainerPort = port
	} else if len(parts) == 3 {
		pm.HostIP = parts[0]
		port, err := strconv.Atoi(parts[1])
		if err != nil {
			panic(err)
		}
		pm.HostPort = port
		port, err = strconv.Atoi(parts[2])
		if err != nil {
			panic(err)
		}
		pm.ContainerPort = port
	}
	return pm
}

func (c ComposeContainer) IngestPortMappings() *PortMappings {
	if len(c.PortMappings) > 0 {
		response := PortMappings{}
		for _, pm := range c.PortMappings {
			response = append(response, *parseComposePortMapping(pm))
		}
		return &response
	}
	return nil
}

func (c *ComposeContainer) EmitPortMappings(mappings *PortMappings) {
	if mappings == nil {
		return
	}
	output := []string{}
	for _, mapping := range *mappings {
		portStr := []string{}
		if mapping.HostPort > 0 {
			portStr = append(portStr, mapping.HostIP)

		}
		if mapping.HostPort > 0 {
			portStr = append(portStr, strconv.Itoa(mapping.HostPort))
		}
		if mapping.ContainerPort > 0 {
			portStr = append(portStr, strconv.Itoa(mapping.ContainerPort))
		}

		portData := strings.Trim(strings.Join(portStr, ":"), ":")
		if len(portData) > 0 {
			if strings.Compare(mapping.Protocol, "udp") == 0 {
				portData = portData + "/udp"
			}
			output = append(output, portData)
		}
	}
	if len(output) > 0 {
		c.PortMappings = output
	}
}

type ComposeLogging struct {
	Driver  string
	Options map[string]string
}

func (c ComposeContainer) IngestLogging() *Logging {
	if c.Logging != nil {
		return &Logging{
			Driver:  c.Logging.Driver,
			Options: c.Logging.Options,
		}
	}
	return nil
}

func (c *ComposeContainer) EmitLogging(l *Logging) {
	if l != nil {
		c.Logging = &ComposeLogging{
			Driver:  l.Driver,
			Options: l.Options,
		}
	}
}

func (ckv *ComposeKV) UnmarshalYAML(unmarshal func(interface{}) error) error {
	err := unmarshal(&ckv.Values)
	if err != nil {
		var keyValues []string
		err = unmarshal(&keyValues)
		if err != nil {
			return err
		}
		response := map[string]string{}
		for _, kvString := range keyValues {
			parts := strings.SplitN(kvString, "=", 2)
			if len(parts) > 1 {
				response[parts[0]] = parts[1]
			} else {
				response[parts[0]] = ""
			}
		}
		ckv.Values = response
	}
	return nil
}

// ComposeKV is a special type for Labels and Environment variables
// since compose allows "k=v" and "k: v" formats
type ComposeKV struct {
	Values map[string]string
}

func (c ComposeContainer) IngestVolumes() *IntermediateVolumes {
	if len(c.Volumes) > 0 {
		response := IntermediateVolumes{}
		for _, vol := range c.Volumes {
			iv := IntermediateVolume{ReadOnly: false}
			parts := strings.Split(vol, ":")
			if len(parts) == 1 {
				iv.Container = parts[0]
			} else if len(parts) == 2 {
				iv.Host = parts[0]
				iv.Container = parts[1]
			} else if len(parts) == 3 {
				iv.Host = parts[0]
				iv.Container = parts[1]
				if parts[2] == "ro" {
					iv.ReadOnly = true
				}
			}
			response = append(response, iv)
		}
		return &response
	}
	return nil
}

func (c *ComposeContainer) EmitVolumes(vols *IntermediateVolumes) {
	if vols == nil {
		return
	}
	output := []string{}
	for _, volume := range *vols {
		readOnly := ""
		if volume.ReadOnly {
			readOnly = "ro"
		}
		volStr := []string{volume.Host, volume.Container, readOnly}
		volData := strings.Trim(strings.Join(volStr, ":"), ":")
		output = append(output, volData)
	}

	if len(output) > 0 {
		c.Volumes = output
	}
}

type ComposeContainer struct {
	Build       *ComposeBuildContext `yaml:"build,omitempty" `
	Command     string               `yaml:"command,omitempty" `
	CPU         int                  `yaml:"cpu_shares,omitempty" `
	DNS         []string             `yaml:"dns,omitempty" `
	Domain      []string             `yaml:"dns_search,omitempty" `
	Entrypoint  string               `yaml:"entrypoint,omitempty" `
	EnvFile     []string             `yaml:"env_file,omitempty" `
	Environment ComposeKV            `yaml:"environment,omitempty" `
	Expose      []int                `yaml:"expose,omitempty" `
	Hostname    string               `yaml:"hostname,omitempty" `
	Image       string               `yaml:"image,omitempty" `
	Labels      ComposeKV            `yaml:"labels,omitempty" `
	Links        []string             `yaml:"links,omitempty" `
	Logging      *ComposeLogging      `yaml:"logging,omitempty" `
	Memory       int                  `yaml:"mem_limit,omitempty" `
	Name         string               `yaml:"-" `
	Network      []string             `yaml:"networks,omitempty" `
	NetworkMode  string               `yaml:"network_mode,omitempty" `
	Pid          string               `yaml:"pid,omitempty" `
	PortMappings []string             `yaml:"ports,omitempty" `
	Privileged   bool                 `yaml:"privileged,omitempty" `
	User         string               `yaml:"user,omitempty" `
	Volumes      []string             `yaml:"volumes,omitempty" `
	VolumesFrom  []string             `yaml:"volumes_from,omitempty" `
	WorkDir      string               `yaml:"working_dir,omitempty" `
}

type ComposeFormat struct {
	Version  string                       `yaml:"version"`
	Services map[string]*ComposeContainer `yaml:"services" `
}

func (f ComposeFormat) IngestContainers(input io.ReadCloser) (*BasePodData, error) {

	body, err := ioutil.ReadAll(input)
	defer input.Close()
	if err != nil && err != io.EOF {
		return nil, err
	}
	cf := &ComposeFormat{}
	err = yaml.Unmarshal(body, cf)
	if err != nil {
		return nil, err
	}

	outputPod := BasePodData{}
	outputPod.Containers = []*BaseContainerFormat{}

	for serviceName, container := range cf.Services {

		ir := BaseContainerFormat{}
		outputPod.Containers = append(outputPod.Containers, &ir)

		ir.Build = container.IngestBuild()
		ir.Command = container.Command
		ir.CPU = container.CPU
		ir.DNS = container.DNS
		ir.Domain = container.Domain
		ir.Entrypoint = container.Entrypoint
		ir.EnvFile = container.EnvFile
		ir.Environment = container.Environment.Values
		ir.Expose = container.Expose
		ir.Hostname = container.Hostname
		ir.Image = container.Image
		ir.Labels = container.Labels.Values
		ir.Links = container.Links
		ir.Logging = container.IngestLogging()
		ir.Memory = container.Memory
		ir.Name = serviceName
		ir.Network = container.Network
		ir.NetworkMode = container.NetworkMode
		ir.Pid = container.Pid
		ir.PortMappings = container.IngestPortMappings()
		ir.Privileged = container.Privileged
		ir.User = container.User
		ir.Volumes = container.IngestVolumes()
		ir.VolumesFrom = container.VolumesFrom
		ir.WorkDir = container.WorkDir
	}
	return &outputPod, nil
}

func (f ComposeFormat) EmitContainers(input *BasePodData) ([]byte, error) {
	output := &ComposeFormat{Version: "2"}
	output.Services = map[string]*ComposeContainer{}

	for _, container := range input.Containers {
		composeContainer := ComposeContainer{}
		output.Services[container.Name] = &composeContainer

		composeContainer.EmitBuild(container.Build)
		composeContainer.Command = container.Command
		composeContainer.CPU = container.CPU
		composeContainer.DNS = container.DNS
		composeContainer.Domain = container.Domain
		composeContainer.Entrypoint = container.Entrypoint
		composeContainer.EnvFile = container.EnvFile
		composeContainer.Environment = ComposeKV{Values: container.Environment}
		composeContainer.Expose = container.Expose
		composeContainer.Hostname = container.Hostname
		composeContainer.Image = container.Image
		composeContainer.Labels = ComposeKV{Values: container.Labels}
		composeContainer.Links = container.Links
		composeContainer.EmitLogging(container.Logging)
		composeContainer.Memory = container.Memory
		composeContainer.Network = container.Network
		composeContainer.NetworkMode = container.NetworkMode
		composeContainer.Pid = container.Pid
		composeContainer.EmitPortMappings(container.PortMappings)
		composeContainer.Privileged = container.Privileged
		composeContainer.User = container.User
		composeContainer.EmitVolumes(container.Volumes)
		composeContainer.VolumesFrom = container.VolumesFrom
		composeContainer.WorkDir = container.WorkDir
	}
	return yaml.Marshal(output)
}
