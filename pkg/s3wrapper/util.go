package s3wrapper

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/openshift/assisted-service/internal/isoeditor"
	"github.com/openshift/assisted-service/internal/isoutil"
	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"
)

const (
	minimalTemplatesVersionFileName = "minimal_templates_version.json"
	minimalTemplatesVersionLatest   = 2 // increase if templates update is needed
)

type templatesVersion struct {
	Version int
}

func FixEndpointURL(endpoint string) (string, error) {
	_, err := url.ParseRequestURI(endpoint)
	if err == nil {
		return endpoint, nil
	}

	prefix := "http://"
	if os.Getenv("S3_USE_SSL") == "true" {
		prefix = "https://"
	}

	new_url := prefix + endpoint
	_, err = url.ParseRequestURI(new_url)
	if err != nil {
		return "", err
	}
	return new_url, nil
}

func ExtractBootFilesFromISOAndUpload(ctx context.Context, log logrus.FieldLogger, isoFileName, isoObjectName, isoURL string, api API) error {
	isoHandler := isoutil.NewHandler(isoFileName, "")

	for fileType := range ISOFileTypes {
		objectName := BootFileTypeToObjectName(isoObjectName, fileType)
		exists, err := api.DoesPublicObjectExist(ctx, objectName)
		if err != nil {
			return errors.Wrapf(err, "Failed searching for object %s", objectName)
		}
		if exists {
			log.Infof("Object %s already exists, skipping upload", objectName)
			continue
		}
		log.Infof("Starting to upload %s from Base ISO %s", fileType, isoObjectName)
		err = uploadFileFromISO(ctx, isoHandler, fileType, objectName, api)
		if err != nil {
			log.WithError(err).Fatalf("Failed to extract and upload file %s from ISO", fileType)
		}

		log.Infof("Successfully uploaded object %s", objectName)
	}
	return nil
}

func DownloadURLToTemporaryFile(url string) (string, error) {
	tmpfile, err := ioutil.TempFile("", "isodownload")
	if err != nil {
		return "", errors.Wrap(err, "Error creating temporary file")
	}
	defer tmpfile.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", errors.Wrapf(err, "Failed fetching from URL %s", url)
	}

	_, err = io.Copy(tmpfile, resp.Body)
	if err != nil {
		return "", errors.Wrapf(err, "Failed downloading file from %s to %s", url, tmpfile.Name())
	}

	return tmpfile.Name(), nil
}

func UploadFromURLToPublicBucket(ctx context.Context, objectName, url string, api API) error {
	resp, err := http.Get(url)
	if err != nil {
		return errors.Wrapf(err, "Failed fetching from URL %s", url)
	}

	err = api.UploadStreamToPublicBucket(ctx, resp.Body, objectName)
	if err != nil {
		return errors.Wrapf(err, "Failed uploading to %s", objectName)
	}

	return nil
}

func uploadFileFromISO(ctx context.Context, isoHandler isoutil.Handler, fileType, objectName string, api API) error {
	filename := ISOFileTypes[fileType]
	reader, err := isoHandler.ReadFile(filename)
	if err != nil {
		return errors.Wrapf(err, "Failed to read file %s from ISO", filename)
	}

	err = api.UploadStreamToPublicBucket(ctx, reader, objectName)
	if err != nil {
		return err
	}
	return nil
}

func BootFileTypeToObjectName(isoObjectName, fileType string) string {
	return strings.TrimSuffix(isoObjectName, ".iso") + "." + fileType
}

func DoAllBootFilesExist(ctx context.Context, isoObjectName string, api API) (bool, error) {
	for _, fileType := range BootFileExtensions {
		objectName := BootFileTypeToObjectName(isoObjectName, fileType)
		exists, err := api.DoesPublicObjectExist(ctx, objectName)
		if err != nil {
			log.Error(err)
			return false, err
		}
		if !exists {
			return false, nil
		}
	}
	return true, nil
}

func CreateAndUploadMinimalIso(ctx context.Context, log logrus.FieldLogger,
	isoPath, minimalIsoObject, openshiftVersion, serviceBaseURL string,
	api API, editorFactory isoeditor.Factory) error {

	log.Infof("Extracting rhcos ISO (%s)", isoPath)
	var minimalIsoPath string
	err := editorFactory.WithEditor(ctx, isoPath, openshiftVersion, log, func(editor isoeditor.Editor) error {
		var createError error
		minimalIsoPath, createError = editor.CreateMinimalISOTemplate(serviceBaseURL)
		return createError
	})
	if err != nil {
		log.Errorf("Error extracting rhcos ISO (%v)", err)
		return err
	}
	defer os.Remove(minimalIsoPath)

	// upload the minimal iso
	log.Infof("Uploading minimal ISO (%s)", minimalIsoPath)
	if err := api.UploadFileToPublicBucket(ctx, minimalIsoPath, minimalIsoObject); err != nil {
		return err
	}

	// Update version file in bucket
	return updateISOTemplatesVersion(ctx, log, api)
}

// HaveLatestMinimalTemplate Returns true if latest version already exists in bucket; otherwise, false.
func HaveLatestMinimalTemplate(ctx context.Context, log logrus.FieldLogger, api API) bool {
	versionFromBucket, err := getISOTemplatesVersion(ctx, log, api)
	if err != nil {
		log.WithError(err).Warning("ISO template version metadata not found, uploading latest")
		// Assume error is due to missing file and upload the latest
		return false
	}

	// We assume that the code contains the latest version
	if versionFromBucket.Version < minimalTemplatesVersionLatest {
		log.Warnf("Templates version is stale: %d", versionFromBucket.Version)
		return false
	}

	// No need to update
	return true
}

func getISOTemplatesVersion(ctx context.Context, log logrus.FieldLogger, api API) (*templatesVersion, error) {
	// Fetch version file from bucket
	reader, _, err := api.Download(ctx, minimalTemplatesVersionFileName)
	if err != nil {
		return nil, err
	}

	// Read and parse version
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	var version templatesVersion
	err = json.Unmarshal(bytes, &version)
	if err != nil {
		return nil, err
	}

	return &version, nil
}

func updateISOTemplatesVersion(ctx context.Context, log logrus.FieldLogger, api API) error {
	currentVersion := &templatesVersion{
		Version: minimalTemplatesVersionLatest,
	}
	b, err := json.Marshal(currentVersion)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(b)
	return api.UploadStream(ctx, reader, minimalTemplatesVersionFileName)
}
