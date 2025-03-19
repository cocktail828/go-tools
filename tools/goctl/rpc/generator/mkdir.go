package generator

import (
	"path/filepath"
	"strings"

	"github.com/cocktail828/go-tools/tools/goctl/internal/ctx"
	"github.com/cocktail828/go-tools/tools/goctl/internal/pathx"
	"github.com/cocktail828/go-tools/tools/goctl/rpc/parser"
	"github.com/cocktail828/go-tools/z/stringx"
)

const (
	wd       = "wd"
	etc      = "etc"
	internal = "internal"
	_config  = "config"
	logic    = "logic"
	server   = "server"
	svc      = "svc"
	pb       = "pb"
	protoGo  = "proto-go"
	call     = "call"
)

type (
	// DirContext defines a rpc service directories context
	DirContext interface {
		GetCall() Dir
		GetEtc() Dir
		GetInternal() Dir
		GetConfig() Dir
		GetLogic() Dir
		GetServer() Dir
		GetSvc() Dir
		GetPb() Dir
		GetProtoGo() Dir
		GetMain() Dir
		GetServiceName() string
		SetPbDir(pbDir, grpcDir string)
	}

	// Dir defines a directory
	Dir struct {
		Base            string
		Filename        string
		Package         string
		GetChildPackage func(childPath string) (string, error)
	}

	defaultDirContext struct {
		inner       map[string]Dir
		serviceName string
		ctx         *ctx.ProjectContext
	}
)

func mkdir(ctx *ctx.ProjectContext, proto parser.Proto, c *ZRpcContext) (DirContext,
	error) {
	inner := make(map[string]Dir)
	etcDir := filepath.Join(ctx.WorkDir, "etc")
	clientDir := filepath.Join(ctx.WorkDir, "client")
	internalDir := filepath.Join(ctx.WorkDir, "internal")
	configDir := filepath.Join(internalDir, "config")
	logicDir := filepath.Join(internalDir, "logic")
	serverDir := filepath.Join(internalDir, "server")
	svcDir := filepath.Join(internalDir, "svc")
	pbDir := filepath.Join(ctx.WorkDir, proto.GoPackage)
	protoGoDir := pbDir
	if c != nil {
		pbDir = c.ProtoGenGrpcDir
		protoGoDir = c.ProtoGenGoDir
	}

	getChildPackage := func(parent, childPath string) (string, error) {
		child := strings.TrimPrefix(childPath, parent)
		abs := filepath.Join(parent, strings.ToLower(child))
		if c.Multiple {
			if err := pathx.MkdirIfNotExist(abs); err != nil {
				return "", err
			}
		}
		childPath = strings.TrimPrefix(abs, ctx.Dir)
		pkg := filepath.Join(ctx.Path, childPath)
		return filepath.ToSlash(pkg), nil
	}

	var callClientDir string
	if !c.Multiple {
		callDir := filepath.Join(ctx.WorkDir,
			strings.ToLower(stringx.ToCamel(proto.Service[0].Name)))
		if strings.EqualFold(proto.Service[0].Name, filepath.Base(proto.GoPackage)) {
			clientDir = stringx.ToSnake(proto.Service[0].Name + "_client")
			callDir = filepath.Join(ctx.WorkDir, clientDir)
		}
		callClientDir = callDir
	} else {
		callClientDir = clientDir
	}
	if c.IsGenClient {
		inner[call] = Dir{
			Filename: callClientDir,
			Package: filepath.ToSlash(filepath.Join(ctx.Path,
				strings.TrimPrefix(callClientDir, ctx.Dir))),
			Base: filepath.Base(callClientDir),
			GetChildPackage: func(childPath string) (string, error) {
				return getChildPackage(callClientDir, childPath)
			},
		}
	}

	inner[wd] = Dir{
		Filename: ctx.WorkDir,
		Package: filepath.ToSlash(filepath.Join(ctx.Path,
			strings.TrimPrefix(ctx.WorkDir, ctx.Dir))),
		Base: filepath.Base(ctx.WorkDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(ctx.WorkDir, childPath)
		},
	}
	inner[etc] = Dir{
		Filename: etcDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(etcDir, ctx.Dir))),
		Base:     filepath.Base(etcDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(etcDir, childPath)
		},
	}
	inner[internal] = Dir{
		Filename: internalDir,
		Package: filepath.ToSlash(filepath.Join(ctx.Path,
			strings.TrimPrefix(internalDir, ctx.Dir))),
		Base: filepath.Base(internalDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(internalDir, childPath)
		},
	}
	inner[_config] = Dir{
		Filename: configDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(configDir, ctx.Dir))),
		Base:     filepath.Base(configDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(configDir, childPath)
		},
	}
	inner[logic] = Dir{
		Filename: logicDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(logicDir, ctx.Dir))),
		Base:     filepath.Base(logicDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(logicDir, childPath)
		},
	}
	inner[server] = Dir{
		Filename: serverDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(serverDir, ctx.Dir))),
		Base:     filepath.Base(serverDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(serverDir, childPath)
		},
	}
	inner[svc] = Dir{
		Filename: svcDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(svcDir, ctx.Dir))),
		Base:     filepath.Base(svcDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(svcDir, childPath)
		},
	}

	inner[pb] = Dir{
		Filename: pbDir,
		Package:  filepath.ToSlash(filepath.Join(ctx.Path, strings.TrimPrefix(pbDir, ctx.Dir))),
		Base:     filepath.Base(pbDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(pbDir, childPath)
		},
	}

	inner[protoGo] = Dir{
		Filename: protoGoDir,
		Package: filepath.ToSlash(filepath.Join(ctx.Path,
			strings.TrimPrefix(protoGoDir, ctx.Dir))),
		Base: filepath.Base(protoGoDir),
		GetChildPackage: func(childPath string) (string, error) {
			return getChildPackage(protoGoDir, childPath)
		},
	}

	for _, v := range inner {
		err := pathx.MkdirIfNotExist(v.Filename)
		if err != nil {
			return nil, err
		}
	}
	serviceName := strings.TrimSuffix(proto.Name, filepath.Ext(proto.Name))
	return &defaultDirContext{
		ctx:         ctx,
		inner:       inner,
		serviceName: strings.ReplaceAll(serviceName, "-", ""),
	}, nil
}

func (d *defaultDirContext) SetPbDir(pbDir, grpcDir string) {
	d.inner[pb] = Dir{
		Filename: pbDir,
		Package:  filepath.ToSlash(filepath.Join(d.ctx.Path, strings.TrimPrefix(pbDir, d.ctx.Dir))),
		Base:     filepath.Base(pbDir),
	}

	d.inner[protoGo] = Dir{
		Filename: grpcDir,
		Package: filepath.ToSlash(filepath.Join(d.ctx.Path,
			strings.TrimPrefix(grpcDir, d.ctx.Dir))),
		Base: filepath.Base(grpcDir),
	}
}

func (d *defaultDirContext) GetCall() Dir {
	return d.inner[call]
}

func (d *defaultDirContext) GetEtc() Dir {
	return d.inner[etc]
}

func (d *defaultDirContext) GetInternal() Dir {
	return d.inner[internal]
}

func (d *defaultDirContext) GetConfig() Dir {
	return d.inner[_config]
}

func (d *defaultDirContext) GetLogic() Dir {
	return d.inner[logic]
}

func (d *defaultDirContext) GetServer() Dir {
	return d.inner[server]
}

func (d *defaultDirContext) GetSvc() Dir {
	return d.inner[svc]
}

func (d *defaultDirContext) GetPb() Dir {
	return d.inner[pb]
}

func (d *defaultDirContext) GetProtoGo() Dir {
	return d.inner[protoGo]
}

func (d *defaultDirContext) GetMain() Dir {
	return d.inner[wd]
}

func (d *defaultDirContext) GetServiceName() string {
	return d.serviceName
}

// Valid returns true if the directory is valid
func (d *Dir) Valid() bool {
	return len(d.Filename) > 0 && len(d.Package) > 0
}
