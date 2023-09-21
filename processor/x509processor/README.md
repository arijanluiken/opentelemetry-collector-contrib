# Resource Processor

<!-- status autogenerated section -->
| Status        |           |
| ------------- |-----------|
| Stability     | [alpha]: logs   |
| Distributions | [contrib] |
| Issues        | [![Open issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aopen%20label%3Aprocessor%2Fx509%20&label=open&color=orange&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aopen+is%3Aissue+label%3Aprocessor%2Fx509) [![Closed issues](https://img.shields.io/github/issues-search/open-telemetry/opentelemetry-collector-contrib?query=is%3Aissue%20is%3Aclosed%20label%3Aprocessor%2Fx509%20&label=closed&color=blue&logo=opentelemetry)](https://github.com/open-telemetry/opentelemetry-collector-contrib/issues?q=is%3Aclosed+is%3Aissue+label%3Aprocessor%2Fx509) |
| [Code Owners](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/CONTRIBUTING.md#becoming-a-code-owner)    | [@arijanluiken](https://www.github.com/arijanluiken) |

[alpha]: https://github.com/open-telemetry/opentelemetry-collector#alpha
[contrib]: https://github.com/open-telemetry/opentelemetry-collector-releases/tree/main/distributions/otelcol-contrib
<!-- end autogenerated section -->

POC code. Not optimized and quite dirty.

Only supports x509 rsa in the current state.

Configuration:

```
x509:
  keyfile: "key.pem"
  encrypt: 
  - attribute_to_encrypt1
  sign:
  - attribute_to_sign1
```

Example:

```
opentelemetry-collector-contrib/processor/x509processor on  x509 via 🐹 v1.21.1 
❯ go test -v
=== RUN   TestLoadConfig
=== PAUSE TestLoadConfig
=== RUN   TestCreateDefaultConfig
--- PASS: TestCreateDefaultConfig (0.00s)
=== RUN   TestCreateProcessor
--- PASS: TestCreateProcessor (0.00s)
=== RUN   TestX509Processor
=== RUN   TestX509Processor/i3_and_c3_data
< map[c2_attribute:i don't need encryption or signing c3_attribute:will be encrypted i3_attribute:will be signed]
> map[_x509_enc_c3_attribute:nXaTJ6xM06BMgIPSwd4x7m0APD0uzWrNb4OjW/wHgLkN+mgF5K971PrHi9L4ELDR/cMzvJSdjdBk8xYEselYxi8nqbEhQZeCLsrE3WTwQ6Yhz23pecLro9g3jUszAa433wqQmce9wqXhz9qsrkbv1ZLPga2ez3A/O4hi2D6GPdpROTxo14tC5NcXizdnyVJPCD+P/crc1ONeyogjPf/OFQYXo8vhDgkp1rSnqFHR9n+QQNkUjRqlY6kyvZEqcobL21j/yJW+Ch2DNY4Kt0Lwz/x7fYEhvSgmJOgBM76s3RFL7ZB/euW4zpG9pnmgjvSgINwFDemhSneMp0pghHgw9g== _x509_key:68240b6303e7b2e2494179696d73abf1382269110b277743bdc5c9620ccb0147 _x509_sig_i3_attribute:RIOgHGAzH53mGaITJGLqmrCPjxDmQ3y13wKVV8/3GzZkbv0xJ1bl2l1U3XtcaOAMrBtf3wA+q0b6l8MTTkKS3JN/ZtIFNHsidIJkXEBHInzK8RnqIidCq1D4jPxbJIeeJP0+dK8bUjyhxFMMKNhZo4oD86FaE1PLxU5p0cJxX7K4J7u/0IY8WxezutrrcsLiI7Yooh3Phvbkrx4dPb3pR9ahUGBYjzb6aY0qO+IR8EG+5c8h1k37LALoxUwDSgEbImKChjuxojIGcPr/5yifQLloa5WL0J3EaUCRJhicfzmxlTjF8qRvJGhtTbDuvVHdP79RxmApcHNKu27i9bQuZw== c2_attribute:i don't need encryption or signing i3_attribute:will be signed]
--- PASS: TestX509Processor (0.00s)
    --- PASS: TestX509Processor/i3_and_c3_data (0.00s)
=== CONT  TestLoadConfig
=== RUN   TestLoadConfig/x509
--- PASS: TestLoadConfig (0.00s)
    --- PASS: TestLoadConfig/x509 (0.00s)
PASS
ok      github.com/open-telemetry/opentelemetry-collector-contrib/processor/x509processor       0.012s
```