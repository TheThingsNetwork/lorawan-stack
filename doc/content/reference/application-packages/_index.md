---
title: "Application Packages"
description: ""
weight: 4
---

This is the reference to work with Application Server application packages.

<!--more-->

Application packages specify state machines running both on the end-device and the Application Server as well as signaling messages exchanged between the end-device's application layer and the Application Server.

{{< cli-only >}}

## Listing the Available Packages

The application packages available for a given device `dev1` of application `app1` can be obtained as follows:

```bash
$ ttn-lw-cli applications packages list app1 dev1
```

This gives the result in the JSON format:

<details><summary>Show output</summary>
```json
{
  "packages": [
    {
      "name": "test-package",
      "default_f_port": 20
    }
  ]
}
```
</details>

## Creating and Updating an Association

In order to associate a given application package to a FPort of a device, you can use the `association set` command:

```bash
$ ttn-lw-cli applications packages associations set app1 dev1 25 --package-name test-package
```

This will associate FPort `25` of device `dev1` of application `app1` with the application package `test-package`:

<details><summary>Show output</summary>
```json
{
  "ids": {
    "end_device_ids": {
      "device_id": "dev1",
      "application_ids": {
        "application_id": "app1"
      }
    },
    "f_port": 25
  },
  "created_at": "2019-12-18T21:28:12.775879582Z",
  "updated_at": "2019-12-18T21:29:08.445380588Z",
  "package_name": "test-package"
}
```
</details>

Some application packages are stateful, and as such their state can be updated using the `data-*` parameters:

```bash
# Create a JSON formatted file containing package data
$ echo '{ "api_key": "AQEA8+q0v..." }' > package-data.json
# Update the association with the new package data
$ ttn-lw-cli applications packages associations set app1 dev1 25 --data-local-file package-data.json
```

This will update the association to use the given `api_key`:

<details><summary>Show output</summary>
```json
{
  "ids": {
    "end_device_ids": {
      "device_id": "dev1",
      "application_ids": {
        "application_id": "app1"
      }
    },
    "f_port": 25
  },
  "created_at": "2019-12-18T21:28:12.775879582Z",
  "updated_at": "2019-12-18T21:37:16.470742803Z",
  "package_name": "test-package",
  "data": {
      "api_key": "AQEA8+q0v..."
    }
}
```
</details>

## Listing the Associations

The package associations of a given device can be listed using the `association list` command:

```bash
$ ttn-lw-cli applications packages associations list app1 dev1
```

<details><summary>Show output</summary>
```json
{
  "associations": [
    {
      "ids": {
        "end_device_ids": {
          "device_id": "dev1",
          "application_ids": {
            "application_id": "app1"
          }
        },
        "f_port": 25
      },
      "created_at": "2019-12-18T21:28:12.775879582Z",
      "updated_at": "2019-12-18T21:29:08.445380588Z",
      "package_name": "test-package"
    }
  ]
}
```
</details>

## Retrieving an Association

The associations can be retrieved using the `association get` command:

```bash
$ ttn-lw-cli applications packages associations get app1 dev1 25 --data
```

<details><summary>Show output</summary>
```json
{
  "ids": {
    "end_device_ids": {
      "device_id": "dev1",
      "application_ids": {
        "application_id": "app1"
      }
    },
    "f_port": 25
  },
  "created_at": "2019-12-18T21:28:12.775879582Z",
  "updated_at": "2019-12-18T21:37:16.470742803Z",
  "package_name": "test-package",
  "data": {
      "api_key": "AQEA8+q0v..."
    }
}
```
</details>

## Deleting an Association

The associations can be deleted using the `association delete` command:

```bash
$ ttn-lw-cli applications packages associations delete app1 dev1 25
```

