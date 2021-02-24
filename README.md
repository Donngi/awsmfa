# AWS MFA resolver in Go
![DEMO](https://github.com/Jimon-s/awsmfa/blob/images/demo.gif)

The 'awsmfa' is a simple cli tool for AWS MFA.
Both AWS STS GetSessionToken and AssumeRole API are supported.

From my exeperience, MFA and STS operation are sometimes a point of a trouble shooting. Therefore, awsmfa is designed to show you request parameters visually.

## Feature
- Visual design
- Available all region
- Easy to set up. Almost all keys are same as aws-cli v2's default.
- Available various cli options like aws-cli v2 (such as --duration-seconds, --serial-number)
- Set default params via awsmfa's configuration file (`${HOME}/.awsmfa/configuration`)
- Support both GetSessionToken and AssumeRole API

## Installing
```
clone this repo

$ go install
```

## Quick start
First, you should set profile in your shared credentials and config file (By default, it's placed `${HOME}/.aws/credentials` and `${HOME}/.aws/config`).
It will be used in executing sts api to obtain temporary credentials.

No worries!
You can easily get templates by using helper options.
- `awsmfa --generate-credentials-skeleton get-session-token`
- `awsmfa --generate-config-skeleton get-session-token`
- `awsmfa --generate-credentials-skeleton assume-role`
- `awsmfa --generate-config-skeleton assume-role`

example: credentials (get-session-token)
```
[sample-before-mfa]
aws_access_key_id     = YOUR_ACCESS_KEY_ID_HERE!!!
aws_secret_access_key = YOUR_SECRET_ACCESS_KEY_HERE!!!
```

example: config (get-session-token)
```
[profile sample-before-mfa]
region     = REGION_TO_CONNECT_IN_EXECUTING_STS_GET_SESSION_TOKEN # Such as ap-northeast-1, us-east-1
output     = json
mfa_serial = YOUR_MFA_SERIAL_HERE!!! # Such as arn:aws:iam::XXXXXXXXXXX:mfa/YYYY

[profile sample]
region = REGION_TO_CONNECT_AFTER_MFA # Such as ap-northeast-1, us-east-1
output = json
```

Then, you simply exec these command.

```
$ awsmfa --profile sample
```

The awsmfa automatically exec sts api and add/update shared credentials.

```
Automatically add new credentials in shared credentials file.

[sample]
aws_access_key_id     = NEW_ACCESSKEY_ID
aws_secret_access_key = NEW_SECRET_ACCESS_KEY
aws_session_token     = NEW_SESSION_TOKEN
expiration            = 2999-11-23T14:15:16Z
```

## Supported API
AWS provides us two types of API to obtain temporary security credentials for cli access.
[AWS: Requesting temporary security credentials](https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_temp_request.html)

You can select the api by using `--mode get-session-token` or `--mode assume-role` (by default, get-session-token is activated).

The available APIs are different according to your environment, please check your IAM setting.

The image of the operation is below.

### GetSessionToken
```
$ awsmfa --profile sample
or
$ awsmfa --profile sample --mode get-session-token
```

![GetSessionToken](https://github.com/Jimon-s/awsmfa/blob/images/get-session-token.jpg)

### AssumeRole
```
$ awsmfa --profile sample --mode assume-role
```

![AssumeRole](https://github.com/Jimon-s/awsmfa/blob/images/assume-role.jpg)

## Priority of params
The awsmfa is designed to match the priority of params with aws cli's default order.

Basically, each params give priority according to the order below.

1. CLI option
2. environment variable
3. shared credentials file (`${HOME}/.aws/credentials`)
4. shared config file (`${HOME}/.aws/config`)
5. awsmfa's configuration file (`${HOME}/.awsmfa/configuration`)
6. awsmfa's build in default value

## License
MIT

