{{/*
Expand the name of the chart.
*/}}
{{- define "mailit.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
*/}}
{{- define "mailit.fullname" -}}
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
Common labels
*/}}
{{- define "mailit.labels" -}}
helm.sh/chart: {{ include "mailit.name" . }}-{{ .Chart.Version }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}

{{/*
API selector labels
*/}}
{{- define "mailit.api.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mailit.name" . }}-api
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Web selector labels
*/}}
{{- define "mailit.web.selectorLabels" -}}
app.kubernetes.io/name: {{ include "mailit.name" . }}-web
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
PostgreSQL host
*/}}
{{- define "mailit.postgresql.host" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "%s-postgresql" (include "mailit.fullname" .) }}
{{- else }}
{{- .Values.postgresql.external.host }}
{{- end }}
{{- end }}

{{/*
PostgreSQL port
*/}}
{{- define "mailit.postgresql.port" -}}
{{- if .Values.postgresql.enabled }}
{{- printf "5432" }}
{{- else }}
{{- .Values.postgresql.external.port | toString }}
{{- end }}
{{- end }}

{{/*
Redis address
*/}}
{{- define "mailit.redis.addr" -}}
{{- if .Values.redis.enabled }}
{{- printf "%s-redis-master:6379" (include "mailit.fullname" .) }}
{{- else }}
{{- .Values.redis.external.addr }}
{{- end }}
{{- end }}
