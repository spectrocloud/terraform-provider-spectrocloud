package spectrocloud

//func TestToAddonDeployment(t *testing.T) {
//	assert := assert.New(t)
//
//	// Create a mock ResourceData object
//	d := prepareAddonDeploymentTestData("depl-test-id")
//
//	m := &client.V1Client{}
//
//	addonDeployment, err := toAddonDeployment(m, d)
//	assert.Nil(err)
//
//	// Verifying apply setting
//	assert.Equal(d.Get("apply_setting"), addonDeployment.SpcApplySettings.ActionType)
//
//	// Verifying cluster profile
//	profiles := d.Get("cluster_profile").([]interface{})
//	for i, profile := range profiles {
//		p := profile.(map[string]interface{})
//		assert.Equal(p["id"].(string), addonDeployment.Profiles[i].UID)
//
//		// Verifying pack values
//		packValues := p["pack"].([]interface{})
//		for j, pack := range packValues {
//			packMap := pack.(map[string]interface{})
//			assert.Equal(packMap["name"], *addonDeployment.Profiles[i].PackValues[j].Name)
//			assert.Equal(packMap["tag"], addonDeployment.Profiles[i].PackValues[j].Tag)
//			assert.Equal(packMap["type"], string(addonDeployment.Profiles[i].PackValues[j].Type))
//			assert.Equal(packMap["values"], addonDeployment.Profiles[i].PackValues[j].Values)
//		}
//
//	}
//}
