package lang

import (
	"os"
	"strings"

	"github.com/spf13/afero"
	"golang.org/x/mod/modfile"
)

type LoadGlobalEnvVarOpts struct {
	GoModFileName  string
	DotEnvFileName string
}

func LoadGlobalEnvVars(fs afero.Fs, opts *LoadGlobalEnvVarOpts) (map[string]string, error) {

	if opts == nil {
		opts = &LoadGlobalEnvVarOpts{}
	}

	if opts.GoModFileName == "" {
		opts.GoModFileName = "go.mod"
	}

	if opts.DotEnvFileName == "" {
		opts.DotEnvFileName = ".env"
	}

	env := make(map[string]string)

	// Load environment variables from the file
	if envFile, err := afero.ReadFile(fs, opts.DotEnvFileName); err == nil {
		for _, line := range strings.Split(string(envFile), "\n") {
			if strings.Contains(line, "=") {
				parts := strings.SplitN(line, "=", 2)
				env[parts[0]] = parts[1]
			}
		}
	}

	// Load environment variables from the environment
	for _, envVar := range os.Environ() {
		parts := strings.SplitN(envVar, "=", 2)
		env[parts[0]] = parts[1]
	}

	pkg, err := parseGoMod(fs, opts.GoModFileName)
	if err != nil {
		return nil, err
	}

	env["GO_MODULE_PACKAGE"] = pkg.Module.Mod.Path
	env["GO_MODULE_VERSION"] = pkg.Module.Mod.Version
	env["GO_VERSION"] = pkg.Go.Version

	return env, nil
}

func parseGoMod(fs afero.Fs, name string) (*modfile.File, error) {
	fles, err := afero.ReadFile(fs, name)
	if err != nil {
		return nil, err
	}

	f, err := modfile.Parse(name, fles, nil)
	if err != nil {
		return nil, err
	}

	return f, nil

}
