# krakend private auth server response

![how to auth provider plugin works](https://github.com/castmetal/krakend-private-auth-server-response/blob/main/autho-provider-plugin.png)

- Creating a default http-server plugin for any endpoints to request for a private auth server with your Auth Token Validation

This plugin validates any endpoint with a private flag on URL and send a default request to your private auth service and create a new header called x-user containing the payload information about your profile customer service.

Otherwise, this plugin manipulates the response with the backend header error and payload for any endpoints,

## Configuration

On your krakend.json add the name plugin and compile your .so version into plugins folder

```
{
    "$schema": "https://www.krakend.io/schema/v3.json",
    "version": 3,
    "name": "API Gateway",
    "extra_config": {
        "plugin/http-server": {
            "name": [
                "krakend-private-auth-server-response"
            ],
            "krakend-private-auth-server-response": {
                "auth_url": "{{ env "AUTH_SERVICE_URL" }}/auth/profile",
                "auth_method": "GET",
                "token_header": "{{ env "TOKEN_HEADER" }}",
                "params": {},
                "private_flag": "[{{ env "PRIVATE_FLAG" }}]",
                "error_flag": "{{ env "ERROR_FLAG" }}"
            }
        }
    }
}
```

- `auth_url`: Your Auth Provider Service Url, to request for private endpoints
- `auth_method`: Your Auth Provider Service Method, to request for private endpoints
- `token_header`: Your Token Header, for send to your Auth Provider Service
- `params`: Params if necessary, for send to your Auth Provider Service
- `private_flag`: Private Flag to check in your endpoints. This example you need to insert /[{private_flag}] to your krakend endpoints to insert an auth private service provider
- `params`: Params if necessary, for send to your Auth Provider Service
- `error_flag`: Error flag name of your: return_error_details, on your backends configuration

## Compile

Check your krakend compatibilities,

Run the example code, change with your local configuration.

- `krakend check-plugin --go 1.17.9 --libc GLIBC-2.31 --sum ./go.sum`
- 1.17.9 is your go version
- GLIBC-2.31 is your libc version

If you received this response: <pre>No incompatibilities found!</pre>:

Run:

- `go build -x -buildmode=plugin -o krakend-private-auth-server-response.so krakend-private-auth-server-response.go`

- Copy `krakend-private-auth-server-response.so` to your plugin folder configuration on krakend.

Example:

```
    "plugin": {
        "pattern": ".so",
        "folder": "./plugins/"
    },
```

## Endpoint config

- Add the private flag to set an endpoint as private, and intercept any request to your auth service provider
- Each endpoint with the private flag will validate with the header token passing ons request

Example:

```
{
    "endpoint": "/users/list/[{{ env "PRIVATE_FLAG" }}]",
    "method": "GET",
    "output_encoding": "json",
    "input_headers": [
        "Authorization",
        "Content-Type",
        "x-auth",
        "x-origin",
        "x-user",
        "x-request-id",
        "Accept-Encoding"
    ],
    "backend": [
        {
            "url_pattern": "/list",
            "encoding": "json",
            "sd": "static",
            "method": "GET",
            "host": [
                "{{ env "USER_SERVICE_URL" }}"
            ],
            "extra_config": {
                "backend/http": {
                    "return_error_details": "{{ env "ERROR_FLAG" }}"
                }
            }
        }
    ]
}
```

- Set the input headers to send:
- x-auth: example of your auth header for send to your auth provider service
- x-origin: tracking the origin of request (url endpoint of krakend will send as header)
- x-user: header pass with payload of your auth provider service
- x-request-id: the unique request-id as uuid v4

If your endpoint set with private flag:

- Before the endpoint execution, krakend-private-auth-server-response send a request to your auth provider service url, with token header set on krakend.json config and intercept request.
- If your auth provider service status code is 200, request will execute in your backend and return the response
- After the response the krakend-private-auth-server-response collect your error details if exists, and send the same status code and payload received to youer response
- If the execution of your back-end is ok, the same response will sent of your client

### Example of Private Headers

This payload is an example about your private endpoint received, after the auth provider service validate the execution (localhost execution example):

```
{
  host: 'localhost:4000',
  'user-agent': 'KrakenD Version 2.0.4',
  'transfer-encoding': 'chunked',
  'accept-encoding': 'gzip, deflate, br',
  'content-type': 'application/x-www-form-urlencoded',
  'x-auth': '{{your_auth_payload}}',
  'x-b3-sampled': '1',
  'x-b3-spanid': '15d67401ab73f5ba',
  'x-b3-traceid': 'fe7808b00bdabf5893ae659d7e42ced2',
  'x-forwarded-for': '::1',
  'x-forwarded-host': 'localhost:8001',
  'x-origin': '/auth/validate-token/[private]',
  'x-request-id': '88de71e7-c667-4292-8b5f-154b64810dab',
  'x-user': '{"_valid-uuid-req-server":"88de71e7-c667-4292-8b5f-154b64810dab",[[[your_payload_response from your auth provider service]]]}'
}
```

- For validation in your VPC endpoint service, validate if x-request-id is equal to your "x-user"."\_valid-uuid-req-server" and if exists x-user header.

## Run KrakenD with FC_ENABLE=1

Run your krakend:

- `CGO_ENABLED=1 FC_ENABLE=1 AUTH_SERVICE_URL="{your_auth_service_url}" TOKEN_HEADER="{your_token_header example: x-auth}" PRIVATE_FLAG="{your_private_flag example: private}" ERROR_FLAG="{your_error_flag example: my_error}" KRAKEND_PORT=8001 krakend run -d -c ./krakend.json -p 8001`

### Send a request to your private endpoint.

Example, if x-auth is your auth service provider header:

- `curl -H "x-auth: YourToken" http://localhost:8001/users/list/[private]`
