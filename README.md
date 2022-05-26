# Api Gateway

## Krakend

![how to Krakend works](https://raw.githubusercontent.com/castmetal/krakend-private-auth-server-response/main/KrakendFlow.png)

> KrakenD is an extensible, declarative, high-performance open-source API Gateway.

- Its core functionality is to create an API that acts as an aggregator of many microservices into single endpoints, doing the heavy-lifting automatically for you: aggregate, transform, filter, decode, throttle, auth, and more.

- KrakenD needs no programming as it offers a declarative way to create the endpoints. It is well structured and layered, and open to extending its functionality using plug-and-play middleware developed by the community or in-house.

## Install

### Docker build example

- RUN:

```sh
docker build --build-arg ENV=prod --build-arg AUTHORIZER_SERVICE_URL="{your authorizer service url example: http://localhost:8000}" --build-arg LOGIN_SERVICE_URL="{your login service url example: http://localhost:4000}" --build-arg PRIVATE_FLAG="{your endpoint private flag example: private}" --build-arg TOKEN_HEADER="x-auth" -t mykrakend .
```

- Run docker exec listening 8001 port tcp and exposing

### Without docker

> Install Krakend on link: [Krakend Install](https://www.krakend.io/download/)

For local test, run it on your terminal with krakend install, example:

```sh
ERROR_FLAG="myerror_flag" PRIVATE_FLAG="private" CGO_ENABLED=1 FC_ENABLE=1 TOKEN_HEADER="x-auth" AUTHORIZER_SERVICE_URL="{your_auth_service_url:port}" LOGIN_SERVICE_URL="{your_service_url:port}" KRAKEND_PORT=8001 krakend run -d -c ./krakend.json -p 8001
```

## How this API Gateway Works

This API Gateway is working with a private server auth provider.

This plugin validates any endpoint with a private flag on URL and send a default request to your private auth service and create a new header called x-user containing the payload information about your profile customer service.

![how to auth provider plugin works](https://raw.githubusercontent.com/castmetal/krakend-private-auth-server-response/main/autho-provider-plugin.png)

## Configure Endpoints

> To create an endpoint you only need to add an endpoint object under the endpoints list with the resource you want to expose. If no method is declared, it’s assumed to be read-only (GET).

The endpoints section looks like this:

```json
{
    "endpoints": [
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
        },
        {
            "endpoint": "/auth/sign",
            "method": "POST",
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
                    "url_pattern": "/validate/token",
                    "encoding": "json",
                    "sd": "static",
                    "method": "POST",
                    "host": [
                        "{{ env "AUTHORIZER_SERVICE_URL" }}"
                    ],
                    "extra_config": {
                        "backend/http": {
                            "return_error_details": "{{ env "ERROR_FLAG" }}"
                        }
                    }
                }
            ]
        }
    ]
}
```

- `{{ env "{your_env}" }}` - Read your env file value
- `PRIVATE_FLAG` - If your endpoint is private
- `backend` - your service configuration: host, url_pattern, etc
- `input_headers` - Headers for send to your backend service
- `return_error_details` - Necessary if your output_encoding is json
- `sd` - service discovery, change to dns if you have

> For more information access: [https://www.krakend.io/docs/endpoints/creating-endpoints/](https://www.krakend.io/docs/endpoints/creating-endpoints/)
