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
const TracingTemplate = `{{ decorate "bold" "Domain" }}: {{ format .Summary.TracingData.URL }}
{{ decorate "bold" "Check IP" }}: {{ format .Summary.TracingData.ConnectAddr }}
{{ decorate "bold" "Status Code" }}: {{ format .Summary.Response.StatusCode }}
{{ decorate "bold" "Status Message" }}: {{ format .Summary.Response.StatusMsg }}
`

// ListTemplate is a template of listing bigshot worker settings
const ListTemplate = `{{ decorate "bold underline" "List" }}
{{- range $item := .Summary }} 
====================================================
{{ decorate "bold" "Name" }}: {{ format $item.Name }}
{{ decorate "bold" "Timeout" }}: {{ format $item.Timeout }}
{{ decorate "bold" "Interval" }}: {{ format $item.Interval }}
{{ decorate "bold" "Regions" }}
{{- if eq (len $item.Regions ) 0 }}
  No worker exists
{{- else }}
  {{- range $region := $item.Regions }}
    - {{ format $region.Region }}
  {{- end }}
{{- end }}

{{ decorate "bold" "Targets" }}
{{- if eq (len $item.Targets ) 0 }}
  No url is targeted
{{- else }}
  {{- range $target := $item.Targets }}
    - {{ format $target.Method }} {{ format $target.URL }}
  {{- end }}
{{- end }}

{{ decorate "bold" "SlackURLs" }}
{{- if eq (len $item.SlackURLs ) 0 }}
No slack alarm exists
{{- else }}
  {{- range $alarm := $item.SlackURLs }}
    - {{ format $alarm }}
  {{- end }}
{{- end }}
{{- end }}
`
