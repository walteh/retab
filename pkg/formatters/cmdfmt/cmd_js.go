//go:build js && wasm

package cmdfmt

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"syscall/js"

	"github.com/rs/zerolog"
	"gitlab.com/tozd/go/errors"
)

func runBasicCmd(ctx context.Context, cmds []string) (string, error) {
	res := js.Global().Call("retab_exec_basic", strings.Join(cmds, " "))
	if res.IsUndefined() {
		zerolog.Ctx(ctx).Error().Msg("failed to execute command")
		return "", errors.New("failed to execute command")
	}

	if res.Type().String() != "string" {
		zerolog.Ctx(ctx).Error().Msg("command returned non-string result")
		return "", errors.New("command returned non-string result")
	}

	return res.String(), nil
}

func runFmtCmd(ctx context.Context, cmds []string, w io.Writer, r io.Reader, opts *BasicExternalFormatterOpts) error {
	zerolog.Ctx(ctx).Info().Msg("reading inputz")

	data, err := io.ReadAll(r)
	if err != nil {
		zerolog.Ctx(ctx).Error().Msg("failed to read input")
		return errors.Errorf("failed to read input: %w", err)
	}

	zerolog.Ctx(ctx).Info().Msg("executing command: " + strings.Join(cmds, " "))

	var marsh []byte = []byte("{}")
	if len(opts.tempFiles) > 0 {
		marsh, err = json.Marshal(opts.tempFiles)
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
