package routes

import (
	"net/http"
	"strconv"
)

func CustomClusterRoutes() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/cloudTypes/{cloudType}",
			Response: ResponseData{
				StatusCode: 201,
				Payload:    map[string]string{"UID": "custom-cluster-1"},
			},
		},
	}
}

func CustomClusterRoutesNegative() []Route {
	return []Route{
		{
			Method: "POST",
			Path:   "/v1/spectroclusters/cloudTypes/{cloudType}",
			Response: ResponseData{
				StatusCode: http.StatusConflict,
				Payload:    getError(strconv.Itoa(http.StatusConflict), "Cluster already exist"),
			},
		},
	}
}
