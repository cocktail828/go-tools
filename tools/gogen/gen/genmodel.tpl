type {{ .Name }} struct { {{ range .Fields }}
    {{ .Name }} {{ .Type }} {{ if .Tag }}`{{ .Tag }}`{{ end }}{{ if .Comment }}// {{ .Comment }}{{ end }}{{ end }}
}
