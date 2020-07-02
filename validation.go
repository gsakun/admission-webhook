package main

import (
	"fmt"
	"reflect"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func (app *Application) Validation() error {
	log.Infoln("START Validation")
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
	/*if _, ok := app.Labels["projectId"]; !ok {
		return fmt.Errorf("projectId not in Application Labels,Please add it.")
	}*/
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
		if com.WorkloadType != "Server" {
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
			if len(con.Env) != 0 {
				for _, env := range con.Env {
					if env.Name == "" {
						return fmt.Errorf("Env name can't be empty")
					}
					if env.Value == "" && env.FromParam == "" {
						return fmt.Errorf("If Env.Name not be empty,Env's Value or FromParam can't be empty")
					}
					if env.FromParam == "spec.nodeName" || env.FromParam == "metadata.name" || env.FromParam == "metadata.namespace" || env.FromParam == "status.podIP" {
						return fmt.Errorf("Only these fields are allowed to be populated fromparam(spec.nodeName,metadata.name,metadata.namespace,status.podIP)")
					}

				}
			}
			if len(con.Config) != 0 {
				for _, v := range con.Config {
					if v.Path == "" || v.Value == "" || v.FileName == "" {
						return fmt.Errorf("application.components.container.config's path 、value、filename can't be empty at the same time")
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
					return fmt.Errorf("application.components.containers.resources.memory's unit is err")
				}
				matched2, err2 := regexp.MatchString(`^[0-9]\d*m$`, con.Resources.Cpu)
				if err2 != nil {
					return fmt.Errorf("Regexp application.components.containers.resources.cpu failed, ErrorInfo is %s", err2)
				}
				if !matched2 {
					return fmt.Errorf("application.components.containers.resources.cpu's unit is err")
				}
				/*if con.Resources.Gpu <= 0 {
					return fmt.Errorf("Regexp application.components.containers.resources.gpu must be greater than 0")
				}*/
				if len(con.Resources.Volumes) != 0 {
					for _, v := range con.Resources.Volumes {
						if v.Name == "" || v.MountPath == "" {
							return fmt.Errorf("application.components.container.resource.volumes's name and mountpath can't be empty at the same time")
						}
						if !v.Disk.Ephemeral && v.Disk.Required == "" {
							return fmt.Errorf("if disk.ephemeral false,disk.required can't be empty")
						}
					}
				}
			}
			if !reflect.DeepEqual(con.LivenessProbe, HealthProbe{}) {
				if !reflect.DeepEqual(con.LivenessProbe.Exec, ExecAction{}) {
					if len(con.LivenessProbe.Exec.Command) == 0 {
						return fmt.Errorf("If Exec option has configured,command can't be null")
					}
					if !reflect.DeepEqual(con.LivenessProbe.HTTPGet, HTTPGetAction{}) {
						return fmt.Errorf("Exec has been configured in livenessProbe,The other options cannot be selected")
					}
					if !reflect.DeepEqual(con.LivenessProbe.TCPSocket, TCPSocketAction{}) {
						return fmt.Errorf("tcpsocket has been configured in livenessProbe,The other options cannot be selected")
					}
				}
				if !reflect.DeepEqual(con.LivenessProbe.HTTPGet, HTTPGetAction{}) {
					if con.LivenessProbe.HTTPGet.Port <= 0 || con.LivenessProbe.HTTPGet.Path == "" {
						return fmt.Errorf("Please check application.components.containers.livenessProbe's httpget field")
					}
					if !reflect.DeepEqual(con.LivenessProbe.TCPSocket, TCPSocketAction{}) {
						return fmt.Errorf("tcpsocket has been configured in livenessProbe,The other options cannot be selected")
					}
				}
				if !reflect.DeepEqual(con.LivenessProbe.TCPSocket, TCPSocketAction{}) {
					if con.LivenessProbe.TCPSocket.Port <= 0 {
						return fmt.Errorf("Please check application.components.containers.livenessProbe's Tcpsocket field")
					}
				} else {
					fmt.Errorf("LivenessProbe's exec httpget tcpsocket are not configured,please unset livenessprobe field.")
				}
				if con.LivenessProbe.InitialDelaySeconds <= 0 || con.LivenessProbe.PeriodSeconds <= 0 || con.LivenessProbe.SuccessThreshold <= 0 || con.LivenessProbe.FailureThreshold <= 0 {
					return fmt.Errorf("LivenessProbe's InitialDelaySeconds PeriodSeconds SuccessThreshold FailureThreshold can't <= 0")
				}
			}
			if !reflect.DeepEqual(con.ReadinessProbe, HealthProbe{}) {
				if !reflect.DeepEqual(con.ReadinessProbe.Exec, ExecAction{}) {
					if len(con.ReadinessProbe.Exec.Command) == 0 {
						return fmt.Errorf("If Exec option has configured,command can't be null")
					}
					if !reflect.DeepEqual(con.ReadinessProbe.HTTPGet, HTTPGetAction{}) {
						return fmt.Errorf("Exec has been configured in ReadinessProbe,The other options cannot be selected")
					}
					if !reflect.DeepEqual(con.ReadinessProbe.TCPSocket, TCPSocketAction{}) {
						return fmt.Errorf("tcpsocket has been configured in ReadinessProbe,The other options cannot be selected")
					}
				}
				if !reflect.DeepEqual(con.ReadinessProbe.HTTPGet, HTTPGetAction{}) {
					if con.ReadinessProbe.HTTPGet.Port <= 0 || con.ReadinessProbe.HTTPGet.Path == "" {
						return fmt.Errorf("Please check application.components.containers.ReadinessProbe's httpget field")
					}
					if !reflect.DeepEqual(con.ReadinessProbe.TCPSocket, TCPSocketAction{}) {
						return fmt.Errorf("tcpsocket has been configured in ReadinessProbe,The other options cannot be selected")
					}
				}
				if !reflect.DeepEqual(con.ReadinessProbe.TCPSocket, TCPSocketAction{}) {
					if con.ReadinessProbe.TCPSocket.Port <= 0 {
						return fmt.Errorf("Please check application.components.containers.ReadinessProbe's Tcpsocket field")
					}
				} else {
					fmt.Errorf("ReadinessProbe's exec httpget tcpsocket are not configured,please unset ReadinessProbe field.")
				}
				if con.ReadinessProbe.InitialDelaySeconds <= 0 || con.ReadinessProbe.PeriodSeconds <= 0 || con.ReadinessProbe.SuccessThreshold <= 0 || con.ReadinessProbe.FailureThreshold <= 0 {
					return fmt.Errorf("ReadinessProbe's InitialDelaySeconds PeriodSeconds SuccessThreshold FailureThreshold can't <= 0")
				}
			}
		}
		if com.ComponentTraits.Replicas <= 0 {
			return fmt.Errorf("app.spec.component.componenttraits.replicas at least 1")
		}
		if com.ComponentTraits.CustomMetric.Enable {
			if com.ComponentTraits.CustomMetric.Uri == "" {
				return fmt.Errorf("If com.ComponentTraits.CustomMetric.Enable is true,com.ComponentTraits.CustomMetric.Enable.uri can't be empty")
			}
		}
		if !reflect.DeepEqual(com.ComponentTraits.Autoscaling, Autoscaling{}) {
			if com.ComponentTraits.Autoscaling.Metric == "" || com.ComponentTraits.Autoscaling.Threshold <= 0 || com.ComponentTraits.Autoscaling.MinReplicas <= 0 || com.ComponentTraits.Autoscaling.MaxReplicas <= com.ComponentTraits.Autoscaling.MinReplicas {
				return fmt.Errorf("Please check autoscaling configuration")
			}
		}
	}
	if (reflect.DeepEqual(app.Spec.OptTraits, ComponentTraitsForOpt{})) {
		return fmt.Errorf("application.opttraits.ingress must be configured")
	} else {
		if (app.Spec.OptTraits.Ingress == AppIngress{}) {
			return fmt.Errorf("application.opttraits.ingress must be configured")
		} else {
			if app.Spec.OptTraits.Ingress.Host == "" || app.Spec.OptTraits.Ingress.Path == "" || app.Spec.OptTraits.Ingress.ServerPort <= 0 {
				return fmt.Errorf("application.opttraits.ingress's host、path and serverPort can't be empty at the same time")
			} else {
				//matched, err := regexp.MatchString(`^\/(\w+\/?)+$`, app.Spec.OptTraits.Ingress.Path)
				if app.Spec.OptTraits.Ingress.Path != "/" {
					//return fmt.Errorf("Regexp application.opttraits.ingress's failed, ErrorInfo is %s", err)
					return fmt.Errorf("application.opttraits.ingress's path must be /")
				}
			}
		}
		if !(reflect.DeepEqual(app.Spec.OptTraits.RateLimit, RateLimit{})) {
			if app.Spec.OptTraits.RateLimit.TimeDuration == "" || app.Spec.OptTraits.RateLimit.RequestAmount <= 0 {
				return fmt.Errorf("application.opttraits.ratelimit.timeduration and requestamount can't be empty at the same time")
			}
			matched, err := checkinterval(app.Spec.OptTraits.RateLimit.TimeDuration)
			if !matched {
				if err != nil {
					return fmt.Errorf("application.opttraits.ratelimit.timeduration regex failed errinfo is %v", err)
				}
				return fmt.Errorf("application.opttraits.ratelimit.timeduration must end with s or m")
			}
			if len(app.Spec.OptTraits.RateLimit.Overrides) != 0 {
				for _, i := range app.Spec.OptTraits.RateLimit.Overrides {
					if i.RequestAmount <= 0 || i.User == "" {
						return fmt.Errorf("application.opttraits.ratelimit.overrides.user and requestamount can't be empty at the same time")
					}
				}
			}
		}
		if len(app.Spec.OptTraits.WhiteList.Users) != 0 {
			for _, i := range app.Spec.OptTraits.WhiteList.Users {
				matched, err := regexp.MatchString(`^.*@.*$`, i)
				if err != nil {
					return fmt.Errorf("Regexp application.opttraits.whitelist.users %s failed, ErrorInfo is %s", i, err)
				}
				if matched {
					continue
				} else {
					return fmt.Errorf("Regexp application.opttraits.whitelist.users %s failed", i)
				}
			}
		}
		if !reflect.DeepEqual(app.Spec.OptTraits.HTTPRetry, HTTPRetry{}) {
			if app.Spec.OptTraits.HTTPRetry.Attempts <= 0 || app.Spec.OptTraits.HTTPRetry.PerTryTimeout == "" {
				return fmt.Errorf("Please check httpretry configuration")
			}
			matched, err := checkinterval(app.Spec.OptTraits.HTTPRetry.PerTryTimeout)
			if !matched {
				if err != nil {
					return fmt.Errorf("application.opttraits.httpretry.pertrytimeout regex failed errinfo is %v", err)
				}
				return fmt.Errorf("application.opttraits.httpretry.pertrytimeout must end with s or m")
			}
		}
		if !(reflect.DeepEqual(app.Spec.OptTraits.CircuitBreaking, CircuitBreaking{})) {
			if !reflect.DeepEqual(app.Spec.OptTraits.CircuitBreaking.ConnectionPool, ConnectionPoolSettings{}) {
				if !reflect.DeepEqual(app.Spec.OptTraits.CircuitBreaking.ConnectionPool.TCP, TCPSettings{}) {
					if app.Spec.OptTraits.CircuitBreaking.ConnectionPool.TCP.MaxConnections <= 0 {
						return fmt.Errorf("app.Spec.OptTraits.CircuitBreaking.ConnectionPool.TCP.MaxConnections must >=0")
					}
					match, err := checkinterval(app.Spec.OptTraits.CircuitBreaking.ConnectionPool.TCP.ConnectTimeout)
					if !match {
						if err != nil {
							return fmt.Errorf("app.Spec.OptTraits.CircuitBreaking.ConnectionPool.TCP.ConnectTimeout regex failed errinfo is %v", err)
						}
						return fmt.Errorf("app.Spec.OptTraits.CircuitBreaking.ConnectionPool.TCP.ConnectTimeout must end with s or m")
					}
					if !reflect.DeepEqual(app.Spec.OptTraits.CircuitBreaking.OutlierDetection, OutlierDetection{}) {
						if app.Spec.OptTraits.CircuitBreaking.OutlierDetection.BaseEjectionTime == "" || app.Spec.OptTraits.CircuitBreaking.OutlierDetection.Interval == "" || app.Spec.OptTraits.CircuitBreaking.OutlierDetection.ConsecutiveErrors <= 0 || app.Spec.OptTraits.CircuitBreaking.OutlierDetection.MaxEjectionPercent <= 0 {
							return fmt.Errorf("Please check httpretry configuration")
						}
					}
					match, err = checkinterval(app.Spec.OptTraits.CircuitBreaking.OutlierDetection.BaseEjectionTime)
					if !match {
						if err != nil {
							return fmt.Errorf("app.Spec.OptTraits.CircuitBreaking.OutlierDetection.BaseEjectionTime regex failed errinfo is %v", err)
						}
						return fmt.Errorf("app.Spec.OptTraits.CircuitBreaking.OutlierDetection.BaseEjectionTime must end with s or m")
					}
					match, err = checkinterval(app.Spec.OptTraits.CircuitBreaking.OutlierDetection.Interval)
					if !match {
						if err != nil {
							return fmt.Errorf("app.Spec.OptTraits.CircuitBreaking.OutlierDetection.Interval regex failed errinfo is %v", err)
						}
						return fmt.Errorf("app.Spec.OptTraits.CircuitBreaking.OutlierDetection.Interval must end with s or m")
					}
				}

			}

		}
	}
	return nil
}

func checkinterval(interval string) (match bool, err error) {
	matched, err := regexp.MatchString(`^[0-9]\d*[s|m|d]$`, interval)
	if err != nil {
		return false, fmt.Errorf("Regexp failed, ErrorInfo is %s", err)
	}
	if !matched {
		return false, nil
	}
	return true, nil
}
