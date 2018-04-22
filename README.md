# Matcha

![Build](https://img.shields.io/badge/build-passing-brightgreen.svg?style=flat)
![License MIT](https://img.shields.io/badge/license-MIT-lightgray.svg?style=flat&maxAge=2592000)

```
 __  __    _  _____ ____ _   _    _    
|  \/  |  / \|_   _/ ___| | | |  / \   
| |\/| | / _ \ | || |   | |_| | / _ \  
| |  | |/ ___ \| || |___|  _  |/ ___ \ 
|_|  |_/_/   \_\_| \____|_| |_/_/   \_\

```

## Introduction

Matcha is a basic library of go language which include a lot of commonly used functions that make development simply.

## Components

- Double indexes byte buffer `ByteBuf`.
- Event driven and pipelined tcp framework include both server and client.
- Some frame decoder and frame encoder for pipelined tcp framework.
- Goroutine wrapper with status tracking support.
- Registry for cluster node election.
- Scheduler for task scheduling execution. Support fixed delay policy, fixed rate policy and [corntab](http://corntab.com) expression.
- Infra interface definitions and helper methods.
- BitSet interface and implementation.
- Set interface and implementation.
- APIs for configuration file loading. Support `JSON`,`YAML` and property file.
- Abstract logger for logging.

## Installation

Install:
```bash
$ go get -u github.com/mervinkid/matcha
```

## Development

Allspark use [dep](https://github.com/golang/dep) for dependency management. 
To install dependencies make sure to `dep` have been installed in user system then typing following command.

```bash
$ dep ensure
```

## User Case

- [JD.COM](https://www.jd.com)

## Contributing

1. Fork it.
2. Create your feature branch. (`$ git checkout feature/my-feature-branch`)
3. Commit your changes. (`$ git commit -am 'What feature I just added.'`)
4. Push to the branch. (`$ git push origin feature/my-feature-branch`)
5. Create a new Pull Request

## Author

[@Mervin](https://mervinz.me) 

## License

The MIT License (MIT). For detail see [LICENSE](LICENSE).