# Resizer

A http image resizer

Resizers images as well as normalising exif jpegs

## Config

Resizer can be configured by file `./config.yml`, `/etc/resizer/config.yml` or using
environment variables. For example `source_type` becomes `RESIZER_SOURCE_TYPE`

Currently Resizer supports two sources.

### Filesystem

```
source_type: filesystem
filesystem_source_path: ./images
```

### S3

```
source_type: s3
s3_source_bucket: my-bucket
s3_source_region: ap-southeast-2
s3_source_prefix: images/
```

## API

### `GET /images/{path}{?w,h,raw}`

The query params avalible are

- `w` width
- `h` height
- `raw` return the raw source image (won't strip exif)

### `POST /images`

Uploads an image normalising exif jpegs along the way