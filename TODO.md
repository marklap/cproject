# Things I'd Do Different/Better

These are some tasks I wish I had more time to take-on/complete.

## For the Prototype

### Proxy to Other Hosts

It seems fairly straightforward to accept requests for other hosts and proxy those requests. The application
can run on multiple hosts without much effort (statically linked binaries are the product of the default make target).

Roughly, the steps to add this functionality are:

1. Create a proxy client that streams results via a channel.
2. Add a `host` member to the  `TailRequest` struct.
3. Spawn a go thread to make a request to a target host.
4. Consume and emit the results from the proxy client channel.

The above changes are additive from the perspective of the customer and the `TailResponse` struct already includes a
`host` member which ensures the change will be backward-compatible.

### Order of Returned Lines

Given more time, I would adjust the application to return the lines in the order the are present in the log file.
This assumes the trade-off in performance is acceptable. The solution I've envisioned is to make a pass through
the log file in reverse order (as is currently implemented), but instead of yielding a line once it's identified
(and passes filters), instead the start and end byte offset of the line would be recorded. Then a second
step in the workflow would reading the file (again) and yield each line as described by it's start and end byte
offset.

### Test Coverage

There are some gaps in coverage for the core function of the application: `io.yieldLines()`. There are edge/corner(?)
cases that should be tested.


### Reduce Functional Complexity

At least two functions are prime candidates for some re-organization: `io.yieldLines()` and `handlers.TailHandler`.
These functions are (relatively massive) and deserve to be broken up into smaller functions to improve
readabilty and maintainability.


## For Production

### Instrumentation

The current app is completely devoid of any instrumentation except for a trivial amount of logged events. This is
a non-starter for production. I would want to see insrumentation for application (memory use, gc, threads, etc),
web service (requests:count/rate; response[code]:count/rate, bytes[read|write]:count/rate, latency, etc),
business metrics (log files read, success-v-error, unique users, etc).

### Containerization

There's benefit to dev'ing, testing and deploying an application in a containerized environment. This increases
confidence that an implementation/modification made in dev will work in production. There may also be some benefit
in portability if that sort of thing is important to folks.