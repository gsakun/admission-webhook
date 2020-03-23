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
					if env.Value == "" && !(env.FromParam == "spec.nodeName" || env.FromParam == "metadata.name" || env.FromParam == "metadata.namespace" || env.FromParam == "status.podIP") {
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
						if !v.Disk.Ephemeral && v.Disk.Required == "" {
							return fmt.Errorf("if disk.ephemeral false,disk.required can't be empty")
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
			//if !(reflect.DeepEqual(com.OptTraits.CircuitBreaking,CircuitBreaking{})){
			//	if com.OptTraits.CircuitBreaking.ConnectionPool
			//}
		}
	}
	return nil
}
