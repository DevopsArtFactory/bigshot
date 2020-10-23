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

package templates

// TracingTemplate is a template for aws provider
const TracingTemplate = `{{ decorate "bold" "Check IP" }}: {{ format .Summary.TracingData.ConnectAddr }}
{{ decorate "bold" "Status Code" }}: {{ format .Summary.Response.StatusCode }}
{{ decorate "bold" "Status Message" }}: {{ format .Summary.Response.StatusMsg }}
{{- range $key, $val := .Summary.Response.Header }}
{{ decorate "bold" $key }}: {{ format $val }}
{{- end }}
`
