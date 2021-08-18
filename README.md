# herofig
The **Hero**ku Con**fig** tool.

## Prerequisites
To run herofig you need to install and log into the [Heroku CLI](https://devcenter.heroku.com/articles/heroku-cli).

# Building
```shell
go build -o herofig
```

## Usage
Similar to the Heroku CLI, the application name must be specified with `-a` or `--app` when it cannot be inferred
from the current working directory. Note that these flags must always be passed as the first argument to `herofig` (for now).

### Pulling the entire application config
```shell
herofig pull

# Into a file
herofig pull my-app.env
```

### Getting the value of a single config variable
```shell
herofig get AWS_S3_BUCKET
```

### Pushing a config file
```shell
herofig push local.env
```

### Pushing only new values from a config file
```shell
herofig push:new local.env
```

### Setting the value of config variables
```shell
herofig set AWS_S3_REGION=eu-north-1 AWS_S3_BUCKET=bucket
```

### Searching for config variables (fuzzy matched)
```shell
herofig search AWS
```

### Comparing configurations
```shell
herofig hash
```