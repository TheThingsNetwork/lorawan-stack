---
title: "Application Packages"
description: ""
---

Application packages specify state machines running both on the end-device and the Application Server as well as signaling messages exchanged between the end-device's application layer and the Application Server.

<!--more-->

{{< cli-only >}}

## Listing the Available Packages

The application packages available for a given device `dev1` of application `app1` can be obtained as follows:

```bash
$ ttn-lw-cli applications packages list app1 dev1
```

This gives the result in the JSON format:

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

## Associations and Default Associations

The link between an application or a device and an application package is achieved through associations.

Associations link a device and a FPort to a specific application package. Since you may want to link all of the devices of an application without manually creating an association for each one of them, you may consider a default association, which links the application and a FPort to a specific application package.

All of the commands for managing associations and default associations are symmetric, and may be switched one with another. For the rest of this reference we will consider default associations.

## Creating and Updating an Association

In order to associate a given application package to a FPort of an application, you can use the `default-associations set` command:

```bash
$ ttn-lw-cli applications packages default-associations set app1 25 --package-name test-package
```

This will associate FPort `25` of application `app1` with the application package `test-package`, as shown by the command output:

```json
{
  "ids": {
    "application_ids": {
      "application_id": "app1"
    },
    "f_port": 25
  },
  "created_at": "2019-12-18T21:28:12.775879582Z",
  "updated_at": "2019-12-18T21:29:08.445380588Z",
  "package_name": "test-package"
}
```

Some application packages are stateful, and as such their state can be updated using the `data-*` parameters:

```bash
# Create a JSON formatted file containing package data
$ echo '{ "api_key": "AQEA8+q0v..." }' > package-data.json
# Update the association with the new package data
$ ttn-lw-cli applications packages default-associations set app1 25 --data-local-file package-data.json
```

This will update the association to use the given `api_key`:

```json
{
  "ids": {
    "application_ids": {
      "application_id": "app1"
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

## Listing the Associations

The package associations of a given device can be listed using the `default-associations list` command:

```bash
$ ttn-lw-cli applications packages default-associations list app1
```

Output:

```json
{
  "associations": [
    {
      "ids": {
        "application_ids": {
          "application_id": "app1"
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

## Retrieving an Association

The associations can be retrieved using the `default-associations get` command:

```bash
$ ttn-lw-cli applications packages default-associations get app1 25 --data
```

Output:

```json
{
  "ids": {
    "application_ids": {
      "application_id": "app1"
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

## Deleting an Association

The associations can be deleted using the `default-associations delete` command:

```bash
$ ttn-lw-cli applications packages associations delete app1 25
```
