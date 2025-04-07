package spectrocloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToTagsMap(t *testing.T) {
	tagRD := getBaseResourceData()
	tagMap := map[string]string{
		"my:key1":         "my:value1",
		"my@key:longer:2": "my@value:longer:2",
		"mykey":           "",
	}
	err := tagRD.Set("tags_map", tagMap)
	if err != nil {
		assert.Fail(t, "Error setting tags_map.")
	}
	tags := toTagsMap(tagRD)
	assert.Equal(t, tags["my:key1"], "my:value1")
	assert.Equal(t, tags["my@key:longer:2"], "my@value:longer:2")
	assert.Equal(t, tags["mykey"], "spectro__tag")
}

func TestFlattenTagsMap(t *testing.T) {
	tagMap := make(map[string]string)
	tagMap["unittest"] = "spectro__tag"
	tagMap["owner"] = "siva"
	tags := flattenTagsMap(tagMap)

	// Check that the tags slice contains the expected tags, regardless of order
	assert.Contains(t, tags, "unittest")
	assert.Contains(t, tags, "owner:"+tagMap["owner"])
}

func TestToMergedTags(t *testing.T) {
	tagRD := getBaseResourceData()
	tags := []string{"mykeytooverwrite:overwriteme", "mytagskey:mytagsvalue"}
	tagMap := map[string]string{
		"my@long:key:colons": "my@long:values:2",
		"mykey":              "",
		"mykeytooverwrite":   "overwrittenbymergefrommap",
	}
	err := tagRD.Set("tags", tags)
	if err != nil {
		assert.Fail(t, "Error setting tags.")
	}
	err = tagRD.Set("tags_map", tagMap)
	if err != nil {
		assert.Fail(t, "Error setting tags_map.")
	}
	mergedTags := toMergedTags(tagRD)
	assert.Equal(t, mergedTags["my@long:key:colons"], "my@long:values:2")        // multiple colons in key and value
	assert.Equal(t, mergedTags["mykey"], "spectro__tag")                         // empty value default
	assert.Equal(t, mergedTags["mykeytooverwrite"], "overwrittenbymergefrommap") // overwrite kv in tags if exists in tags_map
	assert.Equal(t, mergedTags["mytagskey"], "mytagsvalue")                      // preserve kv in tags on merge
}

func TestToMergedTagsNilMap(t *testing.T) {
	tagRD := getBaseResourceData()
	tags := []string{"mykeytooverwrite:overwriteme", "mytagskey:mytagsvalue"}
	err := tagRD.Set("tags", tags)
	if err != nil {
		assert.Fail(t, "Error setting tags.")
	}
	err = tagRD.Set("tags_map", nil)
	if err != nil {
		assert.Fail(t, "Error setting tags_map.")
	}
	mergedTags := toMergedTags(tagRD)
	assert.Equal(t, mergedTags["mykeytooverwrite"], "overwriteme") // overwrite kv in tags if exists in tags_map
	assert.Equal(t, mergedTags["mytagskey"], "mytagsvalue")        // preserve kv in tags on merge
}
