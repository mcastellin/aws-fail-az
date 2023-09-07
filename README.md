# AWS Fail AZ

**AWS Fail AZ** is a command-line program to simulate AZ (Availability Zone) failures on AWS resources.


## Getting Started

`aws-fail-az` is a self-contained Go binary.

## Installation

**Linux / Windows / MacOs**

Download one of the [pre-built binaries][releases].

Example of **aws-fail-az** installation on a `debian:12` Docker container:

```Dockerfile
FROM debian:12

ARG VERSION=0.0.4
RUN apt-get update \
    && apt-get install -qqy curl \
    && rm -rf /var/lib/apt/lists \
    && curl -sSL -o /opt/aws-fail-az-$VERSION.tar.gz \
        https://github.com/mcastellin/aws-fail-az/releases/download/$VERSION/aws-fail-az_Linux_x86_64.tar.gz \
    && tar xf /opt/aws-fail-az-$VERSION.tar.gz -C /opt \
    && mv /opt/aws-fail-az /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/aws-fail-az"]
```

To verify the binary installation was successful, run `aws-fail-az version`.

## Usage

**aws-fail-az** can simulate AZ (Availability Zone) failure on a number of AWS resources. (To see a list of all available resources see the [Failure Configuration](#failure-configuration) section).

### Fail AZs for configured target resources

Once you have created your configuration file, you can apply AZ failure to the infrastructure using the `fail` command:

```shell
export AWS_REGION=us-east-1
export AWS_PROFILE=default

aws-fail-az fail configuration.json
```

Alternatively, the configuration file can be supplied via **stdin** if you render it dynamically:

```shell
export AWS_REGION=us-east-1
export AWS_PROFILE=default

aws-fail-az fail --stdin <<EOF
{
  "azs": [
    "us-east-1b"
  ],
  "targets": [
    {
      "type": "auto-scaling-group",
      "filter": "name=<ASG_NAME>"
    }
  ]
}
EOF

```

### Recover AZs failure from state

**aws-fail-az** will automatically save the original state of AWS resources in DynamoDB before simulating AZs failure.

To restore the state of your resources from the states table use the `recover` command:

```shell
export AWS_REGION=us-east-1
export AWS_PROFILE=default

aws-fail-az recover
```

> No configuration file is needed to restore original state.


## Failure Configuration

To simulate AZ failure in your own AWS account, you need select affected resources using a *JSON* configuration file and feed it to **aws-fail-az** either via command-line argument or via STDIN.

The structure of the configuration file is as follows:

```json
{
  "azs": [
    "us-east-1a",
    "us-east-1b",
    ...
  ],
  "targets": [
    {
      "type": "ecs-service",
      "filter": "cluster=<CLUSTER_NAME>;service=<SERVICE_NAME>",
      "tags": [
        {
          "Name": "Environment",
          "Value": "<ENVIRONMENT_NAME>"
        },
        {
          "Name": "Application",
          "Value": "<APPLICATION_NAME>"
        },
        ...
      ]
    }
  ]
}
```

#### `azs`: list[string]

Use the `azs` field to specify the list of availability zones to fail.

Availability zones are typically identified in AWS by the *region-name* followed by a *letter* (i.e. *us-east-1a, us-east-1b,* ...).

#### `targets`: list[object]

The `targets` field contains a list of objects used by **aws-fail-az** to select AWS resources to attack.

Available fields are the following:

**type** (Required)

The type of resources to select. For a full list of available types see the [Available Resources](#available-resources) section.

**filter** (Optional)

Select resources using a filter expression.

The expression syntax is a list of resource attributes separated by a semi-colon `;` character. Available attributes for filtering vary depending on the type of resource being selected.

> Only one selection strategy between `filter` and `tags` is allowed for every target selector.

**tags** (Optional)

Select resources using tags.

Using the `tags` attribute will select all resource of the specified *type* where all tags are associated.

> Only one selection strategy between `filter` and `tags` is allowed for every target selector.

### Available Resources

| Resources | Available Filters |
|---------|-------------|
| ecs-service           | cluster, service, tags |
| auto-scaling-group    | name, tags |
| elbv2-load-balancer   | name, tags |

### ECS Services

Select ECS service by cluster and service name:

```json
{
  "azs": [
    "us-east-1b"
  ],
  "targets": [
    {
      "type": "ecs-service",
      "filter": "cluster=<CLUSTER_NAME>;service=<SERVICE_NAME>"
    },
    ...
  ]
}
```

Select ECS services by tags:

```json
{
  "azs": [
    "us-east-1b"
  ],
  "targets": [
    {
      "type": "ecs-service",
      "tags": [
        {
          "Name": "Environment",
          "Value": "<ENVIRONMENT_NAME>"
        },
        {
          "Name": "Application",
          "Value": "<APPLICATION_NAME>"
        }
      ]
    },
    ...
  ]
}
```

### Auto Scaling Groups

Select Auto Scaling Groups by name:

```json
{
  "azs": [
    "us-east-1b"
  ],
  "targets": [
    {
      "type": "auto-scaling-group",
      "filter": "name=<ASG_NAME>"
    },
    ...
  ]
}
```

### Elastic Load Balancers

Select Elastic Load Balancers by name:

```json
{
  "azs": [
    "us-east-1b"
  ],
  "targets": [
    {
      "type": "elbv2-load-balancer",
      "filter": "name=<LB_NAME>"
    },
    ...
  ]
}
```

[releases]: https://github.com/mcastellin/aws-fail-az/releases/
