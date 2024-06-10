<p align="center">
  <img src="doc/logo.png" alt="Bisquitt logo">
</p>

<p align="center">
  <a href="https://github.com/energostack/bisquitt/actions/workflows/bisquitt-tests.yaml"><img src="https://img.shields.io/github/workflow/status/energostack/bisquitt/bisquitt%20tests?style=flat-square" alt="Build status"></a>
  <a href="https://pkg.go.dev/github.com/energostack/bisquitt"><img src="https://pkg.go.dev/badge/github.com/energostack/bisquitt.svg" alt="Go Reference"></a>
  <a href="https://github.com/energostack/bisquitt#license"><img src="https://img.shields.io/github/license/energostack/bisquitt?style=flat-square" alt="License"></a>
</p>

Bisquitt is a transparent MQTT-SN gateway. It provides a simple, secure, and
standards-based way to connect resource-constrained IoT devices to MQTT
infrastructure.

Bisquitt supports most [MQTT-SN 1.2] features, allows secure communication using
[DTLS 1.2], and implements an [authentication extension](doc/auth.md) based on
[MQTT-SN 2.0 draft].

Besides the gateway, Bisquitt provides `bisquitt-sub` and `bisquitt-pub`
command-line MQTT-SN clients and can be used as a Go MQTT-SN library.

## Installation

The easiest way to install Bisquitt is to use its Docker image:

### Deprecation warning
Dockerhub registry images is now deprecated. Please use GitHub Container Registry (ghcr.io) instead.
Also, we renamed from **energomonitor** to **energostack**.
```console
$ docker pull ghcr.io/energostack/bisquitt
```

The image contains the gateway itself (`bisquitt`), which is started by default,
and the command-line MQTT-SN clients (`bisquitt-sub` and `bisquitt-pub`).

Alternatively, you can install the gateway and clients using `go install`:

```console
$ go install github.com/energostack/bisquitt/cmd/bisquitt@latest
$ go install github.com/energostack/bisquitt/cmd/bisquitt-pub@latest
$ go install github.com/energostack/bisquitt/cmd/bisquitt-sub@latest
```

Note that Bisquitt requires Go 1.16 or higher.

Let us know if you'd like to have additional installation formats available.

## Usage

The best way to start with Bisquitt is to run it together with [Mosquitto] (an
open-source MQTT server) using Docker Compose:

  1. Create a `mosquitto.conf` file with the following contents:

     ```
     listener 1883 0.0.0.0
     allow_anonymous true
     ```

  1. Create a `docker-compose.yml` file with the following contents:

     ```yaml
     version: "3.7"

     services:
       bisquitt:
         image: energostack/bisquitt
         environment:
           MQTT_HOST: mosquitto
           BISQUITT_USER: bisquitt
           BISQUITT_GROUP: bisquitt
         ports:
           - "1883:1883/udp"
         depends_on:
           - mosquitto

       mosquitto:
         image: eclipse-mosquitto
         ports:
           - "1883:1883"
         volumes:
           - ./mosquitto.conf:/mosquitto/config/mosquitto.conf
     ```

  1. Run the services:

     ```console
     $ docker-compose up
     ```

You can now play with Bisquitt using `bisquitt-sub` and `bisquitt-pub`
command-line MQTT-SN clients. Open a new terminal and open a shell in the
`bisquitt` service container:

```console
$ docker-compose exec bisquitt sh
```

Inside this shell, use `bisquitt-sub` to subscribe to a topic:

```console
# bisquitt-sub -t my-topic
```

Now open another shell in the `bisquitt` service container:

```console
$ docker-compose exec bisquitt sh
```

Inside the second shell, use `bisquitt-pub` to send a message to the topic
you've subscribed to:

```console
# bisquitt-pub -t my-topic -m message
```

You should see the following output in the first shell:

```console
my-topic: message
```

Voilà! You've just sent a message to an MQTT server via an MQTT-SN gateway from
one client and received it back in another one.

If you are interested in what's going on under the hood, add the `--debug`
option to any of the commands above.

For more information on usage, use the `--help` option on `bisquitt`,
`bisquitt-sub` or `bisquitt-pub`:

```console
# bisquitt --help
# bisquitt-pub --help
# bisquitt-sub --help
```

Once you are done playing with Bisquitt, shut down the services:

```console
$ docker-compose down
```

## Features

Bisquitt is a _transparent_ MQTT-SN gateway. This means that the gateway
maintains one connection to an MQTT server for every connected MQTT-SN client.
An MQTT-SN client can therefore be treated like any other MQTT client on the
MQTT server side (for purposes such as authentication, topics access management,
or monitoring).

The implementation is based on [MQTT-SN 1.2]. Its specification is a bit unclear
in some places, which required
[interpretation](doc/specification-interpretation.md).

### Supported MQTT-SN features

  * Connecting and disconnecting (`CONNECT`, `CONNACK`, `DISCONNECT`)
  * Topic registration (`REGISTER`, `REGACK`)
  * Publishing (`PUBLISH`, `PUBACK`, `PUBCOMP`, `PUBREC`, `PUBREL`)
  * Subscribing (`SUBSCRIBE`, `SUBACK`)
  * Last will (`WILLTOPICREQ`, `WILLTOPIC`, `WILLMSGREQ`, `WILLMSG`)
  * Keep alive (`PINGREQ`, `PINGRESP`)
  * QoS levels -1, 0, 1, 2
  * Sleeping clients

### Supported MQTT-SN extensions

  * Authentication (`AUTH`, based on the [MQTT-SN 2.0 draft] and described
    [separately](doc/auth.md))
  * [DTLS 1.2]

### Planned MQTT-SN features

  * Support for MQTT-SN 2.0

### Unsupported MQTT-SN features

  * Last will change (`WILLTOPICUPD`, `WILLTOPICRESP`, `WILLMSGUPD`,
    `WILLMSGRESP`)
  * Gateway advertisement and discovery (`ADVERTISE`, `SEARCHGW`, `GWINFO`)
  * Message forwarding

### Limitations

  * Bisquitt currently does not persist state between service restarts. This
    means it can't guarantee to preserve conversation state, last will messages,
    and topics in all situations. We plan to implement this soon.

  * DTLS certificates can't be reloaded without a service restart.

## Thanks

Bisquitt is inspired and partially based on [gnatt]. Thank you!

## License

Bisquitt is distributed under Eclipse Public License 2.0. See
[`LICENSE`](LICENSE) for more information.

Code in the `util/crypto` directory is distributed under the MIT license.
See [`LICENSE`](util/crypto/LICENSE) there for more information.

[MQTT-SN 1.2]: https://www.oasis-open.org/committees/download.php/66091/MQTT-SN_spec_v1.2.pdf
[MQTT-SN 2.0 draft]: https://www.oasis-open.org/committees/download.php/68568/mqtt-sn-v2.0-wd09.docx
[DTLS 1.2]: https://datatracker.ietf.org/doc/html/rfc6347
[Mosquitto]: https://mosquitto.org/
[gnatt]: https://github.com/alsm/gnatt
