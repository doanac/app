## The Docker voting app

### Initialize project

In this example, we will create a single service application that deploys a web page displaying a message.

Initialize the single file project using `docker-app init --single-file hello-world`. A single file application contains the three sections, `metadata` which corresponds to `metadata.yml`, `parameters` which corresponds to `parameters.yml` and `services` which corresponds to `docker-compose.yml`.

```bash
$ ls -l
-rw-r--r-- 1 README.md
$ docker-app init --single-file hello-world
$ ls -l
-rw-r--r-- 1 README.md
-rw-r--r-- 1 hello-world.dockerapp
$ cat hello-world.dockerapp
# This section contains your application metadata.
# Version of the application
version: 0.1.0
# Name of the application
name: hello-world
# A short description of the application
description:
# Namespace to use when pushing to a registry. This is typically your Hub username.
#namespace: myhubusername
# List of application maintainers with name and email for each
maintainers:
  - name: user
    email:

---
# This section contains the Compose file that describes your application services.
version: "3.6"
services: {}

---
# This section contains the default values for your application parameters.
```

Open `hello-world.dockerapp` with your favorite text editor.

### Edit metadata

Edit the `description`, `namespace` and `maintainers` fields in the metadata section.

### Add variables to the compose file

Add a service `hello` to the `services` section.

---

[hello-world.dockerapp](hello-world.dockerapp):
```yml
[...]
---
# This section contains the Compose file that describes your application services.
services:
  hello:
    image: hashicorp/http-echo
    command: ["-text", "${text}"]
    ports:
      - ${port}:5678

---
[...]
```

### Give variables their default value

In the parameters section, add every variables with the default value you want, e.g.:

---

[hello-world.dockerapp](hello-world.dockerapp):
```yml
[...]
---
# This section contains the default values for your application parameters.
port: 8080
text: Hello, World!
```

### Render

You can now render your application by running `docker-app render` or even personalize the rendered Compose file by running `docker-app render --set text="hello user!"`.

Create a `render` directory and redirect the output of `docker-app render ...` to `render/docker-compose.yml`.

```bash
$ mkdir -p render
$ docker-app render -s text="Hello user!" > render/docker-compose.yml
```

### Deploy

You directly deploy your application by running `docker-app deploy --set text="Hello user!"` or you can use the rendered version and run `docker stack deploy render/docker-compose.yml`.

`http://<ip_of_your_node>:8080` will display your message, e.g. http://127.0.0.1:8080 if you deployed locally.
