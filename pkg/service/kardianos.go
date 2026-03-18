package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	kservice "github.com/kardianos/service"
	"github.com/sandevgo/tuskbot/internal/core"
)

type KardianosManager struct {
	svc kservice.Service
}

var _ core.SystemService = (*KardianosManager)(nil)

func NewKardianosManager(cfg Config) (*KardianosManager, error) {
	options := kservice.KeyValue{
		"WorkingDirectory": cfg.WorkingDirectory,
		"UserService":      cfg.UserService,
	}
	if cfg.SandboxMode {
		options["SystemdScript"] = systemdSandboxScript
		options["LaunchdConfig"] = launchdSandboxConfig
	}

	svcCfg := &kservice.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
		Arguments:   cfg.Arguments,
		Option:      options,
	}
	if cfg.Executable != "" {
		svcCfg.Executable = cfg.Executable
	}

	svc, err := kservice.New(nil, svcCfg)
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	return &KardianosManager{svc: svc}, nil
}

func (m *KardianosManager) Install(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	status, err := m.Status(ctx)
	if err == nil && status != core.SystemServiceStatusUnknown {
		return nil
	}

	if err := m.svc.Install(); err != nil {
		if isAlreadyExistsErr(err) {
			return nil
		}
		return err
	}
	return nil
}

func (m *KardianosManager) Uninstall(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	status, err := m.Status(ctx)
	if err == nil && status == core.SystemServiceStatusUnknown {
		return nil
	}

	if err := m.svc.Uninstall(); err != nil {
		if isNotInstalledErr(err) {
			return nil
		}
		return err
	}
	return nil
}

func (m *KardianosManager) Start(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	status, err := m.Status(ctx)
	if err == nil && status == core.SystemServiceStatusRunning {
		return nil
	}

	if err := m.svc.Start(); err != nil {
		if isAlreadyRunningErr(err) {
			return nil
		}
		return err
	}
	return nil
}

func (m *KardianosManager) Stop(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	status, err := m.Status(ctx)
	if err == nil && status != core.SystemServiceStatusRunning {
		return nil
	}

	if err := m.svc.Stop(); err != nil {
		if isNotRunningErr(err) {
			return nil
		}
		return err
	}
	return nil
}

func (m *KardianosManager) Status(ctx context.Context) (core.SystemServiceStatus, error) {
	if err := ctx.Err(); err != nil {
		return core.SystemServiceStatusUnknown, err
	}

	st, err := m.svc.Status()
	if err != nil {
		if isNotInstalledErr(err) {
			return core.SystemServiceStatusUnknown, nil
		}
		return core.SystemServiceStatusUnknown, err
	}

	switch st {
	case kservice.StatusRunning:
		return core.SystemServiceStatusRunning, nil
	case kservice.StatusStopped:
		return core.SystemServiceStatusStopped, nil
	default:
		return core.SystemServiceStatusUnknown, nil
	}
}

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

func isAlreadyExistsErr(err error) bool {
	if err == nil {
		return false
	}
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "already exists") || strings.Contains(e, "already installed") || strings.Contains(e, "service exists")
}

func isNotInstalledErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, kservice.ErrNotInstalled) {
		return true
	}
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "not installed") || strings.Contains(e, "does not exist") || strings.Contains(e, "could not find")
}

func isAlreadyRunningErr(err error) bool {
	if err == nil {
		return false
	}
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "already started") || strings.Contains(e, "already running") || strings.Contains(e, "service is running")
}

func isNotRunningErr(err error) bool {
	if err == nil {
		return false
	}
	e := strings.ToLower(err.Error())
	return strings.Contains(e, "not loaded") || strings.Contains(e, "not running") || strings.Contains(e, "not started") || strings.Contains(e, "input/output error")
}
