package main

import (
	"fmt"

	"reflect"
	"regexp"

	"github.com/rancher/norman/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// PullAlways means that kubelet always attempts to pull the latest image. Container will fail If the pull fails.
	PullAlways PullPolicy = "Always"
	// PullNever means that kubelet never pulls an image, but only uses a local image. Container will fail if the image isn't present
	PullNever PullPolicy = "Never"
	// PullIfNotPresent means that kubelet pulls if the image isn't present on disk. Container will fail if the image isn't present and the pull fails.
	PullIfNotPresent PullPolicy = "IfNotPresent"
)

const (
	Server          WorkloadType = "Server"
	SingletonServer WorkloadType = "SingletonServer"
	Worker          WorkloadType = "Worker"
	SingletonWorker WorkloadType = "SingletonWorker"
	Task            WorkloadType = "Task"
	SingletonTask   WorkloadType = "SingletonTaskTask"
)

type Application struct {
	types.Namespaced
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

type ApplicationSpec struct {
	Components []Component `json:"components"`
}

type WhiteList struct {
	Users []string `json:"users,omitempty"`
}

type AppIngress struct {
	Host       string `json:"host"`
	Path       string `json:"path,omitempty"`
	ServerPort int32  `json:"serverPort"`
}

type VolumeMounter struct {
	VolumeName   string `json:"volumeName"`
	StorageClass string `json:"storageClass"`
}

type ManualScaler struct {
	Replicas int32 `json:"replicas"`
}

type ComponentTraitsForOpt struct {
	ManualScaler  ManualScaler  `json:"manualScaler,omitempty"`
	VolumeMounter VolumeMounter `json:"volumeMounter,omitempty"`
	Ingress       AppIngress    `json:"ingress"`
	WhiteList     WhiteList     `json:"whiteList,omitempty"`
	Eject         []string      `json:"eject,omitempty"`
	RateLimit     RateLimit     `json:"rateLimit,omitempty"`
}

type RateLimit struct {
	TimeDuration  string     `json:"timeDuration"`
	RequestAmount int32      `json:"requestAmount"`
	Overrides     []Override `json:"overrides,omitempty"`
}

type Override struct {
	RequestAmount int32  `json:"requestAmount"`
	User          string `json:"user"`
}

//负载均衡类型 rr;leastConn;random
//consistentType sourceIP
type IngressLB struct {
	LBType         string `json:"lbType,omitempty"`
	ConsistentType string `json:"consistentType,omitempty"`
}

type ImagePullConfig struct {
	Registry string `json:"registry,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type ComponentTraitsForDev struct {
	ImagePullConfig ImagePullConfig `json:"imagePullConfig,omitempty"`
	StaticIP        bool            `json:"staticIP,omitempty"`
	IngressLB       IngressLB       `json:"ingressLB,omitempty"`
}

type Disk struct {
	Required  string `json:"required"`
	Ephemeral bool   `json:"ephemeral"`
}

type Volume struct {
	Name          string `json:"name"`
	MountPath     string `json:"mountPath"`
	AccessMode    string `json:"accessMode,omitempty"`
	SharingPolicy string `json:"sharingPolicy,omitempty"`
	Disk          Disk   `json:"disk"`
}

type CResource struct {
	Cpu     string   `json:"cpu,omitempty"`
	Memory  string   `json:"memory,omitempty"`
	Gpu     int      `json:"gpu,omitempty"`
	Volumes []Volume `json:"volumes,omitempty"`
}

type EnvVar struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	FromParam string `json:"fromParam"`
}

type AppPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
}

type ComponentContainer struct {
	Name string `json:"name"`

	Image string `json:"image,omitempty"`

	Command []string `json:"command,omitempty"`

	Args []string `json:"args,omitempty"`

	Ports []AppPort `json:"ports,omitempty"`

	Env []EnvVar `json:"env,omitempty"`

	Resources CResource `json:"resources,omitempty"`

	LivenessProbe HealthProbe `json:"livenessProbe,omitempty"`

	ReadinessProbe HealthProbe `json:"readinessProbe,omitempty"`

	ImagePullPolicy PullPolicy `json:"imagePullPolicy,omitempty"`

	Config          []ConfigFile     `json:"config,omitempty"`
	ImagePullSecret string           `json:"imagePullSecret,omitempty"`
	SecurityContext *SecurityContext `json:"securityContext,omitempty"`
}

type WorkloadType string

type Component struct {
	Name       string      `json:"name"`
	Version    string      `json:"version,omitempty"`
	Parameters []Parameter `json:"parameters,omitempty"`

	WorkloadType WorkloadType `json:"workloadType"`

	OsType string `json:"osType,omitempty"`

	Arch string `json:"arch,omitempty"`

	Containers []ComponentContainer `json:"containers,omitempty"`

	WorkloadSettings []WorkloadSetting `json:"workloadSetings,omitempty"`

	DevTraits ComponentTraitsForDev `json:"devTraits,omitempty"`
	OptTraits ComponentTraitsForOpt `json:"optTraits,omitempty"`
}

//int,float,string,bool,json
type Parameter struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
	Required    bool   `json:"required,omitempty"`
	Default     string `json:"default,omitempty"`
}

type SecurityContext struct{}

type ConfigFile struct {
	Path      string `json:"path"`
	FileName  string `json:"fileName"`
	Value     string `json:"value"`
	FromParam string `json:"fromParam,omitempty"`
}

type WorkloadSetting struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Value     string `json:"value"`
	FromParam string `json:"fromParam"`
}

type HealthProbe struct {
	Exec    *ExecAction    `json:"exec,omitempty" protobuf:"bytes,1,opt,name=exec"`
	HTTPGet *HTTPGetAction `json:"httpGet,omitempty" protobuf:"bytes,2,opt,name=httpGet"`
	// TCPSocket specifies an action involving a TCP port.
	// TCP hooks not yet supported
	// TODO: implement a realistic TCP lifecycle hook
	// +optional
	TCPSocket           *TCPSocketAction `json:"tcpSocket,omitempty" protobuf:"bytes,3,opt,name=tcpSocket"`
	InitialDelaySeconds int32            `json:"initialDelaySeconds,omitempty" protobuf:"varint,2,opt,name=initialDelaySeconds"`

	TimeoutSeconds int32 `json:"timeoutSeconds,omitempty" protobuf:"varint,3,opt,name=timeoutSeconds"`

	PeriodSeconds int32 `json:"periodSeconds,omitempty" protobuf:"varint,4,opt,name=periodSeconds"`

	SuccessThreshold int32 `json:"successThreshold,omitempty" protobuf:"varint,5,opt,name=successThreshold"`

	FailureThreshold int32 `json:"failureThreshold,omitempty" protobuf:"varint,6,opt,name=failureThreshold"`
}

type TCPSocketAction struct {
	// Number or name of the port to access on the container.
	// Number must be in the range 1 to 65535.
	// Name must be an IANA_SVC_NAME.
	Port int `json:"port" protobuf:"bytes,1,opt,name=port"`
}

type HTTPGetAction struct {
	// Path to access on the HTTP server.
	// +optional
	Path string `json:"path,omitempty" protobuf:"bytes,1,opt,name=path"`
	// Name or number of the port to access on the container.
	// Number must be in the range 1 to 65535.
	// Name must be an IANA_SVC_NAME.
	Port int `json:"port" protobuf:"bytes,2,opt,name=port"`
	// Host name to connect to, defaults to the pod IP. You probably want to set
	// "Host" in httpHeaders instead.
	// +optional

	HTTPHeaders []HTTPHeader `json:"httpHeaders,omitempty" protobuf:"bytes,5,rep,name=httpHeaders"`
}

type HTTPHeader struct {
	// The header field name
	Name string `json:"name" protobuf:"bytes,1,opt,name=name"`
	// The header field value
	Value string `json:"value" protobuf:"bytes,2,opt,name=value"`
}

type ExecAction struct {
	Command []string `json:"command,omitempty" protobuf:"bytes,1,rep,name=command"`
}

type PullPolicy string

type ApplicationStatus struct {
	ComponentResource map[string]ComponentResources `json:"componentResource,omitempty"`
}

type ComponentResources struct {
	ComponentId        string   `json:"componentId,omitempty"`
	Workload           string   `json:"workload,omitempty"`
	Service            string   `json:"service,omitempty"`
	ConfigMaps         []string `json:"configMaps,omitempty"`
	ImagePullSecret    string   `json:"imagePullSecret,omitempty"`
	Gateway            string   `json:"gateway,omitempty"`
	Policy             string   `json:"policy,omitempty"`
	ClusterRbacConfig  string   `json:"clusterRbacConfig,omitempty"`
	VirtualService     string   `json:"virtualService,omitempty"`
	ServiceRole        string   `json:"serviceRole,omitempty"`
	ServiceRoleBinding string   `json:"serviceRoleBinding,omitempty"`
	DestinationRule    string   `json:"DestinationRule,omitempty"`
}

func (app *Application) Validation() error {
	if app.Name == "" {
		return fmt.Errorf("Please input application name.")
	}
	matched, err := regexp.MatchString(`^[a-z]([-a-z0-9]*[a-z0-9])?`, app.Name)
	if err != nil {
		return fmt.Errorf("Regexp application.name failed, ErrorInfo is %s", err)
	}
	if !matched {
		return fmt.Errorf("Application name %s is invalid a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?'", app.Name)
	}
	if _, ok := app.Labels["projectId"]; !ok {
		return fmt.Errorf("projectId not in Application Labels,Please add it.")
	}
	if _, ok := app.Labels["applicationTemplateId"]; !ok {
		return fmt.Errorf("applicationTemplateId not in Application Labels,Please add it.")
	}
	var componentname string
	var componentversion map[string]int = make(map[string]int)
	for _, com := range app.Spec.Components {
		if com.Name == "" {
			return fmt.Errorf("Component.name can't be empty")
		}
		matched, err := regexp.MatchString(`^[a-z]([-a-z0-9]*[a-z0-9])?`, com.Name)
		if err != nil {
			return fmt.Errorf("Regexp application.name failed, ErrorInfo is %s", err)
		}
		if !matched {
			return fmt.Errorf("Component name %s is invalid a DNS-1035 label must consist of lower case alphanumeric characters or '-', start with an alphabetic character, and end with an alphanumeric character (e.g. 'my-name',  or 'abc-123', regex used for validation is '[a-z]([-a-z0-9]*[a-z0-9])?'", com.Name)
		}
		if componentname == "" {
			componentname = com.Name
		} else {
			if componentname != com.Name {
				return fmt.Errorf("If the application has multiple components their names must be the same")
			}
		}
		if !(com.WorkloadType == "Server") {
			return fmt.Errorf("WorkloadType Need be Server.")
		}
		if com.Version == "" {
			return fmt.Errorf("Please specify the version.")
		}
		if _, ok := componentversion[com.Version]; ok {
			fmt.Errorf("The same component must have different versions")
		} else {
			componentversion[com.Version] = 1
		}
		for _, con := range com.Containers {
			if con.Name == "" {
				return fmt.Errorf("Please specify the %s's container name.", com.Name)
			}
			if len(con.Config) != 0 {
				for _, v := range con.Config {
					if v.Path == "" || v.Value == "" {
						return fmt.Errorf("application.components.container.config's path and value can't be empty at the same time")
					}
					matched, err := regexp.MatchString(`^\/(\w+\/?)+$`, v.Path)
					if err != nil {
						return fmt.Errorf("Regexp application.components.containers.config.path's failed, ErrorInfo is %s", err)
					}
					if !matched {
						return fmt.Errorf("application.components.containers.config.path's syntax is err")
					}
				}
			}
			if con.Image == "" {
				return fmt.Errorf("Image can't be empty")
			}
			if len(con.Ports) != 0 {
				for _, port := range con.Ports {
					if port.ContainerPort <= 0 {
						return fmt.Errorf("Please input correct ContainerPort.")
					}
				}
			}
			if !(reflect.DeepEqual(con.Resources, CResource{})) {
				matched1, err1 := regexp.MatchString(`^[0-9]\d*[MG]i$`, con.Resources.Memory)
				if err1 != nil {
					return fmt.Errorf("Regexp application.components.containers.resources.memory failed, ErrorInfo is %s", err1)
				}
				if !matched1 {
					return fmt.Errorf("application.components.containers.resources.memory's syntax is err")
				}
				matched2, err2 := regexp.MatchString(`^[0-9]\d*m$`, con.Resources.Cpu)
				if err2 != nil {
					return fmt.Errorf("Regexp application.components.containers.resources.cpu failed, ErrorInfo is %s", err2)
				}
				if !matched2 {
					return fmt.Errorf("application.components.containers.resources.cpu's syntax is err")
				}
				/*if con.Resources.Gpu <= 0 {
					return fmt.Errorf("Regexp application.components.containers.resources.gpu must be greater than 0")
				}*/
				if len(con.Resources.Volumes) != 0 {
					for _, v := range con.Resources.Volumes {
						if v.Name == "" || v.MountPath == "" {
							return fmt.Errorf("application.components.container.resource.volumes's name and mountpath can't be empty at the same time")
						}
					}
				}
			}
		}
		if !(reflect.DeepEqual(com.DevTraits, ComponentTraitsForDev{})) {
			if !(com.DevTraits.ImagePullConfig == ImagePullConfig{}) {
				if com.DevTraits.ImagePullConfig.Registry == "" || com.DevTraits.ImagePullConfig.Password == "" || com.DevTraits.ImagePullConfig.Username == "" {
					return fmt.Errorf("application.components.devtraits.imagepullconfig's username、password and registry can't be empty at the same time")
				}
			}
			if !(com.DevTraits.IngressLB == IngressLB{}) {
				if com.DevTraits.IngressLB.ConsistentType != "" && com.DevTraits.IngressLB.LBType != "" {
					fmt.Errorf("You can only choose one of these two strategies")
				}
				if com.DevTraits.IngressLB.ConsistentType != "" && com.DevTraits.IngressLB.ConsistentType != "sourceIP" {
					fmt.Errorf("application.components.devtraits.ingresslb.consistentType only support sourceIP")
				}
				if com.DevTraits.IngressLB.LBType != "" && !(com.DevTraits.IngressLB.LBType == "rr" || com.DevTraits.IngressLB.LBType == "leastConn" || com.DevTraits.IngressLB.LBType == "random") {
					fmt.Errorf("application.components.devtraits.ingresslb.LBType only support rr leastConn random")
				}
			}
		}
		if (reflect.DeepEqual(com.OptTraits, ComponentTraitsForOpt{})) {
			return fmt.Errorf("application.components.opttraits.ingress must be configured")
		} else {
			if (com.OptTraits.Ingress == AppIngress{}) {
				return fmt.Errorf("application.components.opttraits.ingress must be configured")
			} else {
				if com.OptTraits.Ingress.Host == "" || com.OptTraits.Ingress.Path == "" || com.OptTraits.Ingress.ServerPort <= 0 {
					return fmt.Errorf("application.components.opttraits.ingress's host、path and serverPort can't be empty at the same time")
				} else {
					//matched, err := regexp.MatchString(`^\/(\w+\/?)+$`, com.OptTraits.Ingress.Path)
					if com.OptTraits.Ingress.Path != "/" {
						//return fmt.Errorf("Regexp application.components.opttraits.ingress's failed, ErrorInfo is %s", err)
						return fmt.Errorf("application.components.opttraits.ingress's path must be /")
					}
				}
			}
			if (com.OptTraits.ManualScaler == ManualScaler{}) {
				return fmt.Errorf("component.opttraits.manualscaler field cannot be empty")
			} else {
				if com.OptTraits.ManualScaler.Replicas <= 0 {
					return fmt.Errorf("component.opttraits.manualscaler.replicas must be greater than 0")
				}
			}
			if !(reflect.DeepEqual(com.OptTraits.RateLimit, RateLimit{})) {
				if com.OptTraits.RateLimit.TimeDuration == "" || com.OptTraits.RateLimit.RequestAmount <= 0 {
					return fmt.Errorf("application.components.opttraits.ratelimit.timeduration and requestamount can't be empty at the same time")
				}
				if len(com.OptTraits.RateLimit.Overrides) != 0 {
					for _, i := range com.OptTraits.RateLimit.Overrides {
						if i.RequestAmount <= 0 || i.User == "" {
							return fmt.Errorf("application.components.opttraits.ratelimit.overrides.user and requestamount can't be empty at the same time")
						}
					}
				}
			}
			if len(com.OptTraits.WhiteList.Users) != 0 {
				for _, i := range com.OptTraits.WhiteList.Users {
					matched, err := regexp.MatchString(`^.*@.*$`, i)
					if err != nil {
						return fmt.Errorf("Regexp application.components.opttraits.whitelist.users %s failed, ErrorInfo is %s", i, err)
					}
					if matched {
						continue
					} else {
						return fmt.Errorf("Regexp application.components.opttraits.whitelist.users %s failed", i)
					}
				}
			}
		}
	}
	return nil
}
