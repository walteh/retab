package buildrc

import (
	"context"
	"runtime"
	"strings"

	"github.com/go-faster/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/buildrc/pkg/git"
	"golang.org/x/tools/go/packages"
)

var (
	ErrCouldNotParseRemoteURL = errors.Errorf("could not parse remote url")
)

type BuildrcJSON struct {
	Version              string   `json:"version"`
	Revision             string   `json:"revision"`
	Executable           string   `json:"executable"`
	Org                  string   `json:"org"`
	Artifact             string   `json:"artifact"`
	GoPkg                string   `json:"go-pkg"`
	Name                 string   `json:"name"`
	Image                string   `json:"image"`
	TargetPlatform       string   `json:"target-platform"`
	TargetPlatformOutDir string   `json:"target-platform-out-dir"`
	BuildPlatform        string   `json:"build-platform"`
	GoTestablePackages   []string `json:"go-testable-packages"`
}

type BuildrcPackageName string

type BuildrcVersion string

func GetArtifactName(_ context.Context, name string, version string, plat *Platform) string {
	return name + "-" + version + "-" + plat.DashString()
}

func GetRevision(ctx context.Context, gitp git.GitProvider) (string, error) {

	revision, err := gitp.GetCurrentCommitFromRef(ctx, "HEAD")
	if err != nil {
		return "", err
	}

	return revision, nil
}

func GetExecutable(_ context.Context, name string) string {
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name
}

func GetRepo(ctx context.Context, gitp git.GitProvider) (string, string, error) {

	url, err := gitp.GetRemoteURL(ctx)
	if err != nil {
		return "", "", err
	}

	parts := strings.Split(url, "/")

	if len(parts) < 2 {
		return "", "", ErrCouldNotParseRemoteURL
	}

	org := parts[len(parts)-2]

	if strings.Contains(org, ":") && !strings.HasSuffix(org, ":") {
		org = strings.Split(org, ":")[1]
	}

	trimmed := strings.TrimSuffix(parts[len(parts)-1], ".git")

	return org, trimmed, nil
}

func GetGoPkg(_ context.Context, gitp git.GitProvider) (string, error) {

	fle, err := afero.ReadFile(gitp.Fs(), "go.mod")
	if err != nil {
		return "", err
	}

	// find the line with module on it
	lines := strings.Split(string(fle), "\n")

	var modine string

	for _, line := range lines {
		if strings.HasPrefix(line, "module") {
			modine = line
			break
		}
	}

	if modine == "" {
		return "", errors.Errorf("could not find module line in go.mod")
	}

	// split on space
	parts := strings.Split(modine, " ")

	if len(parts) != 2 {
		return "", errors.Errorf("could not parse module line in go.mod")
	}

	return parts[1], nil

}

func GetTestableGoPackages(ctx context.Context, gitp git.GitProvider) ([]string, error) {

	path := ""

	if fls, ok := gitp.Fs().(*afero.BasePathFs); ok {
		tmp, err := fls.RealPath("")
		if err != nil {
			return nil, err
		}
		path = tmp
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode:    packages.NeedName,
		Tests:   true,
		Dir:     path,
		Context: ctx,
	}, "./...")
	if err != nil {
		return nil, err
	}

	resp := []string{}

	for _, pkg := range pkgs {
		if !strings.HasSuffix(pkg.PkgPath, ".test") || strings.Contains(pkg.PkgPath, "/vendor/") || strings.Contains(pkg.PkgPath, "/gen/") {
			continue
		}
		resp = append(resp, strings.TrimSuffix(pkg.PkgPath, ".test"))
	}

	return resp, nil
}

func GetBuildrcJSON(ctx context.Context, gitp git.GitProvider, opts *GetVersionOpts) (*BuildrcJSON, error) {

	brc, err := LoadBuildrc(ctx, gitp)
	if err != nil {
		zerolog.Ctx(ctx).Debug().Err(err).Msg("could not load buildrc")
		return nil, err
	}

	tplat, err := GetTargetPlatform(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Debug().Err(err).Msg("could not get target platform")
		return nil, err
	}

	bplat, err := GetBuildPlatform(ctx)
	if err != nil {
		zerolog.Ctx(ctx).Debug().Err(err).Msg("could not get build platform")
		return nil, err
	}

	version, err := GetVersion(ctx, gitp, brc, opts)
	if err != nil {
		zerolog.Ctx(ctx).Debug().Err(err).Msg("could not get version")
		return nil, err
	}

	org, name, err := GetRepo(ctx, gitp)
	if err != nil {
		zerolog.Ctx(ctx).Debug().Err(err).Msg("could not get repo")
		return nil, err
	}

	revision, err := GetRevision(ctx, gitp)
	if err != nil {
		zerolog.Ctx(ctx).Debug().Err(err).Msg("could not get revision")
		return nil, err
	}

	goPkg, err := GetGoPkg(ctx, gitp)
	if err != nil {
		zerolog.Ctx(ctx).Debug().Err(err).Msg("could not get go pkg")
		return nil, err
	}

	goTestablePackages, err := GetTestableGoPackages(ctx, gitp)
	if err != nil {
		zerolog.Ctx(ctx).Debug().Err(err).Msg("could not get go testable packages")
		return nil, err
	}

	exec := GetExecutable(ctx, name)

	artif := GetArtifactName(ctx, name, version, tplat)

	return &BuildrcJSON{
		Version:              version,
		Revision:             revision,
		Executable:           exec,
		Image:                org + "/" + name,
		Artifact:             artif,
		GoPkg:                goPkg,
		Name:                 name,
		Org:                  org,
		TargetPlatform:       tplat.String(),
		TargetPlatformOutDir: tplat.UnderscoreString(),
		BuildPlatform:        bplat.String(),
		GoTestablePackages:   goTestablePackages,
	}, nil
}
