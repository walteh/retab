package editorconfig

// // NewDynamicConfigurationProvider creates a new configuration provider based on the provided options
// func NewDynamicConfigurationProvider(ctx context.Context, fs afero.Fs, opts DynamicConfigOptions) (format.ConfigurationProvider, error) {
// 	logger := zerolog.Ctx(ctx)

// 	// 1. Use direct content if provided
// 	if opts.RawContent != "" {
// 		logger.Debug().Msg("using provided editorconfig content")
// 		return NewEditorConfigConfigurationProviderFromContent(ctx, opts.RawContent)
// 	}

// 	// 2. Use specific config path if provided
// 	if opts.FilePath != "" {
// 		logger.Debug().Str("path", opts.FilePath).Msg("reading editorconfig from specified path")
// 		file, err := fs.Open(opts.FilePath)
// 		if err != nil {
// 			return nil, errors.Errorf("opening specified editorconfig file: %w", err)
// 		}
// 		defer file.Close()

// 		x, err2, err := editorconfig.ParseGraceful(file)
// 		if err != nil {
// 			return nil, errors.Errorf("getting editorconfig definition: %w", err)
// 		}
// 		if err2 != nil {
// 			return nil, errors.Errorf("parsing editorconfig: %w", err2)
// 		}

// 		return &EditorConfigConfigurationProvider{definitions: x}, nil
// 	}

// 	// 3. Auto-resolve from file path up to search root or git root
// 	if opts.TargetFile != "" {
// 		logger.Debug().Str("file", opts.TargetFile).Str("root", opts.SearchRoot).Msg("auto-resolving editorconfig")
// 		return autoResolveConfig(ctx, fs, opts.TargetFile, opts.SearchRoot, opts.IgnoreGitRoot)
// 	}

// 	// 4. Fallback to defaults
// 	logger.Debug().Msg("no editorconfig options provided, using defaults")
// 	return &EditorConfigConfigurationDefaults{
// 		Defaults: format.NewBasicConfigurationProvider(true, 4, true, false),
// 	}, nil
// }

// // isGitRoot checks if the given directory contains a .git directory
// func isGitRoot(fs afero.Fs, dir string) (bool, error) {
// 	gitPath := filepath.Join(dir, ".git")
// 	exists, err := afero.Exists(fs, gitPath)
// 	if err != nil {
// 		return false, errors.Errorf("checking git directory existence: %w", err)
// 	}
// 	return exists, nil
// }

// // findGitRoot walks up the directory tree to find the nearest directory containing .git
// func findGitRoot(fs afero.Fs, startDir string) (string, error) {
// 	currentDir := startDir
// 	for {
// 		isGit, err := isGitRoot(fs, currentDir)
// 		if err != nil {
// 			return "", err
// 		}
// 		if isGit {
// 			return currentDir, nil
// 		}

// 		parentDir := filepath.Dir(currentDir)
// 		if parentDir == currentDir {
// 			return "", nil // Reached root without finding .git
// 		}
// 		currentDir = parentDir
// 	}
// }

// // autoResolveConfig walks up the directory tree to find the nearest .editorconfig file
// func autoResolveConfig(ctx context.Context, fs afero.Fs, filePath string, searchRoot string, ignoreGitRoot bool) (format.ConfigurationProvider, error) {
// 	logger := zerolog.Ctx(ctx)
// 	currentDir := filepath.Dir(filePath)

// 	// Determine the root directory to stop at
// 	var rootDir string
// 	if searchRoot != "" {
// 		rootDir = searchRoot
// 	} else {
// 		if !ignoreGitRoot {
// 			// Find git root if no search root specified
// 			gitRoot, err := findGitRoot(fs, currentDir)
// 			if err != nil {
// 				return nil, errors.Errorf("finding git root: %w", err)
// 			}
// 			if gitRoot != "" {
// 				rootDir = gitRoot
// 				logger.Debug().Str("gitRoot", gitRoot).Msg("using git root as search boundary")
// 			} else {
// 				rootDir = filepath.Dir(currentDir) // Default to parent of current dir if no git root found
// 			}
// 		} else {
// 			rootDir = filepath.Dir(currentDir)
// 		}
// 	}

// 	for currentDir != rootDir && currentDir != filepath.Dir(currentDir) {
// 		configPath := filepath.Join(currentDir, ".editorconfig")
// 		logger.Debug().Str("path", configPath).Msg("checking for editorconfig")

// 		exists, err := afero.Exists(fs, configPath)
// 		if err != nil {
// 			return nil, errors.Errorf("checking editorconfig existence: %w", err)
// 		}

// 		if exists {
// 			file, err := fs.Open(configPath)
// 			if err != nil {
// 				return nil, errors.Errorf("opening editorconfig file: %w", err)
// 			}
// 			defer file.Close()

// 			x, err2, err := editorconfig.ParseGraceful(file)
// 			if err != nil {
// 				return nil, errors.Errorf("getting editorconfig definition: %w", err)
// 			}
// 			if err2 != nil {
// 				return nil, errors.Errorf("parsing editorconfig: %w", err2)
// 			}

// 			logger.Debug().Str("path", configPath).Msg("found editorconfig")
// 			return &EditorConfigConfigurationProvider{definitions: x}, nil
// 		}

// 		currentDir = filepath.Dir(currentDir)
// 	}

// 	// No editorconfig found, use defaults
// 	logger.Debug().Msg("no editorconfig found in path, using defaults")
// 	return &EditorConfigConfigurationDefaults{
// 		Defaults: format.NewBasicConfigurationProvider(true, 4, true, false),
// 	}, nil
// }
