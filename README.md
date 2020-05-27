# majin
Realistic http load tests with synthetic timeseries simulations based on vegeta

Build:
```
go build .
```

Use:
```
./majin attack --help
Generate a workload simulation and attack according to the simulation parameters given

Usage:
  majin attack [flags]

Flags:
      --base-load float         The base queries per second to use for attack (default 1)
      --body string             The body to send in request
      --burst-duration float    how long  bursts should be (default 30)
      --burst-frequency float   how frequent  bursts should be
      --burst-index float        burst multiplier
      --duration float          Duration to run the attack for. 0 means run the attack forever
      --header headers          The HTTP headers to use
  -h, --help                    help for attack
      --method string           The http method to use (default "GET")
      --noise float             how much noise is in the workload
      --period float            how long each period is - as int of seconds (default 60)
      --random                  Generate random workloads? Use the other parameters as max random values.
      --target string           The target url to attack (default "http://localhost:8080")
      --trend float             The degree of trend to attack with
```