package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"k8s.io/api/admission/v1beta1"
	admissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	runtimeScheme = runtime.NewScheme()
	codecs        = serializer.NewCodecFactory(runtimeScheme)
	deserializer  = codecs.UniversalDeserializer()

	// (https://github.com/kubernetes/kubernetes/issues/57982)
	defaulter = runtime.ObjectDefaulter(runtimeScheme)
)

type Config struct {
	Containers []corev1.Container `yaml:"containers"`
}

var (
	ignoredNamespaces = []string{
		metav1.NamespaceSystem,
		metav1.NamespacePublic,
	}
	requiredLabels = []string{
		nameLabel,
		instanceLabel,
		versionLabel,
		componentLabel,
		partOfLabel,
		managedByLabel,
	}
	/*addLabels = map[string]string{
		nameLabel:      NA,
		instanceLabel:  NA,
		versionLabel:   NA,
		componentLabel: NA,
		partOfLabel:    NA,
		managedByLabel: NA,
	}*/
)

const (
	admissionWebhookAnnotationValidateKey = "admission-webhook-example.qikqiak.com/validate"
	admissionWebhookAnnotationMutateKey   = "admission-webhook-example.qikqiak.com/mutate"
	admissionWebhookAnnotationStatusKey   = "admission-webhook-example.qikqiak.com/status"
	admissionWebhookAnnotationPodNoCreate = "admission-webhook-example.qikqiak.com/podnocreate"

	nameLabel      = "app.kubernetes.io/name"
	instanceLabel  = "app.kubernetes.io/instance"
	versionLabel   = "app.kubernetes.io/version"
	componentLabel = "app.kubernetes.io/component"
	partOfLabel    = "app.kubernetes.io/part-of"
	managedByLabel = "app.kubernetes.io/managed-by"

	NA = "not_available"
)

type WebhookServer struct {
	sidecarConfig *Config
	server        *http.Server
}

// Webhook Server parameters
type WhSvrParameters struct {
	port           int    // webhook server port
	certFile       string // path to the x509 certificate for https
	keyFile        string // path to the x509 private key matching `CertFile`
	sidecarCfgFile string // path to sidecar injector configuration file
}

type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

func init() {
	_ = corev1.AddToScheme(runtimeScheme)
	_ = admissionregistrationv1beta1.AddToScheme(runtimeScheme)
	// defaulting with webhooks:
	// https://github.com/kubernetes/kubernetes/issues/57982
	_ = corev1.AddToScheme(runtimeScheme)
}

func admissionRequired(ignoredList []string, admissionAnnotationKey string, metadata *metav1.ObjectMeta) bool {
	// skip special kubernetes system namespaces
	for _, namespace := range ignoredList {
		if metadata.Namespace == namespace {
			glog.Infof("Skip validation for %v for it's in special namespace:%v", metadata.Name, metadata.Namespace)
			return false
		}
	}

	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	var required bool
	switch strings.ToLower(annotations[admissionAnnotationKey]) {
	default:
		required = true
	case "n", "no", "false", "off":
		required = false
	}
	return required
}

func mutationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	required := admissionRequired(ignoredList, admissionWebhookAnnotationMutateKey, metadata)
	annotations := metadata.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}
	status := annotations[admissionWebhookAnnotationStatusKey]

	if strings.ToLower(status) == "injected" {
		required = false
	}
	glog.Infof("Mutation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

func validationRequired(ignoredList []string, metadata *metav1.ObjectMeta) bool {
	required := admissionRequired(ignoredList, admissionWebhookAnnotationValidateKey, metadata)
	glog.Infof("Validation policy for %v/%v: required:%v", metadata.Namespace, metadata.Name, required)
	return required
}

func updateAnnotation(target map[string]string, added map[string]string) (patch []patchOperation) {
	for key, value := range added {
		/*		if target == nil || target[key] == "" {
				target[key] = value
			} else {*/
		target[key] = value
		//}
	}
	patch = append(patch, patchOperation{
		Op:    "add",
		Path:  "/metadata/annotations",
		Value: target,
	})
	return patch
}

func updateLabels(target map[string]string, added map[string]string) (patch []patchOperation) {
	for key, value := range added {
		/*if target == nil || target[key] == "" {
			target[key] = value
		} else {*/
		target[key] = value
		//}
	}
	patch = append(patch, patchOperation{
		Op:    "add",
		Path:  "/spec/template/metadata/labels",
		Value: target,
	})
	return patch
}

func createPatch(availableAnnotations map[string]string, annotations map[string]string, availableLabels map[string]string, labels map[string]string) ([]byte, error) {
	var patch []patchOperation

	//patch = append(patch, updateAnnotation(availableAnnotations, annotations)...)
	patch = append(patch, updateLabels(availableLabels, labels)...)
	//patch = append(patch, updateSpecTemplateMetadateLabels(podavailableLabels, labels)...)

	return json.Marshal(patch)
}

// validate deployments and services
func (whsvr *WebhookServer) validate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v",
		req.Kind, req.Namespace, req.Name)
	allowed := true
	var result *metav1.Status
	if req.Kind.Kind == "Application" {
		var application Application
		if err := json.Unmarshal(req.Object.Raw, &application); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		glog.Infoln(application)
		err := application.Validation()
		if err != nil {
			allowed = false
			result = &metav1.Status{
				Reason:  "Some fields fail validation",
				Message: err.Error(),
			}
		}
	}

	if req.Kind.Kind == "Pod" {
		var pod corev1.Pod
		if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		glog.Infoln(pod)
		if pod.Annotations[admissionWebhookAnnotationPodNoCreate] == "yes" {
			allowed = false
			result = &metav1.Status{
				Reason: "This pod's deployment is deleting specfic pod",
			}
		}
	}

	return &v1beta1.AdmissionResponse{
		Allowed: allowed,
		Result:  result,
	}
}

// main mutation process
func (whsvr *WebhookServer) mutate(ar *v1beta1.AdmissionReview) *v1beta1.AdmissionResponse {
	req := ar.Request
	switch req.Kind.Kind {
	/*case "Deployment":
	var deployment appsv1.Deployment
	if err := json.Unmarshal(req.Object.Raw, &deployment); err != nil {
		glog.Errorf("Could not unmarshal raw object: %v", err)
		return &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}
	resourceName, resourceNamespace, objectMeta = deployment.Name, deployment.Namespace, &deployment.ObjectMeta
	availableLabels = deployment.Labels
	availableAnnotations = deployment.Annotations*/
	case "Pod":
		var pod corev1.Pod
		if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		glog.Infof("AdmissionReview for Kind=%v, Namespace=%v Name=%v (%v) UID=%v patchOperation=%v UserInfo=%v",
			req.Kind, req.Namespace, req.Name, pod.Name, req.UID, req.Operation, req.UserInfo)
		if !mutationRequired(ignoredNamespaces, &pod.ObjectMeta) {
			glog.Infof("Skipping mutation for %s/%s due to policy check", pod.Namespace, pod.Name)
			return &v1beta1.AdmissionResponse{
				Allowed: true,
			}
		}
		if len(pod.Annotations) == 0 {
			pod.Annotations = map[string]string{admissionWebhookAnnotationStatusKey: "injected"}
		} else {
			pod.Annotations[admissionWebhookAnnotationStatusKey] = "injected"
		}

		//applyDefaultsWorkaround(whsvr.sidecarConfig.Containers)
		//annotations := map[string]string{admissionWebhookAnnotationStatusKey: "injected"}
		/*patchBytes, err := createPodPatch(&pod, whsvr.sidecarConfig, annotations)
		if err != nil {
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}

		glog.Infof("AdmissionResponse: patch=%v\n", string(patchBytes))
		return &v1beta1.AdmissionResponse{
			Allowed: true,
			Patch:   patchBytes,
			PatchType: func() *v1beta1.PatchType {
				pt := v1beta1.PatchTypeJSONPatch
				return &pt
			}(),
		}*/
		//podavailableLabels = deployment.Spec.Template.Labels
		/*case "Service":
		var service corev1.Service
		if err := json.Unmarshal(req.Object.Raw, &service); err != nil {
			glog.Errorf("Could not unmarshal raw object: %v", err)
			return &v1beta1.AdmissionResponse{
				Result: &metav1.Status{
					Message: err.Error(),
				},
			}
		}
		resourceName, resourceNamespace, objectMeta = service.Name, service.Namespace, &service.ObjectMeta
		availableLabels = service.Labels
		availableAnnotations = service.Annotations*/
	}
	return nil
}

// Serve method for webhook server
func (whsvr *WebhookServer) serve(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}
	if len(body) == 0 {
		glog.Error("empty body")
		http.Error(w, "empty body", http.StatusBadRequest)
		return
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		glog.Errorf("Content-Type=%s, expect application/json", contentType)
		http.Error(w, "invalid Content-Type, expect `application/json`", http.StatusUnsupportedMediaType)
		return
	}

	var admissionResponse *v1beta1.AdmissionResponse
	ar := v1beta1.AdmissionReview{}
	if _, _, err := deserializer.Decode(body, nil, &ar); err != nil {
		glog.Errorf("Can't decode body: %v", err)
		admissionResponse = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	} else {
		fmt.Println(r.URL.Path)
		if r.URL.Path == "/mutate" {
			admissionResponse = whsvr.mutate(&ar)
		} else if r.URL.Path == "/validate" {
			admissionResponse = whsvr.validate(&ar)
		}
	}

	admissionReview := v1beta1.AdmissionReview{}
	if admissionResponse != nil {
		admissionReview.Response = admissionResponse
		if ar.Request != nil {
			admissionReview.Response.UID = ar.Request.UID
		}
	}

	resp, err := json.Marshal(admissionReview)
	if err != nil {
		glog.Errorf("Can't encode response: %v", err)
		http.Error(w, fmt.Sprintf("could not encode response: %v", err), http.StatusInternalServerError)
	}
	glog.Infof("Ready to write reponse ...")
	if _, err := w.Write(resp); err != nil {
		glog.Errorf("Can't write response: %v", err)
		http.Error(w, fmt.Sprintf("could not write response: %v", err), http.StatusInternalServerError)
	}
}

//zk
func applyDefaultsWorkaround(containers []corev1.Container) {
	defaulter.Default(&corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: containers,
		},
	})
}

func loadConfig(configFile string) (*Config, error) {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	glog.Infof("New configuration: sha256sum %x", sha256.Sum256(data))

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func addContainer(target, added []corev1.Container, basePath string) (patch []patchOperation) {
	first := len(target) == 0
	var value interface{}
	for _, add := range added {
		value = add
		path := basePath
		if first {
			first = false
			value = []corev1.Container{add}
		} else {
			path = path + "/-"
		}
		patch = append(patch, patchOperation{
			Op:    "add",
			Path:  path,
			Value: value,
		})
	}
	return patch
}
func createPodPatch(pod *corev1.Pod, sidecarConfig *Config, annotations map[string]string) ([]byte, error) {
	var patch []patchOperation
	patch = append(patch, addContainer(pod.Spec.Containers, sidecarConfig.Containers, "/spec/containers")...)
	patch = append(patch, updateAnnotation(pod.Annotations, annotations)...)

	return json.Marshal(patch)
}
