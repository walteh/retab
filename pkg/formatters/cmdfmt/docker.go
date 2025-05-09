package cmdfmt

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/walteh/retab/v2/pkg/format"
)

type DockerExternalFormatter struct {
	Image         string
	Command       []string
	Once          sync.Once
	containerName string
	startError    error
	internal      format.Provider
}

func NewDockerCmdFormatter(command []string, optz ...OptBasicExternalFormatterOptsSetter) *DockerExternalFormatter {

	opts := NewBasicExternalFormatterOpts(optz...)

	if !opts.useDocker {
		panic("useDocker is false")
	}

	if opts.dockerImageName == "" {
		panic("dockerImageName is empty")
	}

	if opts.dockerImageTag == "" {
		panic("dockerImageTag is empty")
	}

	if opts.executable == "" {
		panic("executable is empty")
	}

	cmds := append([]string{opts.executable}, command...)

	containerName := "retab_" + xid.New().String()

	fmtCmds := []string{"docker", "run", "--interactive", "--quiet", "--name", containerName, fmt.Sprintf("%s:%s", opts.dockerImageName, opts.dockerImageTag)}
	fmtCmds = append(fmtCmds, cmds...)

	basic := NewCmdFormatter(fmtCmds, optz...)

	return &DockerExternalFormatter{
		Image:         fmt.Sprintf("%s:%s", opts.dockerImageName, opts.dockerImageTag),
		Command:       fmtCmds,
		containerName: containerName,
		internal:      basic,
	}
}

func (me *DockerExternalFormatter) Format(ctx context.Context, cfg format.Configuration, input io.Reader) (io.Reader, error) {
	if me.startError != nil {
		return nil, me.startError
	}

	// me.Once.Do(func() {
	// 	// create the container
	// 	cmds := []string{"docker", "run", "--name", me.containerName, me.Image}
	// 	cmds = append(cmds, me.Command...)
	// 	out, err := runBasicCmd(ctx, cmds)
	// 	if err != nil {
	// 		zerolog.Ctx(ctx).Error().Err(err).Str("container", me.containerName).Msg("failed to start docker container")
	// 		me.startError = err
	// 	} else {
	// 		zerolog.Ctx(ctx).Info().Str("container", me.containerName).Msgf("docker container started: %s", out)
	// 		go func() {
	// 			// wait for our go process to exit (signal)and clean up docker container
	// 			sig := make(chan os.Signal, 1)
	// 			signal.Notify(sig, os.Interrupt)
	// 			<-sig
	// 			_, err := runBasicCmd(ctx, []string{"docker", "rm", "--force", me.containerName})
	// 			if err != nil {
	// 				zerolog.Ctx(ctx).Error().Err(err).Str("container", me.containerName).Msg("failed to remove docker container")
	// 			}
	// 		}()
	// 	}
	// })

	zerolog.Ctx(ctx).Info().Str("container", me.containerName).Strs("command", me.Command).Msg("formatting with docker")

	return me.internal.Format(ctx, cfg, input)
}
