This is a starting point for the CodeCrafters HTTP Server Challenge
["Build Your Own HTTP server"](https://app.codecrafters.io/courses/http-server/overview).

1. Ensure you have `go (1.22)` installed locally
1. Run `./run_program.sh` to run your program, which is implemented in
   `app/server.go`.

## Tests
```bash
curl -v http://localhost:4221
-- HTTP/1.1 200 OK\r\n\r\n --

curl -v http://localhost:4221/abcdefg
-- HTTP/1.1 404 Not Found\r\n\r\n --

curl -v http://localhost:4221
-- HTTP/1.1 200 OK\r\n\r\n --

curl -v http://localhost:4221/echo/abc
-- HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 3\r\n\r\nabc --

curl -v --header "User-Agent: foobar/1.2.3" http://localhost:4221/user-agent
-- HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 12\r\n\r\nfoobar/1.2.3 --

(sleep 3 && printf "GET / HTTP/1.1\r\n\r\n") | nc localhost 4221 &
(sleep 3 && printf "GET / HTTP/1.1\r\n\r\n") | nc localhost 4221 &
(sleep 3 && printf "GET / HTTP/1.1\r\n\r\n") | nc localhost 4221 &
HTTP/1.1 200 OK\r\n\r\n
HTTP/1.1 200 OK\r\n\r\n
HTTP/1.1 200 OK\r\n\r\n

echo -n 'Hello, World!' > /tmp/foo
curl -i http://localhost:4221/files/foo
HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: 14\r\n\r\nHello, World!

curl -i http://localhost:4221/files/non_existant_file
HTTP/1.1 404 Not Found\r\n\r\n

curl -v --data "12345" -H "Content-Type: application/octet-stream" http://localhost:4221/files/file_123
HTTP/1.1 201 Created\r\n\r\n

curl -v -H "Accept-Encoding: gzip" http://localhost:4221/echo/abc

-- HTTP/1.1 200 OK
Content-Type: text/plain
Content-Encoding: gzip

...  // Body omitted. --

curl -v -H "Accept-Encoding: invalid-encoding" http://localhost:4221/echo/abc
-- HTTP/1.1 200 OK
Content-Type: text/plain

...  // Body omitted. --

curl -v -H "Accept-Encoding: invalid-encoding-1, gzip, invalid-encoding-2" http://localhost:4221/echo/abc
-- HTTP/1.1 200 OK
Content-Type: text/plain
Content-Encoding: gzip

// Body omitted. --

curl -v -H "Accept-Encoding: invalid-encoding-1, invalid-encoding-2" http://localhost:4221/echo/abc
-- HTTP/1.1 200 OK
Content-Type: text/plain

// Body omitted. --

curl -v -H "Accept-Encoding: gzip" http://localhost:4221/echo/abc | hexdump -C
-- HTTP/1.1 200 OK
Content-Encoding: gzip
Content-Type: text/plain
Content-Length: 23

1F 8B 08 00 00 00 00 00  // Hexadecimal representation of the response body
00 03 4B 4C 4A 06 00 C2
41 24 35 03 00 00 00 --

echo -n <uncompressed-str> | gzip | hexdump -C.
```