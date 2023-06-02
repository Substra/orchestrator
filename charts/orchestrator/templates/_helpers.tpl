{{/*
Expand the name of the chart.
*/}}
{{- define "orchestrator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "orchestrator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create unified labels for orchestrator components
*/}}
{{- define "orchestrator.common.selectorLabels" -}}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/name: {{ template "orchestrator.name" . }}
{{- end -}}

{{- define "orchestrator.common.labels" -}}
helm.sh/chart: {{ include "orchestrator.chart" . }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{- define "orchestrator.server.selectorLabels" -}}
app.kubernetes.io/component: "server"
{{ include "orchestrator.common.selectorLabels" . }}
{{- end -}}

{{- define "orchestrator.server.labels" -}}
{{ include "orchestrator.server.selectorLabels" . }}
{{ include "orchestrator.common.labels" . }}
{{- end -}}

{{- define "orchestrator.migrations.labels" -}}
app.kubernetes.io/component: "migrations"
{{ include "orchestrator.common.labels" . }}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "orchestrator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create a fully qualified orchestrator server name
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "orchestrator.server.fullname" -}}
{{- if .Values.orchestrator.fullnameOverride }}
{{- .Values.orchestrator.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- printf "%s-server" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s-server" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}


{{/*
Create a fully qualified migration job name
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "orchestrator.migrations.fullname" -}}
{{- if .Values.migrations.fullnameOverride }}
{{- .Values.migrations.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- printf "%s-migrations" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s-migrations" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}


{{/*
Create the name of the service account to use
*/}}
{{- define "orchestrator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "orchestrator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}


{{/*
Generate ingress backend entry that is compatible with all ports types.
Usage:
{{ include "orchestrator.ingress.backend" (dict "serviceName" "backendName" "servicePort" "backendPort") }}

Params:
  - serviceName - String. Name of an existing service backend
  - servicePort - String/Int. Port name (or number) of the service. It will be translated to different yaml depending if it is a string or an integer.
*/}}
{{- define "orchestrator.ingress.backend" -}}
service:
  name: {{ .serviceName }}
  port:
    {{- if typeIs "string" .servicePort }}
    name: {{ .servicePort }}
    {{- else if or (typeIs "int" .servicePort) (typeIs "float64" .servicePort) }}
    number: {{ .servicePort | int }}
    {{- end }}
{{- end -}}


{{/*
Renders a value that contains template.
Usage:
{{ include "common.tplvalues.render" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "common.tplvalues.render" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}


{{/*
Return the proper image name, with option for a default tag
example:
    {{ include "substra-orc.images.name" (dict "img" .Values.path.to.the.image "defaultTag" $.Chart.AppVersion) }}
*/}}
{{- define "substra-orc.images.name" -}}
    {{- $tag := (.img.tag | default .defaultTag) }}
    {{- if .img.registry -}}
    {{- printf "%s/%s:%s" .img.registry .img.repository $tag -}}
    {{- else -}}
    {{- printf "%s:%s" .img.repository $tag -}}
    {{- end -}}
{{- end -}}


{{- define "substra-orc.postgresql.secret-name" -}}
    {{- if .Values.postgresql.auth.credentialsSecretName -}}
        {{- .Values.postgresql.auth.credentialsSecretName }}
    {{- else -}}
        {{- template "orchestrator.server.fullname" . }}-database
    {{- end -}}
{{- end -}}

{{/*
The hostname we should connect to (external is defined, otherwise integrated)
*/}}
{{- define "substra-orc.postgresql.host" -}}
    {{- if .Values.postgresql.host }}
        {{- .Values.postgresql.host }}
    {{- else }}
        {{- template "postgresql.primary.fullname" (index .Subcharts "integrated-postgresql") }}.{{ .Release.Namespace }}
    {{- end }}
{{- end -}}

{{/*
Disable SSL if using the integrated Postgres, otherwise leave users with the option of setting their own.
*/}}
{{- define "substra-orc.postgresql.connectionParameters" -}}
    {{- if .Values.postgresql.connectionParameters -}}
        {{ .Values.postgresql.connectionParameters }}
    {{- else if index .Values "integrated-postgresql" "enabled" -}}
        sslmode=disable
    {{- end }}
{{- end -}}