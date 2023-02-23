package spectrocloud

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/hapi/client"
	"github.com/spectrocloud/hapi/models"
)

func toClusterLocationConfigs(d *schema.ResourceData) *models.V1ClusterLocation {
	if d.Get("location_config") != nil {
		for _, locationConfig := range d.Get("location_config").([]interface{}) {
			return toClusterLocationConfig(locationConfig.(map[string]interface{}))
		}
	}

	return nil
}

func toClusterLocationConfig(config map[string]interface{}) *models.V1ClusterLocation {

	countryCode := ""
	if config["country_code"] != nil {
		countryCode = config["country_code"].(string)
	}

	countryName := ""
	if config["country_name"] != nil {
		countryName = config["country_name"].(string)
	}

	regionCode := ""
	if config["region_code"] != nil {
		regionCode = config["region_code"].(string)
	}

	regionName := ""
	if config["region_name"] != nil {
		regionName = config["region_name"].(string)
	}

	return &models.V1ClusterLocation{
		CountryCode: countryCode,
		CountryName: countryName,
		GeoLoc:      toClusterGeoLoc(config),
		RegionCode:  regionCode,
		RegionName:  regionName,
	}
}

func toClusterGeoLoc(config map[string]interface{}) *models.V1GeolocationLatlong {
	var latitude float64
	if config["latitude"] != nil {
		latitude = config["latitude"].(float64)
	}

	var longitude float64
	if config["longitude"] != nil {
		longitude = config["longitude"].(float64)
	}

	return &models.V1GeolocationLatlong{
		Latitude:  latitude,
		Longitude: longitude,
	}
}

func flattenLocationConfig(location *models.V1ClusterLocation) []interface{} {
	result := make(map[string]interface{})
	configs := make([]interface{}, 0)

	if location.GeoLoc != nil {
		result["latitude"] = location.GeoLoc.Latitude
		result["longitude"] = location.GeoLoc.Longitude
	}

	result["country_code"] = location.CountryCode
	result["country_name"] = location.CountryName
	result["region_code"] = location.RegionCode
	result["region_name"] = location.RegionName

	configs = append(configs, result)

	return configs
}

func updateLocationConfig(c *client.V1Client, d *schema.ResourceData) error {
	if locationConfigs := toClusterLocationConfigs(d); locationConfigs != nil {
		return c.ApplyClusterLocationConfig(d.Id(), &models.V1SpectroClusterLocationInputEntity{
			Location: locationConfigs,
		})
	}
	return nil
}
