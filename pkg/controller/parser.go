package controller

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"strconv"

	"github.com/DevopsArtFactory/bigshot/pkg/schema"
)

// ParseTemplates parses items to list of template
func ParseTemplates(items []map[string]*dynamodb.AttributeValue) ([]schema.Config, error) {
	var configs []schema.Config

	for _, item := range items {
		config, err := ChangeItemToConfig(item)
		if err != nil {
			return nil, err
		}
		configs = append(configs, *config)
	}

	return configs, nil
}

// ChangeItemToConfig changes item value from dynamoDB to schema.Config
func ChangeItemToConfig(item map[string]*dynamodb.AttributeValue) (*schema.Config, error) {
	config := &schema.Config{
		Name: *item["name"].S,
	}

	interval, err := strconv.Atoi(*item["interval"].N)
	if err != nil {
		return nil, err
	}
	config.Interval = interval

	timout, err := strconv.Atoi(*item["timeout"].N)
	if err != nil {
		return nil, err
	}
	config.Timeout = timout

	regions := []schema.Region{}
	for _, region := range item["regions"].L {
		regions = append(regions, schema.Region{Region: *region.M["region"].S})
	}
	config.Regions = regions

	targets := []schema.Target{}
	for _, target := range item["targets"].L {
		t := schema.Target{
			URL:    *target.M["url"].S,
			Method: *target.M["method"].S,
		}

		if _, ok := target.M["header"]; ok {
			headers := map[string]string{}
			for key, val := range target.M["header"].M {
				headers[key] = *val.S
			}
			t.Header = headers
		}

		if _, ok := target.M["body"]; ok {
			bodies := map[string]string{}
			for key, val := range target.M["body"].M {
				bodies[key] = *val.S
			}
			t.Body = bodies
		}

		targets = append(targets, t)
	}
	config.Targets = targets

	return config, nil
}
