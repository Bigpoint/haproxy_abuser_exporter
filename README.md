# haproxy_abuser_exporter

## TODO

Usage:
```
  -endpoint string
        Endpoint that is exposed for prometheus (default "/metrics")
  -gpc string
        The HAproxy GRC that contains the connections (default "gpc0")
  -instance string
        If specified will enhance the metrics with a 'instance' field
  -port int
        Port to listen on (default 9322)
  -reqRate string
        The HAproxy register that contains the connections (default "http_req_rate(10000)")
```

### Author & License

 Author: Hendrik Jonas Meyer <h.meyer@bigpoint.net>
 License: Apache-2.0
 Copyright: Copyright 2017 Bigpoint GmbH
