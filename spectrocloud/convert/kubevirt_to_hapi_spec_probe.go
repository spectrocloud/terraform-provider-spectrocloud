package convert

import (
	"github.com/spectrocloud/hapi/models"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	kubevirtapiv1 "kubevirt.io/api/core/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func ToHapiVmProbe(probe *kubevirtapiv1.Probe) *models.V1VMProbe {
	if probe == nil {
		return nil
	}

	return &models.V1VMProbe{
		Exec:                ToHapiVmExecAction(probe.Exec),
		FailureThreshold:    probe.FailureThreshold,
		GuestAgentPing:      probe.GuestAgentPing,
		HTTPGet:             ToHapiVmHttpGetHandlerAction(probe.HTTPGet),
		InitialDelaySeconds: probe.InitialDelaySeconds,
		PeriodSeconds:       probe.PeriodSeconds,
		SuccessThreshold:    probe.SuccessThreshold,
		TCPSocket:           ToHapiVmTcpSocketHandlerAction(probe.TCPSocket),
		TimeoutSeconds:      probe.TimeoutSeconds,
	}
}

func ToHapiVmTcpSocketHandlerAction(socket *v1.TCPSocketAction) *models.V1VMTCPSocketAction {
	if socket == nil {
		return nil
	}

	return &models.V1VMTCPSocketAction{
		Host: socket.Host,
		Port: ToHapiVmIntOrString(socket.Port),
	}
}

func ToHapiVmHttpGetHandlerAction(get *v1.HTTPGetAction) *models.V1VMHTTPGetAction {
	if get == nil {
		return nil
	}

	return &models.V1VMHTTPGetAction{
		Host:        get.Host,
		Path:        get.Path,
		Port:        ToHapiVmIntOrString(get.Port),
		Scheme:      string(get.Scheme),
		HTTPHeaders: ToHapiVmHTTPHeaders(get.HTTPHeaders),
	}
}

func ToHapiVmHTTPHeaders(headers []v1.HTTPHeader) []*models.V1VMHTTPHeader {
	var Headers []*models.V1VMHTTPHeader
	for _, header := range headers {
		Headers = append(Headers, &models.V1VMHTTPHeader{
			Name:  types.Ptr(header.Name),
			Value: types.Ptr(header.Value),
		})
	}
	return Headers
}

func ToHapiVmIntOrString(port intstr.IntOrString) *string {
	if port.Type == intstr.Int {
		return types.Ptr(string(port.IntVal))
	}
	return types.Ptr(port.StrVal)
}

func ToHapiVmExecAction(exec *v1.ExecAction) *models.V1VMExecAction {
	if exec == nil {
		return nil
	}

	return &models.V1VMExecAction{
		Command: exec.Command,
	}
}

func ToHapiVmTopologySpreadConstraints(constraints []v1.TopologySpreadConstraint) []*models.V1VMTopologySpreadConstraint {
	var Constraints []*models.V1VMTopologySpreadConstraint
	for _, constraint := range constraints {
		Constraints = append(Constraints, &models.V1VMTopologySpreadConstraint{
			LabelSelector:     ToHapiVmLabelSelector(constraint.LabelSelector),
			MaxSkew:           &constraint.MaxSkew,
			TopologyKey:       types.Ptr(constraint.TopologyKey),
			WhenUnsatisfiable: types.Ptr(string(constraint.WhenUnsatisfiable)),
		})
	}
	return Constraints
}

func ToHapiVmTolerations(tolerations []v1.Toleration) []*models.V1VMToleration {
	var Tolerations []*models.V1VMToleration
	for _, toleration := range tolerations {
		var TolerationSeconds int64
		if toleration.TolerationSeconds != nil {
			TolerationSeconds = *toleration.TolerationSeconds
		}
		Tolerations = append(Tolerations, &models.V1VMToleration{
			Effect:            string(toleration.Effect),
			Key:               toleration.Key,
			Operator:          string(toleration.Operator),
			TolerationSeconds: TolerationSeconds,
			Value:             toleration.Value,
		})
	}
	return Tolerations
}
