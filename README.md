# Resizer

A http image resizer

Resizers images as well as normalising exif jpegs

## Config

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