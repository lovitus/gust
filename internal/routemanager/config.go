package routemanager

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/shlex"
)

const (
	configVersion      = 1
	defaultMTU         = 1420
	portableConfigName = "route-manager.json"
)

type Tunnel struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Routes string `json:"routes"`
	Target string `json:"target"`
	Mode   string `json:"mode,omitempty"`
	Args   string `json:"args,omitempty"`
}

const TunnelModeFree = "free"

type Config struct {
	Version int      `json:"version"`
	Tunnels []Tunnel `json:"tunnels"`
}

type RouteOptions struct {
	Routes []string
	DNS    []string
	MTU    int
}

func DefaultConfigPath() (string, error) {
	userDir, userErr := os.UserConfigDir()
	legacy := ""
	if userErr == nil {
		legacy = filepath.Join(userDir, "gust", portableConfigName)
	}
	executable, executableErr := os.Executable()
	if executableErr == nil {
		portable := filepath.Join(filepath.Dir(executable), portableConfigName)
		if usePortableConfig(portable, legacy) {
			return portable, nil
		}
	}
	if userErr != nil {
		return "", userErr
	}
	return legacy, nil
}

func usePortableConfig(portable, legacy string) bool {
	if info, err := os.Stat(portable); err == nil && !info.IsDir() {
		file, err := os.OpenFile(portable, os.O_WRONLY, 0)
		if err == nil {
			_ = file.Close()
			return true
		}
		return false
	}
	probe, err := os.CreateTemp(filepath.Dir(portable), ".gost-route-manager-write-")
	if err != nil {
		return false
	}
	probePath := probe.Name()
	_ = probe.Close()
	_ = os.Remove(probePath)

	if legacy == "" {
		return true
	}
	data, err := os.ReadFile(legacy)
	if errors.Is(err, os.ErrNotExist) {
		return true
	}
	if err != nil {
		return false
	}
	file, err := os.OpenFile(portable, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if errors.Is(err, os.ErrExist) {
		return true
	}
	if err != nil {
		return false
	}
	if _, err = file.Write(data); err != nil {
		_ = file.Close()
		_ = os.Remove(portable)
		return false
	}
	if err = file.Close(); err != nil {
		_ = os.Remove(portable)
		return false
	}
	return true
}

func Load(path string) (Config, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Config{Version: configVersion}, nil
	}
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("读取配置失败: %w", err)
	}
	if cfg.Version == 0 {
		cfg.Version = configVersion
	}
	return cfg, nil
}

// Save deliberately truncates an existing file instead of replacing it. This
// preserves the original user's ownership when an elevated UI saves changes.
func Save(path string, cfg Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	cfg.Version = configVersion
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err = f.Write(append(b, '\n')); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}

func ParseRouteOptions(input string) (RouteOptions, error) {
	opts := RouteOptions{MTU: defaultMTU}
	dnsContinuation := false
	for _, raw := range strings.Split(input, ",") {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		key, value, hasValue := strings.Cut(item, "=")
		if hasValue {
			dnsContinuation = false
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "dns":
				dnsContinuation = true
				for _, server := range strings.FieldsFunc(value, func(r rune) bool { return r == '|' || r == ';' || r == ' ' }) {
					if net.ParseIP(server) == nil {
						return RouteOptions{}, fmt.Errorf("无效 DNS 地址 %q", server)
					}
					opts.DNS = append(opts.DNS, server)
				}
			case "mtu":
				mtu, err := strconv.Atoi(strings.TrimSpace(value))
				if err != nil || mtu < 576 || mtu > 9000 {
					return RouteOptions{}, fmt.Errorf("MTU 必须在 576 到 9000 之间")
				}
				opts.MTU = mtu
			default:
				return RouteOptions{}, fmt.Errorf("未知路由选项 %q", key)
			}
			continue
		}
		if dnsContinuation && net.ParseIP(item) != nil {
			opts.DNS = append(opts.DNS, item)
			continue
		}
		dnsContinuation = false
		if _, _, err := net.ParseCIDR(item); err != nil {
			return RouteOptions{}, fmt.Errorf("无效路由 %q，请使用 CIDR", item)
		}
		opts.Routes = append(opts.Routes, item)
	}
	if len(opts.Routes) == 0 {
		return RouteOptions{}, errors.New("至少需要一条 CIDR 路由")
	}
	return opts, nil
}

func NormalizeTarget(target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", errors.New("目标 SOCKS 不能为空")
	}
	if !strings.Contains(target, "://") {
		target = "socks5://" + target
	}
	u, err := url.Parse(target)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("无效的 -F 目标 %q", target)
	}
	return u.String(), nil
}

func ParseForwardArgs(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errors.New("目标 SOCKS / 自定义 -F 链不能为空")
	}
	if !strings.HasPrefix(input, "-F") {
		target, err := NormalizeTarget(input)
		if err != nil {
			return nil, err
		}
		return []string{"-F", target}, nil
	}
	args, err := shlex.Split(input)
	if err != nil {
		return nil, fmt.Errorf("解析 -F 链失败: %w", err)
	}
	for i := 0; i < len(args); {
		if strings.HasPrefix(args[i], "-F=") && len(args[i]) > len("-F=") {
			i++
			continue
		}
		if args[i] != "-F" || i+1 >= len(args) || strings.TrimSpace(args[i+1]) == "" {
			return nil, errors.New("自定义目标只能包含一个或多个 -F，例如 -F socks5://a:1080 -F relay+wss://b:443")
		}
		i += 2
	}
	return args, nil
}

func ParseFreeArgs(input string) ([]string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, errors.New("自由 gost 参数不能为空")
	}
	args, err := shlex.Split(input)
	if err != nil {
		return nil, fmt.Errorf("解析自由 gost 参数失败: %w", err)
	}
	if len(args) == 0 {
		return nil, errors.New("自由 gost 参数不能为空")
	}
	first := strings.ToLower(filepath.Base(args[0]))
	if first == "gost" || first == "gost.exe" || first == ManagedBackendName || first == ManagedBackendName+".exe" {
		return nil, errors.New("自由参数前不需要输入 gost，只输入 -L / -F 等参数")
	}
	return args, nil
}

func BuildArgs(t Tunnel) ([]string, error) {
	if strings.TrimSpace(t.Name) == "" {
		return nil, errors.New("记录名字不能为空")
	}
	if t.Mode == TunnelModeFree {
		return ParseFreeArgs(t.Args)
	}
	opts, err := ParseRouteOptions(t.Routes)
	if err != nil {
		return nil, err
	}
	forwardArgs, err := ParseForwardArgs(t.Target)
	if err != nil {
		return nil, err
	}

	h := sha256.Sum256([]byte(t.ID + "\x00" + t.Name))
	// Allocate a stable /30 inside 198.18.0.0/15 for each saved tunnel.
	block := (uint32(h[0])<<8 | uint32(h[1])) % (1 << 15)
	base := uint32(198)<<24 | uint32(18)<<16 | block<<2
	tunIP := net.IPv4(byte(base>>24), byte(base>>16), byte(base>>8), byte(base+1)).String() + "/30"
	ifName := fmt.Sprintf("grm%x", h[:5])

	u := &url.URL{Scheme: "tungo", Host: ":0"}
	q := u.Query()
	q.Set("name", ifName)
	q.Set("net", tunIP)
	q.Set("routes", strings.Join(opts.Routes, ","))
	q.Set("mtu", strconv.Itoa(opts.MTU))
	if len(opts.DNS) > 0 {
		q.Set("dns", strings.Join(opts.DNS, ","))
	}
	u.RawQuery = q.Encode()
	return append([]string{"-L", u.String()}, forwardArgs...), nil
}
