package service

const systemdSandboxScript = `[Unit]
Description={{.Description}}
ConditionFileIsExecutable={{.Path|cmdEscape}}
{{range $i, $dep := .Dependencies}}
{{$dep}} {{end}}

[Service]
StartLimitInterval=5
StartLimitBurst=10
ExecStart={{.Path|cmdEscape}}{{range .Arguments}} {{.|cmd}}{{end}}
{{if .ChRoot}}RootDirectory={{.ChRoot|cmd}}{{end}}
{{if .WorkingDirectory}}WorkingDirectory={{.WorkingDirectory|cmdEscape}}{{end}}
{{if .UserName}}User={{.UserName}}{{end}}
{{if .ReloadSignal}}ExecReload=/bin/kill -{{.ReloadSignal}} "$MAINPID"{{end}}
{{if .PIDFile}}PIDFile={{.PIDFile|cmd}}{{end}}
{{if and .LogOutput .HasOutputFileSupport -}}
StandardOutput=file:{{.LogDirectory}}/{{.Name}}.out
StandardError=file:{{.LogDirectory}}/{{.Name}}.err
{{- end}}
{{if gt .LimitNOFILE -1 }}LimitNOFILE={{.LimitNOFILE}}{{end}}
{{if .Restart}}Restart={{.Restart}}{{end}}
{{if .SuccessExitStatus}}SuccessExitStatus={{.SuccessExitStatus}}{{end}}
RestartSec=120
EnvironmentFile=-/etc/sysconfig/{{.Name}}
{{range $k, $v := .EnvVars -}}
Environment={{$k}}={{$v}}
{{end -}}

NoNewPrivileges=true
PrivateTmp=true
PrivateDevices=true
ProtectSystem=strict
ProtectControlGroups=true
ProtectKernelModules=true
ProtectKernelTunables=true
RestrictSUIDSGID=true
LockPersonality=true
UMask=0077
{{if .WorkingDirectory}}ReadWritePaths={{.WorkingDirectory|cmdEscape}}{{end}}

[Install]
WantedBy=multi-user.target
`

const launchdSandboxConfig = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Disabled</key>
	<false/>
	{{- if .EnvVars}}
	<key>EnvironmentVariables</key>
	<dict>
		{{- range $k, $v := .EnvVars}}
		<key>{{html $k}}</key>
		<string>{{html $v}}</string>
		{{- end}}
	</dict>
	{{- end}}
	<key>KeepAlive</key>
	<{{bool .KeepAlive}}/>
	<key>Label</key>
	<string>{{html .Name}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{html .Path}}</string>
		{{- if .Config.Arguments}}
		{{- range .Config.Arguments}}
		<string>{{html .}}</string>
		{{- end}}
	{{- end}}
	</array>
	<key>ProcessType</key>
	<string>Background</string>
	<key>Umask</key>
	<integer>63</integer>
	<key>HardResourceLimits</key>
	<dict>
		<key>NumberOfFiles</key>
		<integer>4096</integer>
	</dict>
	<key>SoftResourceLimits</key>
	<dict>
		<key>NumberOfFiles</key>
		<integer>2048</integer>
	</dict>
	{{- if .ChRoot}}
	<key>RootDirectory</key>
	<string>{{html .ChRoot}}</string>
	{{- end}}
	<key>RunAtLoad</key>
	<{{bool .RunAtLoad}}/>
	<key>SessionCreate</key>
	<{{bool .SessionCreate}}/>
	{{- if .StandardErrorPath}}
	<key>StandardErrorPath</key>
	<string>{{html .StandardErrorPath}}</string>
	{{- end}}
	{{- if .StandardOutPath}}
	<key>StandardOutPath</key>
	<string>{{html .StandardOutPath}}</string>
	{{- end}}
	{{- if .UserName}}
	<key>UserName</key>
	<string>{{html .UserName}}</string>
	{{- end}}
	{{- if .WorkingDirectory}}
	<key>WorkingDirectory</key>
	<string>{{html .WorkingDirectory}}</string>
	{{- end}}
</dict>
</plist>
`
