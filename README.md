# home-ssm

A home lab implementation of AWS SSM Stored Parameter API.

> **_NOTE:_** Use at your own risk. There isn't PBAC nor robust validation and error checking. 
> The implementation is at best "mostly" consistent with the real thing.

## Overview

This project (kinda, sorta) implements AWS SSM Stored Parameter API. The goal is to enable the use of well known frameworks, such as Terraform, in the home lab setting.

For example, using home-ssm, it's possible to: 
* put-parameter using AWS CLI 
* get-parameter using Terraform

AWS CLI

```shell

#!/bin/bash

REGION=us-east-1
PROFILE=home-ssm
ENDPOINT=http://localhost:9080/ssm

aws ssm \
    --region ${REGION} \
    --profile ${PROFILE} \
    --endpoint "${ENDPOINT}" \
    --output json \
    put-parameter \
    --name /home/mydb/password \
    --type SecureString \
    --value some-long-password
```

Terraform:

```terraform

provider "aws" {
  profile = "home-ssm"
  region  = "us-east-1"

  endpoints {
    ssm = "http://localhost:9080/ssm"
  }
}

data "aws_ssm_parameter" "database_password" {
   name = "/home/mydb/password"
   with_decryption = true
}
```
## Configuration

### home-ssm-config.yaml

Region and one value for each stanza, credentials and keys, is required. The application uses the first Key as the default key.

Region, AccessKey, SecretKey, Username, Alias, KeyId are arbitrary though you'll probably want consistancy with ~/.aws/{config,credentials} files. 

The credentials are used for authenticating the v4sig headers; no PBAC is enforced.

The Key data value must be base64 encoded; see comments below. Parameters of type SecureString are encrypted and decrypted based on the KeyId argument. Config ID and Alias values are used for lookup of the KeyId argument. 

```yaml
region: us-east-1

credentials:
  - accessKey: "my-access"
    secretKey: "really-long-key"
    username: "John.Doe"

# AES-256 uses a 32-byte (256-bit) key
# openssl rand -base64 32
keys:
  - alias: aws/ssm
    id: 844c1364-08b8-11f0-aeb7-33cf4b255e16
    key: DkVsBYNRbORxQ6vtjUCex54YdfYfxd3c5PcP/ZruwUs=
  - alias: home-ssm
    id:  d0c49d70-4fae-4a20-84f0-d03fb6d670cb
    key: rvl7SbrNObB5MMQDUUAoInJXpyCA3QDqELyuwa2G48M=

```
## Execution

```shell

./home-ssm -help
Usage of ./home-ssm:
  -config string
    	Path to the home-ssm config file. (default ".home-ssm-config.yaml")
  -db-path string
    	Path to badger database folder. (default ".home-ssm-db")
```
