//go:build js

package cmdfmt

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"syscall/js"

	"github.com/rs/zerolog"
	"github.com/walteh/retab/v2/pkg/format"
	"gitlab.com/tozd/go/errors"
)

func NewExecFormatter(ctx context.Context, opts *BasicExternalFormatterOpts, cmds ...string) format.Provider {
	return ExternalFormatterToProvider(&basicExternalFormatter{opts.Indent, opts.TempFiles, func(r io.Reader, w io.Writer) func() error {
		if len(cmds) < 1 {
			return func() error {
				return errors.New("no command specified")
			}
		}

		cmd := func() error {
			zerolog.Ctx(ctx).Info().Msg("reading inputz")

			data, err := io.ReadAll(r)
			if err != nil {
				zerolog.Ctx(ctx).Error().Msg("failed to read input")
				return errors.Errorf("failed to read input: %w", err)
			}

			zerolog.Ctx(ctx).Info().Msg("executing command: " + strings.Join(cmds, " "))

			var marsh []byte = []byte("{}")
			if len(opts.TempFiles) > 0 {
				marsh, err = json.Marshal(opts.TempFiles)
				if err != nil {
					zerolog.Ctx(ctx).Error().Msg("failed to marshal input")
					return errors.Errorf("failed to marshal input: %w", err)
				}
			}

			res := js.Global().Call("retab_exec", strings.Join(cmds, " "), string(data), string(marsh))
			if res.IsUndefined() {
				zerolog.Ctx(ctx).Error().Msg("failed to execute command")
				return errors.New("failed to execute command")
			}

			if res.Type().String() != "string" {
				zerolog.Ctx(ctx).Error().Msg("command returned non-string result")
				return errors.New("command returned non-string result")
			}

			str := res.String()
			if str == "" {
				zerolog.Ctx(ctx).Error().Msg("command returned empty string")
				return errors.New("command returned empty string")
			}

			if strings.HasPrefix(str, "error:") {
				zerolog.Ctx(ctx).Error().Msg("command returned error: " + str)
				return errors.New("command returned error: " + str)
			}

			zerolog.Ctx(ctx).Info().Msg("command result: " + str)

			_, err = io.WriteString(w, str)
			if err != nil {
				zerolog.Ctx(ctx).Error().Msg("failed to write command result")
				return errors.Errorf("failed to write command result: %w", err)
			}

			return nil
		}

		return cmd
	}})
}
