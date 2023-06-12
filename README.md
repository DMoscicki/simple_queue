# simple_queue
Queue with REST interface.

Created without package initialization and only internal resources of the Golang language

Example GET request:
`curl -XPUT http://127.0.0.1/pet?v=cat`,
`curl -XPUT http://127.0.0.1/pet?v=dog`,
`curl -XPUT http://127.0.0.1/role?v=manager`,
`curl -XPUT http://127.0.0.1/role?v=executive`

When you make GET request you can use the timeout argument for waiting the answer:
curl http://127.0.0.1/pet?timeout=N
