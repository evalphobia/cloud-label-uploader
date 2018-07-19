csv-file-downloader
----

[![GoDoc][1]][2] [![License: MIT][3]][4] [![Release][5]][6] [![Build Status][7]][8] [![Go Report Card][13]][14] [![Code Climate][19]][20] [![BCH compliance][21]][22]

[1]: https://godoc.org/github.com/evalphobia/csv-file-downloader?status.svg
[2]: https://godoc.org/github.com/evalphobia/csv-file-downloader
[3]: https://img.shields.io/badge/License-MIT-blue.svg
[4]: LICENSE.md
[5]: https://img.shields.io/github/release/evalphobia/csv-file-downloader.svg
[6]: https://github.com/evalphobia/csv-file-downloader/releases/latest
[7]: https://travis-ci.org/evalphobia/csv-file-downloader.svg?branch=master
[8]: https://travis-ci.org/evalphobia/csv-file-downloader
[9]: https://coveralls.io/repos/evalphobia/csv-file-downloader/badge.svg?branch=master&service=github
[10]: https://coveralls.io/github/evalphobia/csv-file-downloader?branch=master
[11]: https://codecov.io/github/evalphobia/csv-file-downloader/coverage.svg?branch=master
[12]: https://codecov.io/github/evalphobia/csv-file-downloader?branch=master
[13]: https://goreportcard.com/badge/github.com/evalphobia/csv-file-downloader
[14]: https://goreportcard.com/report/github.com/evalphobia/csv-file-downloader
[15]: https://img.shields.io/github/downloads/evalphobia/csv-file-downloader/total.svg?maxAge=1800
[16]: https://github.com/evalphobia/csv-file-downloader/releases
[17]: https://img.shields.io/github/stars/evalphobia/csv-file-downloader.svg
[18]: https://github.com/evalphobia/csv-file-downloader/stargazers
[19]: https://codeclimate.com/github/evalphobia/csv-file-downloader/badges/gpa.svg
[20]: https://codeclimate.com/github/evalphobia/csv-file-downloader
[21]: https://bettercodehub.com/edge/badge/evalphobia/csv-file-downloader?branch=master
[22]: https://bettercodehub.com/

`csv-file-downloader` download files from url in CSV.
And create CSV file with label and path for Google Cloud AutoML.

# Installation

Install csv-file-downloader by command below,

```bash
$ go get github.com/evalphobia/csv-file-downloader
```

# Usage

## root command

```bash
$ csv-file-downloader
Commands:

  help       show help
  download   Download files from --file csv
  list       Create csv list file from --output dir
```

## download command

```bash
$ csv-file-downloader help download
Download files from --file csv

Options:

  -h, --help           display help information
  -f, --file          *download list file --file='/path/to/dir/input.csv'
  -n, --name          *column name for filename --name='name'
  -l, --label         *column name for label --label='group'
  -u, --url           *column name for URL --url='path'
  -p, --parallel[=2]   parallel number --parallel=2
  -o, --out            outout dir --out='/path/to/dir/out'
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
$ csv-file-downloader download -f ./my_file_list.csv -o ./save -n "id" -l "label" -u "image_url"


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

## upload command

```bash
$ csv-file-downloader help uploader
Upload files to GCS from --input dir

Options:

  -h, --help                      display help information
  -i, --input                    *image dir path --input='/path/to/image_dir'
  -t, --type[=jpg,jpeg,png,gif]   comma separate file extensions --type='jpg,jpeg,png,gif'
  -a, --all                       use all files
  -b, --bucket                   *bucket name of GCS --bucket='<your-bucket-name>'
  -d, --prefix                   *prefix for GCS --prefix='foo/bar'
  -p, --parallel[=2]              parallel number --parallel=2
```

```bash
# Create file list from given dir and save it to output CSV file.
$ export GOOGLE_API_GO_PRIVATEKEY=`cat /path/to/gcs.pem`
$ export GOOGLE_API_GO_EMAIL=gcs@example.iam.gserviceaccount.com
$ csv-file-downloader upload -i ./save -b 'example-vcm' --prefix 'automl_model/20180401'

# upload files to gs://example-vcm/automl_model/20180401/ ...
```

## list command

```bash
$ csv-file-downloader help list
Create csv list file from --output dir

Options:

  -h, --help                      display help information
  -i, --input                    *image dir path --input='/path/to/image_dir'
  -o, --output[=./output.csv]    *output TSV file path --output='./output.csv'
  -t, --type[=jpg,jpeg,png,gif]   comma separate file extensions --type='jpg,jpeg,png,gif'
  -p, --prefix                    prefix for file path --prefix='gs://<your-bucket-name>'
```

```bash
# Create file list from given dir and save it to output CSV file.
$ csv-file-downloader list -i ./save -o result.csv -p "gs://my-bucket/test-project"


# Check saved CSV file.
$ cat result.csv

gs://my-bucket/test-project/cat/1.jpg,cat
gs://my-bucket/test-project/cat/3.JPG,cat
gs://my-bucket/test-project/dog/2.jpg,dog
gs://my-bucket/test-project/human/4.png,human
gs://my-bucket/test-project/human/5.png,human
```
