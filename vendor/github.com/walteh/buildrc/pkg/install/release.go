package install

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/afero"
	"github.com/walteh/buildrc/pkg/file"
)

func InstallLatestGithubRelease(ctx context.Context, ofs afero.Fs, fls afero.Fs, org string, name string, token string) error {

	var err error

	// get the latest release
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/"+org+"/"+name+"/releases/latest", nil)
	if err != nil {
		return err
	}

	req.Header.Add("Accept", "application/json")

	if token != "" {
		req.Header.Add("Authorization", token)
	}

	dat := &http.Client{}
	resp, err := dat.Do(req)
	if err != nil {
		return err
	}

	var respdata struct {
		Assets []struct {
			BrowserDownloadURL string `json:"browser_download_url"`
			Name               string `json:"name"`
		} `json:"assets"`
		URL string `json:"url"`
	}

	bdy, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(bdy, &respdata)
	if err != nil {
		zerolog.Ctx(ctx).Error().Str("payload", string(bdy)).Err(err).Msg("error unmarshalling")

		return err
	}

	zerolog.Ctx(ctx).Debug().Interface("respdata", respdata).Msg("got respdata")

	targetPlat := runtime.GOOS + "-" + runtime.GOARCH

	if os.Getenv("GOARM") != "" {
		targetPlat += "-" + os.Getenv("GOARM")
	}

	dl := ""

	for _, asset := range respdata.Assets {
		if strings.HasSuffix(asset.Name, targetPlat+".tar.gz") {
			dl = asset.BrowserDownloadURL
			break
		}
	}

	if dl == "" {
		return fmt.Errorf("no release found for %s", targetPlat)
	}

	fle, err := downloadFile(ctx, fls, dl)
	if err != nil {
		return err
	}

	defer fle.Close()

	// untar the release
	out, err := file.Untargz(ctx, fls, fle.Name())
	if err != nil {
		return err
	}

	// install the release
	err = InstallAs(ctx, ofs, fls, out.Name(), name)
	if err != nil {
		return err
	}

	return nil

}

func downloadFile(ctx context.Context, fls afero.Fs, str string) (fle afero.File, err error) {

	base := filepath.Base(str)

	// Create the file
	out, err := afero.TempDir(fls, "", "")
	if err != nil {
		return nil, err
	}

	fle, err = fls.Create(filepath.Join(out, base))
	if err != nil {
		return nil, err
	}

	// Get the data
	resp, err := http.Get(str)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(fle, resp.Body)
	if err != nil {
		return nil, err
	}

	return fle, nil
}
