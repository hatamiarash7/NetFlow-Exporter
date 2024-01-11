# NetFlow Exporter

[![Release][release_badge]][release_link]
[![License][badge_license]][link_license]
[![Image size][badge_size_latest]][link_docker_hub]

It's a simple Prometheus exporter for NetFlow. Supported versions:

- NetFlow v1
- NetFlow v5
- NetFlow v9

## HowTo

You can use binary or Docker image.

```bash
docker run -d \
    -p 9438:9438 \
    hatamiarash7/netflow-exporter:v1.0.0
```

Or

```bash
./netflow-exporter
```

## Configuration

There is multiple runtime flags to configure the exporter:

| Flag              | Description                               | Default    |
| ----------------- | ----------------------------------------- | ---------- |
| `-log-level`      | Log level                                 | `info`     |
| `-log-format`     | Log format                                | `text`     |
| `-listen-address` | Network address to accept NetFlow packets | `:2055`    |
| `-metric-address` | Network address to expose metrics         | `:9438`    |
| `-metrics-path`   | Path under which to expose metrics        | `/metrics` |
| `-include`        | Include filter for NetFlow packets        | `Count$`   |
| `-exclude`        | Exclude filter for NetFlow packets        | `Time`     |
| `-sample-expire`  | How long a sample is valid for            | `60s`      |

---

## Support 💛

[![Donate with Bitcoin](https://img.shields.io/badge/Bitcoin-bc1qmmh6vt366yzjt3grjxjjqynrrxs3frun8gnxrz-orange)](https://donatebadges.ir/donate/Bitcoin/bc1qmmh6vt366yzjt3grjxjjqynrrxs3frun8gnxrz) [![Donate with Ethereum](https://img.shields.io/badge/Ethereum-0x0831bD72Ea8904B38Be9D6185Da2f930d6078094-blueviolet)](https://donatebadges.ir/donate/Ethereum/0x0831bD72Ea8904B38Be9D6185Da2f930d6078094)

<div><a href="https://payping.ir/@hatamiarash7"><img src="https://cdn.payping.ir/statics/Payping-logo/Trust/blue.svg" height="128" width="128"></a></div>

## Contributing 🤝

Don't be shy and reach out to us if you want to contribute 😉

1. Fork it!
2. Create your feature branch: `git checkout -b my-new-feature`
3. Commit your changes: `git commit -am 'Add some feature'`
4. Push to the branch: `git push origin my-new-feature`
5. Submit a pull request

[release_badge]: https://github.com/hatamiarash7/netflow-exporter/actions/workflows/release.yml/badge.svg
[release_link]: https://github.com/hatamiarash7/netflow-exporter/actions/workflows/docker.yaml
[link_license]: https://github.com/hatamiarash7/netflow-exporter/blob/master/LICENSE
[badge_license]: https://img.shields.io/github/license/hatamiarash7/netflow-exporter.svg?longCache=true
[badge_size_latest]: https://img.shields.io/docker/image-size/hatamiarash7/netflow-exporter/latest?maxAge=30
[link_docker_hub]: https://hub.docker.com/r/hatamiarash7/netflow-exporter/
