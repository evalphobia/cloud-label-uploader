cloud-label-uploader
----

[![GoDoc][1]][2] [![License: MIT][3]][4] [![Release][5]][6] [![Build Status][7]][8] [![Go Report Card][13]][14] [![Code Climate][19]][20] [![BCH compliance][21]][22]

[1]: https://godoc.org/github.com/evalphobia/cloud-label-uploader?status.svg
[2]: https://godoc.org/github.com/evalphobia/cloud-label-uploader
[3]: https://img.shields.io/badge/License-MIT-blue.svg
[4]: LICENSE.md
[5]: https://img.shields.io/github/release/evalphobia/cloud-label-uploader.svg
[6]: https://github.com/evalphobia/cloud-label-uploader/releases/latest
[7]: https://travis-ci.org/evalphobia/cloud-label-uploader.svg?branch=master
[8]: https://travis-ci.org/evalphobia/cloud-label-uploader
[9]: https://coveralls.io/repos/evalphobia/cloud-label-uploader/badge.svg?branch=master&service=github
[10]: https://coveralls.io/github/evalphobia/cloud-label-uploader?branch=master
[11]: https://codecov.io/github/evalphobia/cloud-label-uploader/coverage.svg?branch=master
[12]: https://codecov.io/github/evalphobia/cloud-label-uploader?branch=master
[13]: https://goreportcard.com/badge/github.com/evalphobia/cloud-label-uploader
[14]: https://goreportcard.com/report/github.com/evalphobia/cloud-label-uploader
[15]: https://img.shields.io/github/downloads/evalphobia/cloud-label-uploader/total.svg?maxAge=1800
[16]: https://github.com/evalphobia/cloud-label-uploader/releases
[17]: https://img.shields.io/github/stars/evalphobia/cloud-label-uploader.svg
[18]: https://github.com/evalphobia/cloud-label-uploader/stargazers
[19]: https://codeclimate.com/github/evalphobia/cloud-label-uploader/badges/gpa.svg
[20]: https://codeclimate.com/github/evalphobia/cloud-label-uploader
[21]: https://bettercodehub.com/edge/badge/evalphobia/cloud-label-uploader?branch=master
[22]: https://bettercodehub.com/

`cloud-label-uploader` download files from url in CSV.
And create CSV file with label and path for Google Cloud AutoML.

# Installation

Install cloud-label-uploader by command below,

```bash
$ go get github.com/evalphobia/cloud-label-uploader
```

# Usage

## root command

```bash
$ cloud-label-uploader
Commands:

  help       show help
  download   Download files from --file csv
  list       Create list file from --input dir images
  upload     Upload files to Cloud Bucket(S3, GCS) from --input dir
  vott       Create object-detection list file from VoTT results
```

## download command

```bash
$ cloud-label-uploader help download
Download files from --file csv

Options:

  -h, --help           display help information
  -i, --input         *image list file --input='/path/to/dir/input.csv'
  -n, --name          *column name for filename --name='name'
  -l, --label         *column name for label --label='group'
  -u, --url           *column name for URL --url='path'
  -m, --parallel[=2]   parallel number (multiple download) --parallel=2
  -o, --output         outout dir --output='/path/to/dir/'
```

```bash
# Save CSV file with name, label and URL.
$ cat my_file_list.csv

id,label,image_url
1,cat,http://example.com/foo.jpg
2,dog,http://example.com/bar.jpg
3,cat,https://example.com/foo2.JPG
4,human,https://example.com/baz.png?q=1
5,human,https://example.com/baz2.png


# Download files from URL in CSV.
$ cloud-label-uploader download -i ./my_file_list.csv -o ./save -n "id" -l "label" -u "image_url"


# Chech downloaded files.
$ tree ./save

./save
├── cat
│   ├── 1.jpg
│   ├── 3.JPG
├── dog
│   ├── 2.jpg
└── human
    ├── 4.png
    └── 5.png

3 directories, 5 files
```


## list command

```bash
$ cloud-label-uploader help list
Create list file from --input dir images

Options:

  -h, --help                      display help information
  -i, --input                    *image dir path --input='/path/to/image_dir'
  -o, --output[=./output.csv]    *output CSV file path --output='./output.csv'
  -a, --all                       use all files
  -t, --type[=jpg,jpeg,png,gif]   comma separate file extensions --type='jpg,jpeg,png,gif'
  -f, --format[=csv]              set output format --format='[csv,sagemaker]'
  -p, --prefix                   *prefix for file path --prefix='gs://<your-bucket-name>'
```

```bash
# Create file list from given dir and save it to output CSV file.
$ cloud-label-uploader list -i ./save -o result.csv -p "gs://my-bucket/test-project"


# Check saved CSV file.
$ cat result.csv

gs://my-bucket/test-project/cat/1.jpg,cat
gs://my-bucket/test-project/cat/3.JPG,cat
gs://my-bucket/test-project/dog/2.jpg,dog
gs://my-bucket/test-project/human/4.png,human
gs://my-bucket/test-project/human/5.png,human
```


## upload command

```bash
$ cloud-label-uploader help upload
Upload files to Cloud Bucket(S3, GCS) from --input dir

Options:

  -h, --help                      display help information
  -i, --input                    *image dir path --input='/path/to/image_dir'
  -t, --type[=jpg,jpeg,png,gif]   comma separate file extensions --type='jpg,jpeg,png,gif'
  -a, --all                       use all files
  -l, --label                     label file for training (outputted CSV file) --label='/path/to/output.csv'
  -c, --provider                 *cloud provider name for the bucket --provider='[s3,gcs]'
  -b, --bucket                   *bucket name of S3/GCS --bucket='<your-bucket-name>'
  -p, --prefix                   *prefix for S3/GCS --prefix='foo/bar'
  -m, --parallel[=2]              parallel number (multiple upload) --parallel=2
```

```bash
# Before uploading, create GCS bucket
# $ gsutil mb gs://example-bucket

# Create file list from given dir and save it to output CSV file.
# $ export GOOGLE_API_GO_PRIVATEKEY=`cat /path/to/gcs.pem`
# $ export GOOGLE_API_GO_EMAIL=gcs@example.iam.gserviceaccount.com
$ export GOOGLE_APPLICATION_CREDENTIALS=/path/to/gcs.json
$ cloud-label-uploader upload -i ./save -b 'example-bucket' -t 'jpg,png' -p 'automl_model/20180401' -c 'gcs' -l './result.csv' -m 10

# upload files to gs://example-bucket/automl_model/20180401/ ...
```


## vott command

```bash
$ cloud-label-uploader help vott
Create object-detection list file from VoTT results

Options:

  -h, --help                    display help information
  -j, --json                   *VoTT json results dir path --image='/path/to/vott_json_dir'
  -o, --output[=./output.csv]  *output CSV file path --output='./output.csv'
  -p, --prefix[=gs://]         *prefix for file path --prefix='gs://<your-bucket-name>'
  -r, --recursive[=false]       read files in sub directories
```

```bash
# Create file list from given dir and save it to output CSV file.
$ cloud-label-uploader list -j ./vott/result -o result.csv -p "gs://my-bucket/test-project/"


# Check saved CSV file.
$ cat result.csv

UNASSIGNED,gs://my-bucket/test-project/cat/1.jpg,cat,0.17785499052004333,0.3945237235067437,0.28786057692307687,0.3945237235067437,0.28786057692307687,0.5815947133911368,0.17785499052004333,0.5815947133911368
UNASSIGNED,gs://my-bucket/test-project/cat/3.JPG,cat,0.7158391915641477,0.44460227272727265,0.8379421133567663,0.44460227272727265,0.8379421133567663,0.5285943675889327,0.7158391915641477,0.5285943675889327
UNASSIGNED,gs://my-bucket/test-project/dog/2.jpg,dog,0.6664113285482123,0.4608950407608696,0.7451203615926327,0.4608950407608696,0.7451203615926327,0.6013332201086957,0.6664113285482123,0.6013332201086957
UNASSIGNED,gs://my-bucket/test-project/human/4.png,human,0.730020144907909,0.4057476065751445,0.8625490926327194,0.4057476065751445,0.8625490926327194,0.5787572254335259,0.730020144907909,0.5787572254335259
UNASSIGNED,gs://my-bucket/test-project/human/5.png,human,,0.6799583559046587,0.4373871026011561,0.7823545842361863,0.4373871026011561,0.7823545842361863,0.5737953847543352,0.6799583559046587,0.5737953847543352
```
