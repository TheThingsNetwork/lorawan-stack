---
title: "Upgrading"
description: ""
---

Follow these instructions to upgrade an existing {{% tts %}} instance.

Replace the tag in your **docker-compose.yml** file with the newer version of {{% tts %}}:

```yaml
# file: docker-compose.yml
services:
  # ...
  stack:
    image: 'thethingsnetwork/lorawan-stack:<the tag>'
  # ...

```

Pull the new images:

```bash
$ docker-compose pull
Pulling cockroach ... done
Pulling redis     ... done
Pulling stack     ... done
```

Shut down the stack:

```bash
$ docker-compose down
Stopping root_stack_1     ... done
Stopping root_cockroach_1 ... done
Stopping root_redis_1     ... done
Removing root_stack_1     ... done
Removing root_cockroach_1 ... done
Removing root_redis_1     ... done
Removing network root_default
```

Run migrations:

```bash
$ docker-compose run --rm stack is-db migrate
Creating network "root_default" with the default driver
Creating root_cockroach_1 ... done
Creating root_redis_1     ... done
  INFO Connecting to Identity Server database...
  INFO Detected database CockroachDB CCL v19.2.6 (x86_64-unknown-linux-gnu, built 2020/04/06 18:05:31, go1.12.12)
  INFO Migrating tables...                     
  INFO Successfully migrated  
```

Start stack:

```bash
$ docker-compose up -d
root_cockroach_1 is up-to-date
root_redis_1 is up-to-date
Creating root_stack_1 ... done
```