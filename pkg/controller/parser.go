/*
Copyright 2020 The bigshot Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package controller

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"

	"github.com/DevopsArtFactory/bigshot/pkg/schema"
)

// ParseTemplates parses items to list of template
func ParseTemplates(items []map[string]*dynamodb.AttributeValue) ([]schema.Template, error) {
	var templates []schema.Template

	for _, item := range items {
		template, err := ChangeItemToTempalte(item)
		if err != nil {
			return nil, err
		}
		templates = append(templates, *template)
	}

	return templates, nil
}

// ChangeItemToTempalte changes item value from dynamoDB to schema.Template
func ChangeItemToTempalte(item map[string]*dynamodb.AttributeValue) (*schema.Template, error) {
	var template schema.Template
	if err := dynamodbattribute.UnmarshalMap(item, &template); err != nil {
		return nil, err
	}


	return &template, nil
}
