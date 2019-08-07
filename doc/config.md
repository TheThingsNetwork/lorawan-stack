# Configuration

The Things Stack binary can be configured by passing parameters.

## Default config

The Things Stack can be used for local testing purposes without any additional configuration. To test and use the components to the fullest, you can use the `--help` flag to show available parameters, and refer to execution logs for recommendations.

## Sources

### Runtime flags

Example of runtime flags:

```bash
$ ttn-lw-stack start \
  --cluster.keys db114f80fd0ebf2a7b69db7e5a56fce248e53600e256b3a85d5b0ab844bc1aa8 \
  --http.cookie.hash-key 40847E55ED0CB34B3D491DC557326BF875FCE34EE0C8F50194E1BB3488055FA96D5CC4F3CF6C30C5F4922D8CEB4F72A1FE61317E1A7BC88619617AD6CEA983B3 \
  --http.cookie.block-key 38E31BCAD8CFC067ABC9F2988967E387E15DF9ADDA14E63F446ED955EEEA4637
```

### Configuration files

You can specify a YAML configuration file with the `-c` flag. `~/.ttn-lw-<binary>.yml` is used by default, with the binary type being `stack` or `cli`.

Example of a YAML configuration file:

```yaml
cluster:
  keys:
  - db114f80fd0ebf2a7b69db7e5a56fce248e53600e256b3a85d5b0ab844bc1aa8

http:
  cookie:
    hash-key: 40847E55ED0CB34B3D491DC557326BF875FCE34EE0C8F50194E1BB3488055FA96D5CC4F3CF6C30C5F4922D8CEB4F72A1FE61317E1A7BC88619617AD6CEA983B3
    block-key: 38E31BCAD8CFC067ABC9F2988967E387E15DF9ADDA14E63F446ED955EEEA4637
```

You can then start a binary and pass this config:

```bash
$ ./ttn-lw-stack -c config.yml
```

### Environment variables

Environment variables are the uppercased flags, with any separators (`.` or `-`) replaced by underscores (`_`). Environment variables are prefixed with `TTN_LW_`

Example of environment variable as options:

```bash
$ export TTN_LW_COOKIE_HASH_KEY=40847E55ED0CB34B3D491DC557326BF875FCE34EE0C8F50194E1BB3488055FA96D5CC4F3CF6C30C5F4922D8CEB4F72A1FE61317E1A7BC88619617AD6CEA983B3
$ export TTN_LW_LOG_LEVEL=debug
$ ttn-lw-identity-server start
```
