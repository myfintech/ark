package cloudutils

import (
	"context"
	"io"

	"github.com/pkg/errors"
	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"

	"github.com/moby/buildkit/util/appcontext"
	"github.com/myfintech/ark/src/go/lib/log"
)

var appCTX = appcontext.Context()

// BlobCheck queries a binary storage location for a designated file
func BlobCheck(ctx context.Context, bucketURL, checkString string) (bool, error) {
	if ctx == nil {
		ctx = appCTX
	}

	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		return false, errors.Wrap(err, "unable to open bucket for check")
	}
	defer bucket.Close()
	log.Debugf("checking existence of %s/%s", bucketURL, checkString)
	return bucket.Exists(ctx, checkString)
}

// NewBlobReader pulls a file from a binary storage location
func NewBlobReader(ctx context.Context, bucketURL, fileKey string) (io.Reader, func(), error) {
	if ctx == nil {
		ctx = appCTX
	}

	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to open bucket for check")
	}

	reader, err := bucket.NewReader(ctx, fileKey, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to create reader for file")
	}
	log.Debugf("preparing reader for %s/%s", bucketURL, fileKey)
	return reader, func() {
		bucket.Close()
		reader.Close()
	}, nil
}

// NewBlobWriter pushes a file from a local location to a binary storage location
func NewBlobWriter(ctx context.Context, bucketURL, fileKey string) (io.Writer, func(), error) {
	if ctx == nil {
		ctx = appCTX
	}

	bucket, err := blob.OpenBucket(ctx, bucketURL)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to open bucket for writing to remote destination")
	}

	writer, err := bucket.NewWriter(ctx, fileKey, nil)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unable to create writer for destination file")
	}
	log.Debugf("preparing writer for %s/%s", bucketURL, fileKey)
	return writer, func() {
		bucket.Close()
		writer.Close()
	}, nil
}

// DeleteBlob removes a blob from a bucket
func DeleteBlob(ctx context.Context, bucketURL, fileKey string) error {
	if ctx == nil {
		ctx = appCTX
	}

	bucket, err := blob.OpenBucket(ctx, bucketURL)
	defer func() {
		_ = bucket.Close()
	}()
	if err != nil {
		return err
	}

	if err = bucket.Delete(ctx, fileKey); err != nil {
		return err
	}

	return nil
}
