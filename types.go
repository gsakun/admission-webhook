package main

import (
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
	ManualScaler    ManualScaler    `json:"manualScaler,omitempty"`
	VolumeMounter   VolumeMounter   `json:"volumeMounter,omitempty"`
	Ingress         AppIngress      `json:"ingress"`
	WhiteList       WhiteList       `json:"whiteList,omitempty"`
	Eject           []string        `json:"eject,omitempty"`
	Fusing          Fusing          `json:"fusing,omitempty"` //zk
	RateLimit       RateLimit       `json:"rateLimit,omitempty"`
	CircuitBreaking CircuitBreaking `json:"circuitbreaking,omitempty"` //zk
}

//zk
type CircuitBreaking struct {
	LoadBalancer      LoadBalancerSettings   `json:"loadBalancer,omitempty"`
	ConnectionPool    ConnectionPoolSettings `json:"connectionPool,omitempty"`
	OutlierDetection  OutlierDetection       `json:"outlierDetection,omitempty"`
	PortLevelSettings []PortTrafficPolicy    `json:"portLevelSettings,omitempty"`
}

type ConnectionPoolSettings struct {

	// Settings common to both HTTP and TCP upstream connections.
	TCP TCPSettings `json:"tcp,omitempty"`

	// HTTP connection pool settings.
	HTTP HTTPSettings `json:"http,omitempty"`
}

type PortSelector struct {
	// Choose one of the fields below.

	// Valid port number
	Number uint32 `json:"number,omitempty"`

	// Valid port name
	Name string `json:"name,omitempty"`
}

type OutlierDetection struct {
	// Number of errors before a host is ejected from the connection
	// pool. Defaults to 5. When the upstream host is accessed over HTTP, a
	// 5xx return code qualifies as an error. When the upstream host is
	// accessed over an opaque TCP connection, connect timeouts and
	// connection error/failure events qualify as an error.
	ConsecutiveErrors int32 `json:"consecutiveErrors,omitempty"`

	// Time interval between ejection sweep analysis. format:
	// 1h/1m/1s/1ms. MUST BE >=1ms. Default is 10s.
	Interval string `json:"interval,omitempty"`

	// Minimum ejection duration. A host will remain ejected for a period
	// equal to the product of minimum ejection duration and the number of
	// times the host has been ejected. This technique allows the system to
	// automatically increase the ejection period for unhealthy upstream
	// servers. format: 1h/1m/1s/1ms. MUST BE >=1ms. Default is 30s.
	BaseEjectionTime string `json:"baseEjectionTime,omitempty"`

	// Maximum % of hosts in the load balancing pool for the upstream
	// service that can be ejected. Defaults to 10%.
	MaxEjectionPercent int32 `json:"maxEjectionPercent,omitempty"`
}

// Settings common to both HTTP and TCP upstream connections.
type TCPSettings struct {
	// Maximum number of HTTP1 /TCP connections to a destination host.
	MaxConnections int32 `json:"maxConnections,omitempty"`

	// TCP connection timeout.
	ConnectTimeout string `json:"connectTimeout,omitempty"`
}

// Settings applicable to HTTP1.1/HTTP2/GRPC connections.
type HTTPSettings struct {
	// Maximum number of pending HTTP requests to a destination. Default 1024.
	HTTP1MaxPendingRequests int32 `json:"http1MaxPendingRequests,omitempty"`

	// Maximum number of requests to a backend. Default 1024.
	HTTP2MaxRequests int32 `json:"http2MaxRequests,omitempty"`

	// Maximum number of requests per connection to a backend. Setting this
	// parameter to 1 disables keep alive.
	MaxRequestsPerConnection int32 `json:"maxRequestsPerConnection,omitempty"`

	// Maximum number of retries that can be outstanding to all hosts in a
	// cluster at a given time. Defaults to 3.
	MaxRetries int32 `json:"maxRetries,omitempty"`
}

type SimpleLB string

const (
	// Round Robin policy. Default
	SimpleLBRoundRobin SimpleLB = "ROUND_ROBIN"

	// The least request load balancer uses an O(1) algorithm which selects
	// two random healthy hosts and picks the host which has fewer active
	// requests.
	SimpleLBLeastConn SimpleLB = "LEAST_CONN"

	// The random load balancer selects a random healthy host. The random
	// load balancer generally performs better than round robin if no health
	// checking policy is configured.
	SimpleLBRandom SimpleLB = "RANDOM"

	// This option will forward the connection to the original IP address
	// requested by the caller without doing any form of load
	// balancing. This option must be used with care. It is meant for
	// advanced use cases. Refer to Original Destination load balancer in
	// Envoy for further details.
	SimpleLBPassthrough SimpleLB = "PASSTHROUGH"
)

//zk
type Fusing struct {
	PodList []string `json:"podlist,omitempty"`
	Action  string   `json:"action,omitempty"`
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

// Traffic policies that apply to specific ports of the service
type PortTrafficPolicy struct {
	Port             PortSelector           `json:"port"`
	LoadBalancer     LoadBalancerSettings   `json:"loadBalancer,omitempty"`
	ConnectionPool   ConnectionPoolSettings `json:"connectionPool,omitempty"`
	OutlierDetection OutlierDetection       `json:"outlierDetection,omitempty"`
}
type LoadBalancerSettings struct {
	Simple         SimpleLB         `json:"simple,omitempty"`
	ConsistentHash ConsistentHashLB `json:"consistentHash,omitempty"`
}

type ConsistentHashLB struct {
	HTTPHeaderName  string `json:"httpHeaderName,omitempty"`
	UseSourceIP     bool   `json:"useSourceIp,omitempty"`
	MinimumRingSize uint64 `json:"minimumRingSize,omitempty"`
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
	ImagePullConfig ImagePullConfig `json:"imagePullConfig"`
	StaticIP        bool            `json:"staticIP,omitempty"`
	IngressLB       IngressLB       `json:"ingressLB,omitempty"`
}

type Disk struct {
	Required  string `json:"required,omitempty"`
	Ephemeral bool   `json:"ephemeral"`
}

type CVolume struct {
	Name          string `json:"name"`
	MountPath     string `json:"mountPath"`
	AccessMode    string `json:"accessMode,omitempty"`
	SharingPolicy string `json:"sharingPolicy,omitempty"`
	Disk          Disk   `json:"disk"`
}

type CResource struct {
	Cpu     string    `json:"cpu,omitempty"`
	Memory  string    `json:"memory,omitempty"`
	Gpu     int       `json:"gpu,omitempty"`
	Volumes []CVolume `json:"volumes,omitempty"`
}

type CEnvVar struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	FromParam string `json:"fromParam,omitempty"`
}

type AppPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol,omitempty"`
}

type ComponentContainer struct {
	Name string `json:"name"`

	Image string `json:"image"`

	Command []string `json:"command,omitempty"`

	Args []string `json:"args,omitempty"`

	Ports []AppPort `json:"ports,omitempty"`

	Env []CEnvVar `json:"env,omitempty"`

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
	Version    string      `json:"version"`
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

type Handler struct {
	Exec      ExecAction      `json:"exec,omitempty" protobuf:"bytes,1,opt,name=exec"`
	HTTPGet   HTTPGetAction   `json:"httpGet,omitempty" protobuf:"bytes,2,opt,name=httpGet"`
	TCPSocket TCPSocketAction `json:"tcpSocket,omitempty" protobuf:"bytes,3,opt,name=tcpSocket"`
}

type HealthProbe struct {
	Handler             `json:haneler",inline" protobuf:"bytes,1,opt,name=handler"`
	InitialDelaySeconds int32 `json:"initialDelaySeconds,omitempty" protobuf:"varint,2,opt,name=initialDelaySeconds"`

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
