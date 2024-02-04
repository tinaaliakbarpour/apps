package initialize

import (
	"context"
	"embed"
	"path/filepath"
	"regexp"

	"github.com/gobuffalo/genny/v2"
	"github.com/gobuffalo/plush/v4"
	"github.com/ignite/cli/ignite/pkg/placeholder"
	"github.com/ignite/cli/ignite/pkg/xgenny"
	"github.com/ignite/cli/ignite/templates/module"
)

const (
	PathAppModule = "app"
	PathAppGo     = "app/app.go"
)

type (
	PlaceHolder        string
	Change             string
	AppDynamicChanges  map[PlaceHolder]Change
	After              string
	Before             string
	AppDynamicReplaces map[Before]After
)

// InitOptions ...
type InitOptions struct {
	AppName string
	AppPath string
	Version string
}

// files/**/* maybe don't needed

//go:embed files/**/*
var files embed.FS

//go:embed placeholders/*
var placeholders embed.FS

// NewGenerator returns the generator to scaffold code to import wasm module inside an app.
func NewGenerator(ctx context.Context, opts InitOptions) (*genny.Generator, error) {
	g := genny.New()

	filePathVersion := "files/" + opts.Version

	filesTemplate := xgenny.NewEmbedWalker(
		files,
		filePathVersion,
		opts.AppPath+"/"+module.PathAppModule,
	)
	if err := g.Box(filesTemplate); err != nil {
		return g, err
	}

	plushCtx := plush.NewContextWithContext(ctx)

	g.Transformer(xgenny.Transformer(plushCtx))

	return g, nil
}

// NewAppModify returns generator with modifications required to register wasmd in the app.
func NewAppModify(replacer placeholder.Replacer, opts InitOptions) *genny.Generator {
	g := genny.New()

	// g.File(placeholders)

	g.RunFn(appModify(replacer, opts))
	return g
}

// app.go modification when importing wasm.
func appModify(replacer placeholder.Replacer, opts InitOptions) genny.RunFn {
	return func(r *genny.Runner) error {

		path := filepath.Join(opts.AppPath, module.PathAppGo)
		f, err := r.Disk.Find(path)
		if err != nil {
			return err
		}
		dc, err := newAppDynamicChange(opts.Version)
		if err != nil {
			return err
		}
		content := f.String()
		for placeholder, text := range dc {
			replacementImport := string(text) + string(placeholder)
			content = replacer.Replace(content, string(placeholder), replacementImport)
		}
		rp, err := newAppDynamicReplaces(opts.Version)
		if err != nil {
			return err
		}
		for before, after := range rp {
			reg, err := regexp.Compile(string(before))
			if err != nil {
				return err
			}
			content = reg.ReplaceAllString(content, string(after))
		}
		newFile := genny.NewFileS(path, content)

		return r.File(newFile)
	}
}

func newAppDynamicChange(version string) (AppDynamicChanges, error) {
	appDynamicChanges := make(AppDynamicChanges)
	ph, err := parsePlaceHoldersFromYaml("placeholders/" + version + ".yaml")
	if err != nil {
		return nil, err
	}
	for _, item := range ph.ChangeItems {
		appDynamicChanges[PlaceHolder(item.PlaceHolder)] = Change(item.Text)
	}

	return appDynamicChanges, nil
}

func newAppDynamicReplaces(version string) (AppDynamicReplaces, error) {
	appDynamicReplaces := make(AppDynamicReplaces)
	rs, err := parseReplacerFromYaml("placeholders/" + version + ".yaml")
	if err != nil {
		return nil, err
	}
	for _, item := range rs.ReplaceItems {
		appDynamicReplaces[Before(item.Before)] = After(item.After)
	}
	return appDynamicReplaces, nil
}
