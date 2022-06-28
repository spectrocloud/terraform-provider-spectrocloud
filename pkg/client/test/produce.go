package test

import (
	"fmt"
	"github.com/spectrocloud/hapi/apiutil/transport"
	"github.com/spectrocloud/hapi/models"
	clusterC "github.com/spectrocloud/hapi/spectrocluster/client/v1"
	userC "github.com/spectrocloud/hapi/user/client/v1"
	"reflect"
)

func prepareClusterMethod(clusterClient clusterC.ClientService, params interface{}, name string) (reflect.Value, []reflect.Value) {
	method := reflect.ValueOf(clusterClient).MethodByName(name)
	fmt.Println("method type num out:", method.Type().NumOut())
	return prepareParams(method, params)
}

func prepareUserMethod(userClient userC.ClientService, params interface{}, name string) (reflect.Value, []reflect.Value) {
	method := reflect.ValueOf(userClient).MethodByName(name)
	fmt.Println("method type num out:", method.Type().NumOut())
	return prepareParams(method, params)
}

func prepareParams(method reflect.Value, params interface{}) (reflect.Value, []reflect.Value) {
	in := make([]reflect.Value, method.Type().NumIn())
	fmt.Println("method type num in:", method.Type().NumIn())
	for i := 0; i < method.Type().NumIn(); i++ {
		object := params
		fmt.Println(i, "->", object)
		in[i] = reflect.ValueOf(object)
	}
	return method, in
}

func produceResults(retry Retry, method reflect.Value, in []reflect.Value, ch chan int, done chan bool) {
	for i := 0; i < retry.runs; i++ {
		go func(chnl chan int) {

			result := method.Call(in)
			err := result[1].Interface()
			fmt.Println(result[0].Convert(*models.V1SpectroClustersUIDConfigNamespacesGetOK))
			if err != nil {
				if _, ok := err.(*transport.TcpError); ok {
					chnl <- -1
					return
				}
				if _, ok := err.(*transport.TransportError); ok && err.(*transport.TransportError).HttpCode == retry.expected_code {
					chnl <- retry.expected_code
					return
				} else {
					chnl <- 500
					return
				}
			} else {
				chnl <- 200
				return
			}
		}(ch)
	}
	done <- true
}
