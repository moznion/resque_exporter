resque_exporter
==

An exporter of [Prometheus](https://prometheus.io) for [resque](https://github.com/resque/resque)'s queue status.

Usage
--

```
Options:

  -h, --help            display help
  -v, --version         display version and revision
  -p, --port[=5555]     set port number
  -c, --config          set path to config file
```

e.g.

```
$ ./resque_exporter --config /path/to/config.yml
```

Sample Output
--

```
# HELP jobs_in_queue Number of remained jobs in queue
# TYPE jobs_in_queue gauge
jobs_in_queue{queue_name="image_converting"} 0
jobs_in_queue{queue_name="log_compression"} 0
```

Mechanism
--

This exporter accesses to redis to aggregate queue status.

1. Collect name of queues via `<namespace>:queues` entry (by using [SMEMBERS](http://redis.io/commands/smembers))
1. Get number of remained jobs for each queue via `<namespace>:queue:<queue_name>` entry (by using [LLEN](http://redis.io/commands/llen))

Note
--

This exporter also supports resque compatible job-queue engine (e.g. [jesque](https://github.com/gresrun/jesque)).

[For developers] How to build to release
--

Execute `make build VERSION=${version}` on __Docker available__ environment. Built binaries will be on `bin` directory.

License
--

```
The MIT License (MIT)
Copyright © 2016 moznion, http://moznion.net/ <moznion@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the “Software”), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```

