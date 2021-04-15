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

{{- define "orchestrator.eventForwarder.selectorLabels" -}}
app.kubernetes.io/component: "event-forwarder"
{{ include "orchestrator.common.selectorLabels" . }}
{{- end -}}

{{- define "orchestrator.eventForwarder.labels" -}}
{{ include "orchestrator.eventForwarder.selectorLabels" . }}
{{ include "orchestrator.common.labels" . }}
{{- end -}}

{{- define "orchestrator.rabbitmqOperator.selectorLabels" -}}
app.kubernetes.io/component: "rabbitmq-operator"
{{ include "orchestrator.common.selectorLabels" . }}
{{- end -}}

{{- define "orchestrator.rabbitmqOperator.labels" -}}
{{ include "orchestrator.rabbitmqOperator.selectorLabels" . }}
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
Create a fully qualified orchestrator server name
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "orchestrator.eventForwarder.fullname" -}}
{{- if .Values.forwarder.fullnameOverride }}
{{- .Values.forwarder.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- printf "%s-event-forwarder" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s-event-forwarder" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create a fully qualified orchestrator server name
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "orchestrator.rabbitmqOperator.fullname" -}}
{{- if .Values.rabbitmqOperator.fullnameOverride }}
{{- .Values.rabbitmqOperator.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- printf "%s-rabbitmq-operator" .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s-rabbitmq-operator" .Release.Name $name | trunc 63 | trimSuffix "-" }}
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
